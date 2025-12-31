package rpc

import (
	"fmt"
	"testing"
	"time"

	"github.com/arcology-network/streamer/broker"
	"github.com/google/uuid"
)

// echoActor simulates a server-side Actor, returning a Completion directly upon receiving an Invocation
type echoActor struct{}

func (a *echoActor) Consume(evt broker.ActorEvent) {
	inv, ok := evt.(*broker.RPCInvocation)
	if !ok {
		return
	}

	// create Completion
	comp := &broker.RPCCompletion{
		ID:      inv.ID,
		Payload: fmt.Sprintf("response for %s", inv.Method),
		Error:   "",
	}

	// go GlobalRPCClient.Controller.CompletionNotify(comp)
	go GlobalRPCClient.Streamer.Send(inv.ReplyTo, comp)
}

func TestRPCClientServer(t *testing.T) {
	// init Streamer
	stream := broker.NewStatefulStreamer()

	// init global RPCClient
	InitGlobalRPCClient(stream, 10, 2)

	// create RPC server
	serviceName := "echoService"
	actor := &echoActor{}
	NewRPCService(serviceName, actor, stream, 10)

	// Start the Controller of the global RPCClient (already started inside InitGlobalRPCClient)
	// Start the server-side Controller (already started inside NewRPCService)
	// Note: Both will Serve()

	go stream.Serve()
	reqID := uuid.NewString()
	// ----------------- async -----------------
	waiter := make(broker.CompletionWaiter, 1)
	GlobalRPCClient.Controller.Invoke(serviceName, "AsyncMethod", reqID, "hello async", waiter)

	select {
	case r := <-waiter:
		if r.Payload != "response for AsyncMethod" {
			t.Errorf("unexpected async payload: %v", r.Payload)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for RPCCompletion (async)")
	}

	// ----------------- sync -----------------
	resp, err := GlobalRPCClient.SyncCall(serviceName, "SyncMethod", reqID, "hello sync", 3*time.Second)
	if err != nil {
		t.Fatalf("SyncRPC error: %v", err)
	}
	if resp.Payload != "response for SyncMethod" {
		t.Errorf("unexpected sync payload: %v", resp.Payload)
	}

	t.Logf("RPC test passed, async reqID=%s, sync respID=%s", reqID, resp.ID)
}
