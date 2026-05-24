package storage

import "sync"

var (
	integrationSubscribersMu sync.Mutex
	integrationSubscribers   = make(map[int]chan struct{})
	nextIntegrationSubID     int
)

func SubscribeIntegrationChanges() (int, <-chan struct{}) {
	integrationSubscribersMu.Lock()
	defer integrationSubscribersMu.Unlock()

	nextIntegrationSubID++
	id := nextIntegrationSubID
	ch := make(chan struct{}, 1)
	integrationSubscribers[id] = ch
	return id, ch
}

func UnsubscribeIntegrationChanges(id int) {
	integrationSubscribersMu.Lock()
	defer integrationSubscribersMu.Unlock()

	ch, ok := integrationSubscribers[id]
	if !ok {
		return
	}
	delete(integrationSubscribers, id)
	close(ch)
}

func NotifyIntegrationChanged() {
	integrationSubscribersMu.Lock()
	defer integrationSubscribersMu.Unlock()

	for _, ch := range integrationSubscribers {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}
