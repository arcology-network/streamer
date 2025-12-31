package broker

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

type mockActor struct {
	t          *testing.T
	controller *RPCController
}

func (a *mockActor) Consume(evt ActorEvent) {
	switch e := evt.(type) {
	case *RPCInvocation:

		comp := &RPCCompletion{
			ID:      e.ID,
			Payload: fmt.Sprintf("response for %s", e.Method),
			Error:   "",
		}

		a.controller.pendingLock.Lock()
		waiter, exists := a.controller.completionWaiters[e.ID]
		a.controller.pendingLock.Unlock()
		if exists {
			waiter <- comp
		}
	case *RPCCompletion:

	default:
		a.t.Fatalf("unexpected event type: %T", e)
	}
}

type mockBuffer struct {
	listeners []streamListener
}

func (b *mockBuffer) RegisterListener(l streamListener) {
	b.listeners = append(b.listeners, l)
}

func (b *mockBuffer) Notify(item interface{}) {
	for _, l := range b.listeners {
		l.Notify(item)
	}
}

func (b *mockBuffer) Add(newItem interface{}) {
	b.Notify(newItem)
}
func (b *mockBuffer) Serve() {
}

func TestRPCController_InvokeAndSync(t *testing.T) {
	stream := &StatefulStreamer{
		buffers: make(map[string]streamBuffer),
	}

	actor := &mockActor{t: t}
	controller := NewRPCController(actor, stream, 10)
	actor.controller = controller

	// ===== register service invocation listener =====
	svcBuf := &mockBuffer{}
	stream.buffers["service1"] = svcBuf
	svcBuf.RegisterListener(controller.GetListener("service1"))

	// ===== register reply listener =====
	replyBuf := &mockBuffer{}
	stream.buffers[controller.respChanName] = replyBuf
	replyBuf.RegisterListener(&RpcCompletionListener{coordinator: controller})

	// 🔑 start controller
	go controller.Serve()

	reqID := uuid.NewString()
	// async call
	waiter := make(CompletionWaiter, 1)
	controller.Invoke("service1", "TestMethod", reqID, "hello", waiter)

	inv := &RPCInvocation{
		ID:      reqID,
		Method:  "TestMethod",
		ReplyTo: controller.respChanName,
	}
	actor.Consume(inv)

	select {
	case r := <-waiter:
		if r.Payload != "response for TestMethod" {
			t.Errorf("unexpected payload: %v", r.Payload)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for RPCCompletion (async)")
	}

	// sync call
	reqID = uuid.NewString()
	resp, err := controller.SyncRPC(
		"service1",
		"SyncMethod",
		reqID,
		"sync payload",
		time.Second,
	)
	if err != nil {
		t.Fatalf("SyncRPC error: %v", err)
	}

	if resp.Payload != "response for SyncMethod" {
		t.Fatalf("unexpected SyncRPC payload: %v", resp.Payload)
	}
}
