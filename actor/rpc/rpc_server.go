package rpc

import (
	"github.com/arcology-network/streamer/broker"
)

// Calculate the buffer size dynamically based on concurrency, with a minimum of minSize and a maximum of maxSize
func autoBufferSize(concurrent int) int {
	const minSize = 10
	const maxSize = 1000

	if concurrent < minSize {
		return minSize
	} else if concurrent > maxSize {
		return maxSize
	}
	return concurrent
}

// RPCService encapsulates a server-side actor + channel
type RPCService struct {
	Name       string                // Service name, corresponding to stream channel
	Actor      broker.RPCActor       // Processing logic
	Consumer   broker.StreamConsumer // StreamConsumer
	Controller *broker.RPCController // RPCController
	bufSize    int
}

// NewRPCService creates an RPC server
func NewRPCService(name string, actor broker.RPCActor, stream *broker.StatefulStreamer, maxConcurrent int) *RPCService {
	bufSize := autoBufferSize(maxConcurrent)

	// create controller
	ctrl := broker.NewRPCController(actor, stream, bufSize)

	//**********************************************************
	// create consumer（note：inputs points to the channel registered by itself）
	consumer := broker.NewDefaultConsumer(name, []string{name}, ctrl)
	stream.RegisterConsumer(consumer)

	// start controller Serve
	go ctrl.Serve()

	return &RPCService{
		Name:       name,
		Actor:      actor,
		Consumer:   consumer,
		Controller: ctrl,
		// buffer:     buf,
		bufSize: bufSize,
	}
}
