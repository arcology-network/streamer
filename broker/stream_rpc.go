/*
 *   Copyright (c) 2024 Arcology Network

 *   This program is free software: you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation, either version 3 of the License, or
 *   (at your option) any later version.

 *   This program is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *   GNU General Public License for more details.

 *   You should have received a copy of the GNU General Public License
 *   along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package broker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/arcology-network/streamer/logger"
)

// type RPCSource int

// const (
// 	RPCSourceInternal RPCSource = iota
// 	RPCSourceExternal
// )

// RPC types
type RPCInvocation struct {
	ID         string
	ServerName string
	Method     string
	Payload    interface{}
	ReplyTo    string
	// Source     RPCSource //
}

type RPCCompletion struct {
	ID      string
	Payload interface{}
	Error   string
}

type CompletionWaiter chan *RPCCompletion

// ---------------- Completion Listener ----------------
// reply listener，Used to distribute responses entering the global reply buffer to the waiting parties
type RpcCompletionListener struct {
	coordinator *RPCController
}

func (l *RpcCompletionListener) Notify(item interface{}) {
	comp, ok := item.(*RPCCompletion)
	if !ok {
		return
	}

	l.coordinator.DeliverCompletion(comp)
}

type RPCActor interface {
	Consume(evt ActorEvent)
}

// ---------------- RPCController ----------------
// RPCController implementation, also satisfying the StreamController interface
type RPCController struct {
	reqChan      chan *RPCInvocation
	actor        RPCActor
	stream       *StatefulStreamer
	respChanName string

	completionWaiters map[string]chan *RPCCompletion
	pendingLock       sync.Mutex

	// local lock to prevent concurrent buffer creation races for topics
	createBufLock sync.Mutex
}

// NewRPCController：Create the controller and initialize the global reply buffer
func NewRPCController(actor RPCActor, stream *StatefulStreamer, bufSize int) *RPCController {
	rc := &RPCController{
		reqChan:           make(chan *RPCInvocation, bufSize),
		actor:             actor,
		stream:            stream,
		respChanName:      "__rpc_reply_global",
		completionWaiters: make(map[string]chan *RPCCompletion),
	}

	// Create/register the global reply buffer (if it was previously created, overwrite it with the same object)
	// Use ensureStreamBuffer to ensure compatibility (avoid directly writing to ss.buffers to prevent race conditions)
	rc.ensureStreamBuffer(rc.respChanName, bufSize)

	// register the reply listener on the global reply buffer
	if buf, ok := rc.stream.buffers[rc.respChanName]; ok {
		buf.RegisterListener(&RpcCompletionListener{coordinator: rc})
	}

	return rc
}

// ensureStreamBuffer ensures that stream.buffers[name] exists and has started Serve()
// The size parameter is used to set the buffer size; if it already exists, it is ignored.
func (rc *RPCController) ensureStreamBuffer(name string, size int) {
	// fast path
	if buf := rc.stream.buffers[name]; buf != nil {
		return
	}

	// protect creation to reduce double-creation races
	rc.createBufLock.Lock()
	defer rc.createBufLock.Unlock()

	// re-check after obtaining lock
	if buf := rc.stream.buffers[name]; buf != nil {
		return
	}

	// create buffer and start Serve
	buf := NewDefaultStreamBuffer(name, size)
	rc.stream.buffers[name] = buf
	go buf.Serve()
}

// ------- Implement the StreamController interface -------
// GetListener is used to bind the named input to this controller (for registering a Consumer)
func (rc *RPCController) GetListener(name string) streamListener {
	return &defaultListener{
		name:       name,
		controller: rc,
	}
}

// Notify receives data from the defaultListener (buffer -> listener -> controller)
func (rc *RPCController) Notify(name string, data interface{}) {
	// Supports directly passing in *RPCRequest
	req, ok := data.(*RPCInvocation)
	if !ok {
		logger.Log.Warn(context.Background(), "Received data is not RPCInvocation")
		return
	}
	// If the caller does not specify ReplyTo, use the global reply channel
	if req.ReplyTo == "" {
		req.ReplyTo = rc.respChanName
	}

	// ensure the reply buffer exists (safety; probably already ensured in NewRPCController)
	rc.ensureStreamBuffer(req.ReplyTo, 1)
	rc.reqChan <- req
}

// Serve processes requests from the request queue, passes them to the actor, and sends the response back to the reply buffer
func (rc *RPCController) Serve() {
	for inv := range rc.reqChan {
		log.Printf("Serve got invocation %s %s", inv.ID, inv.Method)

		rc.actor.Consume(inv)
	}
}

func (rc *RPCController) GenerateDotLabel() string { return "RPCController" }

// GetActor returns an object that implements Actor (returns nil if the actor does not implement Actor)
func (rc *RPCController) GetActor() Actor {
	if a, ok := rc.actor.(Actor); ok {
		return a
	}
	return nil
}

func (rc *RPCController) DeliverCompletion(comp *RPCCompletion) {
	rc.pendingLock.Lock()
	waiter, ok := rc.completionWaiters[comp.ID]
	if ok {
		delete(rc.completionWaiters, comp.ID)
	}
	rc.pendingLock.Unlock()

	if ok {
		select {
		case waiter <- comp:
		default:
		}
	}
}

func (rc *RPCController) CancelWaiter(id string, err error) {
	rc.pendingLock.Lock()
	waiter, ok := rc.completionWaiters[id]
	if ok {
		delete(rc.completionWaiters, id)
	}
	rc.pendingLock.Unlock()

	if ok {
		select {
		case waiter <- &RPCCompletion{
			ID:    id,
			Error: err.Error(),
		}:
		default:
		}
	}
}

// ***********************************************************
func (rc *RPCController) Invoke(
	service string,
	method string,
	reqID string,
	payload interface{},
	waiter CompletionWaiter,
) {
	rc.pendingLock.Lock()
	rc.completionWaiters[reqID] = waiter
	rc.pendingLock.Unlock()

	rc.ensureStreamBuffer(rc.respChanName, 1)

	inv := &RPCInvocation{
		ID:         reqID,
		ServerName: service,
		Method:     method,
		Payload:    payload,
		ReplyTo:    rc.respChanName,
		// Source:     RPCSourceExternal,
	}

	rc.ensureStreamBuffer(service, 64)

	rc.stream.Send(service, inv)
}

// ------- Synchronous RPC -------

// SyncRPC initiates a synchronous RPC call (by default, it uses the global reply buffer name of the controller)
// method: method name; payload: parameters; timeout: wait time
func (rc *RPCController) SyncRPC(
	service, method, reqID string,
	payload interface{},
	timeout time.Duration,
) (*RPCCompletion, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	waiter := make(chan *RPCCompletion, 1)
	rc.Invoke(service, method, reqID, payload, waiter)

	select {
	case comp := <-waiter:
		if comp.Error != "" {
			return comp, fmt.Errorf(comp.Error)
		}
		return comp, nil
	case <-ctx.Done():
		rc.CancelWaiter(reqID, ctx.Err())
		return nil, ctx.Err()
	}
}
