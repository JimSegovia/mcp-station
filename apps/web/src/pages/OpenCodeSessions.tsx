import { useEffect, useState, useRef } from "react";
import { Trash2, RefreshCw, Bot, Terminal, Clock, LoaderCircle, Send, MessageSquare, Plus, Brain, Wrench } from "lucide-react";
import { useOpenCodeSessionStore } from "@/store/opencodeSessionStore";
import OpenCodeTerminal from "@/components/OpenCodeTerminal";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";

interface MessagePart {
  type: string;
  text?: string;
  reasoning?: string;
  tool_name?: string;
}

interface MessageItem {
  info: { id: string; role: string };
  parts: MessagePart[];
}

const toolIcons: Record<string, React.ReactNode> = {
  opencode_ask: <Bot className="h-3.5 w-3.5" />,
  opencode_run: <Terminal className="h-3.5 w-3.5" />,
  ui_prompt: <Send className="h-3.5 w-3.5" />,
};

export default function OpenCodeSessions() {
  const { sessions, loading, load, deleteSession, cleanExpired } =
    useOpenCodeSessionStore();

  const [prompt, setPrompt] = useState("");
  const [agent, setAgent] = useState("build");
  const [activeSession, setActiveSession] = useState("");
  const [sending, setSending] = useState(false);
  const [errorMsg, setErrorMsg] = useState("");
  const [messages, setMessages] = useState<MessageItem[]>([]);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const messagesContainerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    load();
  }, [load]);

  useEffect(() => {
    if (activeSession) {
      fetchMessages(activeSession);
    } else {
      setMessages([]);
    }
  }, [activeSession]);

  useEffect(() => {
    const el = messagesContainerRef.current;
    if (!el) return;
    const isNearBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 100;
    if (isNearBottom) {
      messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
    }
  }, [messages]);

  const fetchMessages = async (sid: string) => {
    try {
      const res = await fetch(`/api/opencode/sessions/${encodeURIComponent(sid)}/messages`);
      if (!res.ok) return;
      const data: MessageItem[] = await res.json();
      setMessages(data);
    } catch {
      // ignore
    }
  };

  const handleSend = async () => {
    if (!prompt.trim() || sending) return;
    setSending(true);
    setErrorMsg("");
    try {
      const body: Record<string, string> = { prompt: prompt.trim(), agent };
      if (activeSession) body.sessionId = activeSession;

      const res = await fetch("/api/opencode/sessions", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });
      if (!res.ok) {
        const err = await res.json().catch(() => ({ error: `HTTP ${res.status}` }));
        setErrorMsg(err.error || "Error desconocido");
      } else {
        const data = await res.json();
        setPrompt("");
        setErrorMsg("");
        if (!activeSession && data.sessionId) {
          setActiveSession(data.sessionId);
        }
        if (data.sessionId) {
          load();
          setTimeout(() => fetchMessages(data.sessionId), 500);
        }
      }
    } catch (e: any) {
      setErrorMsg(e.message || String(e));
    } finally {
      setSending(false);
    }
  };

  const activeSessionData = sessions.find((s) => s.sessionId === activeSession);

  return (
    <div className="max-w-2xl space-y-6">
      <OpenCodeTerminal />

      {/* Prompt input */}
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-base">Enviar prompt a OpenCode</CardTitle>
          <CardDescription>
            {activeSession
              ? "Continuando sesion existente"
              : "Se creara una nueva sesion"}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="flex gap-2 flex-wrap">
            <select
              value={agent}
              onChange={(e) => setAgent(e.target.value)}
              className="h-9 w-24 rounded-md border border-border bg-background text-foreground px-2 text-xs font-mono"
            >
              <option value="build">build</option>
              <option value="plan">plan</option>
              <option value="explore">explore</option>
            </select>
            <select
              value={activeSession}
              onChange={(e) => setActiveSession(e.target.value)}
              className="h-9 flex-1 min-w-40 rounded-md border border-border bg-background text-foreground px-2 text-xs font-mono"
            >
              <option value="">+ Nueva sesion</option>
              {sessions.map((s) => (
                <option key={s.sessionId} value={s.sessionId}>
                  {s.toolName}/{s.sessionId.slice(0, 12)}... — {s.prompt.slice(0, 40)}
                </option>
              ))}
            </select>
          </div>
          <div className="flex gap-2">
            <Input
              value={prompt}
              onChange={(e) => setPrompt(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && handleSend()}
              placeholder={
                activeSession
                  ? "Continuar conversacion..."
                  : "Ej: Crea un archivo README.md con la estructura del proyecto"
              }
              className="flex-1 font-mono text-xs"
            />
            <Button onClick={handleSend} disabled={!prompt.trim() || sending}>
              {sending ? (
                <LoaderCircle className="h-3.5 w-3.5 animate-spin" />
              ) : (
                <Send className="h-3.5 w-3.5 mr-1" />
              )}
              Enviar
            </Button>
          </div>
          {errorMsg && (
            <div className="rounded-md bg-destructive/10 border border-destructive/30 p-2">
              <p className="text-xs text-destructive">{errorMsg}</p>
            </div>
          )}
          {sending && !activeSession && (
            <div className="flex items-center gap-2 text-xs text-muted-foreground">
              <LoaderCircle className="h-3 w-3 animate-spin" />
              Esperando respuesta de OpenCode...
            </div>
          )}
        </CardContent>
      </Card>

      {/* Conversation view */}
      {activeSession && (
        <Card>
          <CardHeader className="pb-2">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <MessageSquare className="h-4 w-4 text-muted-foreground" />
                <CardTitle className="text-sm font-mono">
                  {activeSession.slice(0, 16)}...
                </CardTitle>
                {activeSessionData && (
                  <Badge variant="outline" className="text-[10px]">
                    {activeSessionData.toolName}
                  </Badge>
                )}
              </div>
              <div className="flex gap-1">
                <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => fetchMessages(activeSession)}>
                  <RefreshCw className="h-3 w-3" />
                </Button>
                <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => { setActiveSession(""); setMessages([]); }}>
                  <Plus className="h-3.5 w-3.5" />
                </Button>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <div ref={messagesContainerRef} className="bg-black/90 rounded-md border border-border p-3 max-h-96 overflow-y-auto font-mono text-xs leading-relaxed space-y-2">
              {messages.length === 0 ? (
                <span className="text-muted-foreground">Sin mensajes — envia un prompt para empezar</span>
              ) : (
                messages.map((msg, i) => (
                  <div key={i} className="space-y-1">
                    <div className="text-[10px] text-muted-foreground mb-0.5">
                      {msg.info.role === "user" ? "You" : "Assistant"}
                    </div>
                    {msg.parts.map((p, j) => {
                      if (p.type === "reasoning" || p.reasoning) {
                        return (
                          <details key={j} className="text-xs">
                            <summary className="text-gray-500 cursor-pointer hover:text-gray-400 flex items-center gap-1">
                              <Brain className="h-3 w-3 inline" />
                              Thinking
                            </summary>
                            <p className="text-gray-400 whitespace-pre-wrap break-all leading-relaxed mt-1 pl-3 border-l border-gray-700">
                              {p.reasoning || p.text || ""}
                            </p>
                          </details>
                        );
                      }
                      if (p.type === "tool_call" || p.tool_name) {
                        return (
                          <div key={j} className="text-xs text-blue-400 flex items-center gap-1">
                            <Wrench className="h-3 w-3" />
                            {p.tool_name || "tool"}
                          </div>
                        );
                      }
                      return (
                        <p key={j} className="text-xs text-green-300/90 whitespace-pre-wrap break-all leading-relaxed">
                          {p.text || ""}
                        </p>
                      );
                    })}
                  </div>
                ))
              )}
              {sending && (
                <div className="text-yellow-400 animate-pulse">[assistant] Pensando...</div>
              )}
              <div ref={messagesEndRef} />
            </div>
          </CardContent>
        </Card>
      )}

      <Separator />

      {/* Session list */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Historial</h1>
          <p className="text-sm text-muted-foreground mt-1">
            {sessions.length} sesiones · auto-limpieza cada 24h
          </p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={load} disabled={loading}>
            <RefreshCw
              className={`h-3.5 w-3.5 mr-1 ${loading ? "animate-spin" : ""}`}
            />
            Actualizar
          </Button>
          <Button variant="ghost" size="sm" onClick={cleanExpired}>
            <Trash2 className="h-3.5 w-3.5 mr-1" />
            Limpiar
          </Button>
        </div>
      </div>

      {loading && sessions.length === 0 && (
        <div className="flex items-center justify-center py-12">
          <LoaderCircle className="h-6 w-6 animate-spin text-muted-foreground" />
        </div>
      )}

      {!loading && sessions.length === 0 && (
        <Card>
          <CardContent className="py-12 text-center">
            <Bot className="h-8 w-8 text-muted-foreground mx-auto mb-3" />
            <p className="text-sm text-muted-foreground">
              No hay sesiones registradas
            </p>
            <p className="text-xs text-muted-foreground mt-1">
              Envia un prompt arriba o espera que XiaoZhi use opencode_ask
            </p>
          </CardContent>
        </Card>
      )}

      <div className="space-y-3">
        {sessions.map((s) => (
          <Card
            key={s.sessionId}
            className={`cursor-pointer transition-colors ${
              activeSession === s.sessionId ? "border-primary/50 bg-primary/5" : ""
            }`}
            onClick={() => setActiveSession(s.sessionId)}
          >
            <CardHeader className="pb-3">
              <div className="flex items-start justify-between gap-3">
                <div className="flex items-center gap-2 min-w-0">
                  <div className="text-muted-foreground">
                    {toolIcons[s.toolName] ?? <Bot className="h-3.5 w-3.5" />}
                  </div>
                  <div>
                    <CardTitle className="text-sm font-mono">
                      {s.sessionId.slice(0, 16)}...
                    </CardTitle>
                    <CardDescription className="mt-0.5">
                      {s.prompt.length > 60
                        ? s.prompt.slice(0, 60) + "..."
                        : s.prompt}
                    </CardDescription>
                  </div>
                </div>
                <div className="flex items-center gap-2 shrink-0">
                  <Badge variant="outline" className="text-[10px]">
                    {s.toolName}
                  </Badge>
                  <Badge
                    variant={s.status === "completed" ? "success" : "secondary"}
                    className="text-[10px]"
                  >
                    {s.status || "pending"}
                  </Badge>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7"
                    onClick={(e) => {
                      e.stopPropagation();
                      deleteSession(s.sessionId, s.trackId);
                    }}
                  >
                    <Trash2 className="h-3.5 w-3.5 text-muted-foreground" />
                  </Button>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <div className="flex items-center gap-3 text-xs text-muted-foreground">
                <span className="flex items-center gap-1">
                  <Clock className="h-3 w-3" />
                  {s.age}
                </span>
                {s.title && <span className="truncate">· {s.title}</span>}
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
