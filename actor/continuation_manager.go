package actor

import (
	"fmt"
	"sync"
	"time"

	"github.com/arcology-network/streamer/broker"
)

type Continuation interface {
	Resume(comp *broker.RPCCompletion)
	Cancel(err error)
}

type ContinuationManager struct {
	mu    sync.Mutex
	items map[string]*continuationEntry
}

type continuationEntry struct {
	id        string
	cont      Continuation
	timer     *time.Timer
	createdAt time.Time
}

func NewContinuationManager() *ContinuationManager {
	return &ContinuationManager{
		items: make(map[string]*continuationEntry),
	}
}

func (m *ContinuationManager) Register(
	id string,
	cont Continuation,
	timeout time.Duration,
) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.items[id]; exists {
		panic("duplicate continuation id: " + id)
	}

	entry := &continuationEntry{
		id:        id,
		cont:      cont,
		createdAt: time.Now(),
	}

	if timeout > 0 {
		entry.timer = time.AfterFunc(timeout, func() {
			m.timeout(id)
		})
	}

	m.items[id] = entry
}

func (m *ContinuationManager) Resume(
	id string,
	comp *broker.RPCCompletion,
) bool {
	entry := m.pop(id)
	if entry == nil {
		return false
	}

	entry.cont.Resume(comp)
	return true
}

func (m *ContinuationManager) timeout(id string) {
	entry := m.pop(id)
	if entry == nil {
		return
	}

	entry.cont.Cancel(
		fmt.Errorf("rpc continuation timeout: %s", id),
	)
}

func (m *ContinuationManager) pop(id string) *continuationEntry {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry := m.items[id]
	if entry == nil {
		return nil
	}

	delete(m.items, id)

	if entry.timer != nil {
		entry.timer.Stop()
	}

	return entry
}

func (m *ContinuationManager) CancelAll(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, entry := range m.items {
		if entry.timer != nil {
			entry.timer.Stop()
		}
		entry.cont.Cancel(err)
		delete(m.items, id)
	}
}
