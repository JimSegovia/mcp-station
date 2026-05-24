package opencode

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

const maxLogLines = 500

type Manager struct {
	client  *Client
	cmd     *exec.Cmd
	port    int
	binPath string
	mu      sync.Mutex
	logs    []string
	logsMu  sync.RWMutex
	ready   bool
	cancel  context.CancelFunc
}

func NewManager(port int, username, password, binPath string) *Manager {
	return &Manager{
		client:  NewClient(fmt.Sprintf("http://127.0.0.1:%d", port), username, password),
		port:    port,
		binPath: binPath,
		logs:    make([]string, 0, maxLogLines),
	}
}

func (m *Manager) Client() *Client {
	return m.client
}

func (m *Manager) Logs() []string {
	m.logsMu.RLock()
	defer m.logsMu.RUnlock()
	out := make([]string, len(m.logs))
	copy(out, m.logs)
	return out
}

func (m *Manager) appendLog(line string) {
	m.logsMu.Lock()
	defer m.logsMu.Unlock()
	m.logs = append(m.logs, line)
	if len(m.logs) > maxLogLines {
		m.logs = m.logs[len(m.logs)-maxLogLines:]
	}
}

func resolveBinary(binPath string) (string, []string) {
	if binPath != "" {
		return binPath, nil
	}

	searchPaths := []string{
		"opencode",
		os.ExpandEnv("$HOME/.opencode/bin/opencode"),
		"/usr/local/bin/opencode",
		os.ExpandEnv("$HOME/go/bin/opencode"),
	}

	if runtime.GOOS == "windows" {
		return "wsl", []string{"bash", "-lc"}
	}

	for _, p := range searchPaths {
		if _, err := exec.LookPath(p); err == nil {
			return p, nil
		}
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	return "opencode", nil
}

func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.ready {
		return nil
	}

	if m.isPortOpen() {
		m.appendLog("Port already in use, attaching to existing opencode serve")
		log.Printf("opencode: port %d already in use, attaching to existing server", m.port)
		m.ready = true
		return nil
	}

	ctx, m.cancel = context.WithCancel(ctx)

	bin, prefixArgs := resolveBinary(m.binPath)

	srvArgs := []string{"serve",
		"--port", fmt.Sprintf("%d", m.port),
		"--hostname", "127.0.0.1",
	}

	var fullArgs []string
	if bin == "wsl" {
		binPath := "$HOME/.opencode/bin/opencode"
		if m.binPath != "" {
			binPath = m.binPath
		}
		cmdStr := fmt.Sprintf(
			`export PATH="$HOME/.opencode/bin:$HOME/go/bin:$PATH" && %s %s 2>&1`,
			binPath, strings.Join(srvArgs, " "))
		fullArgs = []string{"bash", "-l", "-c", cmdStr}
	} else {
		fullArgs = append(prefixArgs, srvArgs...)
	}

	m.appendLog(fmt.Sprintf("Starting: %s %s", bin, strings.Join(fullArgs, " ")))
	log.Printf("opencode: starting: %s %s", bin, strings.Join(fullArgs, " "))

	m.cmd = exec.CommandContext(ctx, bin, fullArgs...)
	setProcessGroup(m.cmd)

	if m.client.password != "" {
		m.cmd.Env = append(os.Environ(),
			"OPENCODE_SERVER_PASSWORD="+m.client.password,
		)
		if m.client.username != "" {
			m.cmd.Env = append(m.cmd.Env,
				"OPENCODE_SERVER_USERNAME="+m.client.username,
			)
		}
	} else {
		m.cmd.Env = os.Environ()
	}

	stdout, _ := m.cmd.StdoutPipe()
	stderr, _ := m.cmd.StderrPipe()

	if err := m.cmd.Start(); err != nil {
		m.appendLog(fmt.Sprintf("ERROR: %v", err))
		return fmt.Errorf("start opencode serve: %w", err)
	}

	if stdout != nil {
		go m.captureOutput(stdout)
	}
	if stderr != nil {
		go m.captureOutput(stderr)
	}

	m.appendLog(fmt.Sprintf("Process started (pid=%d)", m.cmd.Process.Pid))
	log.Printf("opencode: started serve on port %d (pid=%d)", m.port, m.cmd.Process.Pid)

	go func() {
		if err := m.cmd.Wait(); err != nil {
			time.Sleep(300 * time.Millisecond)
			m.appendLog(fmt.Sprintf("Process exited: %v", err))
			log.Printf("opencode: serve exited: %v", err)
			for _, l := range m.Logs() {
				log.Printf("opencode [output]: %s", l)
			}
		}
		m.mu.Lock()
		m.ready = false
		m.cmd = nil
		m.mu.Unlock()
	}()

	m.ready = true
	return nil
}

func (m *Manager) captureOutput(r io.Reader) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		m.appendLog(strings.TrimSpace(scanner.Text()))
	}
}

func (m *Manager) SendCommand(cmdText string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cmd == nil || m.cmd.Process == nil {
		return "", fmt.Errorf("opencode serve is not running")
	}

	stdin, err := m.cmd.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("cannot get stdin: %w", err)
	}

	_, err = io.WriteString(stdin, cmdText+"\n")
	if err != nil {
		return "", fmt.Errorf("write to stdin: %w", err)
	}

	m.appendLog("> " + cmdText)
	return "Command sent", nil
}

func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}

	if m.cmd != nil && m.cmd.Process != nil {
		m.cmd.Process.Signal(nil)
		time.Sleep(500 * time.Millisecond)
		m.cmd.Process.Kill()
		m.cmd = nil
	}
	m.ready = false
	m.appendLog("Server stopped")
	log.Println("opencode: serve stopped")
}

func (m *Manager) WaitReady(timeout time.Duration) error {
	deadline := time.After(timeout)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			m.appendLog("Timeout waiting for opencode serve")
			return fmt.Errorf("timeout waiting for opencode serve")
		case <-ticker.C:
			if _, err := m.client.Health(); err == nil {
				m.appendLog("OpenCode serve is ready")
				log.Println("opencode: serve is ready")
				return nil
			}
		}
	}
}

func (m *Manager) isPortOpen() bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", m.port), 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func (m *Manager) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.ready
}

func (m *Manager) RunWSLCommand(cmdStr string) {
	var c *exec.Cmd
	if runtime.GOOS == "windows" {
		c = exec.Command("wsl", "bash", "-c", cmdStr)
	} else {
		c = exec.Command("bash", "-c", cmdStr)
	}
	output, err := c.CombinedOutput()
	if err != nil {
		log.Printf("opencode: config command failed: %v output: %s", err, string(output))
	} else {
		log.Printf("opencode: config command ok: %s", strings.TrimSpace(string(output)))
	}
}
