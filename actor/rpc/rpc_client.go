package rpc

import (
	"fmt"
	"sync"
	"time"

	"github.com/arcology-network/streamer/broker"
)

// ------------------------
// use examples
// ------------------------

// 1. Register the global client during system initialization
// func init() {
//     ss := broker.NewStatefulStreamer()
//     actor.InitGlobalRPCClient(ss, 100)
// }

// 2. Can be used anywhere in the system
// resp, err := actor.GlobalRPCClient.SyncRPC("ServiceA", "Echo", "Hello", 2*time.Second)
//
//	if err != nil {
//	    log.Fatalf("RPC error: %v", err)
//	}
//
// fmt.Println("RPC resp:", resp.Payload)

// RPCClient encapsulates the global client, supporting requests to any service
type RPCClient struct {
	Streamer   *broker.StatefulStreamer
	Controller *broker.RPCController
	once       sync.Once
}

var RpcTimeout time.Duration

var GlobalRPCClient *RPCClient

// InitGlobalRPCClient Initialize global client
func InitGlobalRPCClient(stream *broker.StatefulStreamer, maxConcurrent int, requestTimeoutSeconds int) {
	bufSize := autoBufferSize(maxConcurrent)

	clientActor := &mockClientActor{}
	ctrl := broker.NewRPCController(clientActor, stream, bufSize)
	clientActor.controller = ctrl

	GlobalRPCClient = &RPCClient{
		Streamer:   stream,
		Controller: ctrl,
		once:       sync.Once{},
	}

	RpcTimeout = time.Duration(requestTimeoutSeconds) * time.Second

	go ctrl.Serve()
}

// SyncCall synchronous RPC
func (c *RPCClient) SyncCall(serviceName, method string, msgId string, payload interface{}, timeout time.Duration) (*broker.RPCCompletion, error) {
	if c == nil || c.Controller == nil {
		return nil, fmt.Errorf("RPC client not initialized")
	}
	return c.Controller.SyncRPC(serviceName, method, msgId, payload, timeout)
}

// mockClientActor is used to satisfy the RPCController interface (for testing purposes)
type mockClientActor struct {
	controller *broker.RPCController
}

func (m *mockClientActor) Consume(evt broker.ActorEvent) {
	comp, ok := evt.(*broker.RPCCompletion)
	if !ok {
		// The client does not handle invocation / msg
		return
	}

	m.controller.DeliverCompletion(comp)
}
