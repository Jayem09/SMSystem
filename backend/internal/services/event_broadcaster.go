package services

import (
	"sync"
	"sync/atomic"
	"time"
)

const (
	EventOrderCreated    = "order_created"
	EventOrderCompleted  = "order_completed"
	EventOrderDeleted    = "order_deleted"
	EventStockAdjusted   = "stock_adjusted"
	EventExpenseAdded    = "expense_added"
	EventTransferUpdated = "transfer_updated"
)

const broadcasterChannelBuffer = 16
const broadcasterControlBuffer = 32

type DashboardEvent struct {
	Type      string      `json:"type"`
	BranchID  uint        `json:"branch_id,omitempty"`
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
}

type SSEClient struct {
	ID        uint64
	Channel   chan DashboardEvent
	BranchID  uint
	CreatedAt time.Time
}

type EventBroadcaster struct {
	clients      map[uint64]*SSEClient
	register     chan *SSEClient
	unregister   chan *SSEClient
	broadcast    chan DashboardEvent
	mutex        sync.RWMutex
	nextClientID uint64
}

var (
	broadcasterOnce sync.Once
	broadcaster     *EventBroadcaster
)

func newEventBroadcaster() *EventBroadcaster {
	b := &EventBroadcaster{
		clients:      make(map[uint64]*SSEClient),
		register:     make(chan *SSEClient, broadcasterControlBuffer),
		unregister:   make(chan *SSEClient, broadcasterControlBuffer),
		broadcast:    make(chan DashboardEvent, 100),
		nextClientID: 0,
	}

	go b.run()

	return b
}

func InitBroadcaster() *EventBroadcaster {
	broadcasterOnce.Do(func() {
		broadcaster = newEventBroadcaster()
	})

	return broadcaster
}

func GetBroadcaster() *EventBroadcaster {
	return InitBroadcaster()
}

func (b *EventBroadcaster) run() {
	for {
		select {
		case client := <-b.register:
			if client == nil {
				continue
			}

			b.mutex.Lock()
			b.clients[client.ID] = client
			b.mutex.Unlock()
		case client := <-b.unregister:
			if client == nil {
				continue
			}

			b.mutex.Lock()
			existingClient, exists := b.clients[client.ID]
			if exists {
				delete(b.clients, client.ID)
				close(existingClient.Channel)
			}
			b.mutex.Unlock()
		case event := <-b.broadcast:
			if event.Timestamp.IsZero() {
				event.Timestamp = time.Now().UTC()
			}

			b.dispatch(event)
		}
	}
}

func (b *EventBroadcaster) Register(branchID uint) *SSEClient {
	client := &SSEClient{
		ID:        atomic.AddUint64(&b.nextClientID, 1),
		Channel:   make(chan DashboardEvent, broadcasterChannelBuffer),
		BranchID:  branchID,
		CreatedAt: time.Now().UTC(),
	}

	b.register <- client

	return client
}

func (b *EventBroadcaster) Unregister(client *SSEClient) {
	if client == nil {
		return
	}

	b.unregister <- client
}

func (b *EventBroadcaster) Broadcast(event DashboardEvent) {
	select {
	case b.broadcast <- event:
	default:
	}
}

func (b *EventBroadcaster) BroadcastToBranch(eventType string, branchID uint, payload interface{}) {
	b.Broadcast(DashboardEvent{
		Type:      eventType,
		BranchID:  branchID,
		Payload:   payload,
		Timestamp: time.Now().UTC(),
	})
}

func (b *EventBroadcaster) ClientCount() int {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	return len(b.clients)
}

func (b *EventBroadcaster) dispatch(event DashboardEvent) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	for _, client := range b.clients {
		if !shouldDeliverEvent(client, event) {
			continue
		}

		select {
		case client.Channel <- event:
		default:
		}
	}
}

func shouldDeliverEvent(client *SSEClient, event DashboardEvent) bool {
	if client == nil {
		return false
	}

	if event.BranchID == 0 || client.BranchID == 0 {
		return true
	}

	return client.BranchID == event.BranchID
}
