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

package actor

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/arcology-network/streamer/actor/rpc"
	"github.com/arcology-network/streamer/broker"
	brokerpk "github.com/arcology-network/streamer/broker"
	scommon "github.com/arcology-network/streamer/common"
	"github.com/arcology-network/streamer/logger"
)

type Actor struct {
	name   string
	broker *brokerpk.StatefulStreamer

	chain            *BusinessChain
	execCtx          *ExecutionContext
	rpcIndex         map[string]*ControlledNode
	rpcCompleteTopic string

	inbox chan brokerpk.ActorEvent

	continuations *ContinuationManager
}

func CreateActor(name string, broker *brokerpk.StatefulStreamer, businesses []Business, businessNames []string, filters []*Filter, concurrency int) *Actor {
	rpccomplete := name + ".rpc.complete"
	actor := &Actor{
		inbox:            make(chan brokerpk.ActorEvent, 1024),
		name:             name,
		broker:           broker,
		execCtx:          NewExecutionContext(name, rpccomplete, broker, concurrency),
		rpcIndex:         make(map[string]*ControlledNode, len(businesses)),
		rpcCompleteTopic: rpccomplete,

		continuations: NewContinuationManager(),
	}

	//create chain
	list_inputs := make([][]string, 0, len(businesses))
	list_outputs := make([]map[string]int, len(businesses))
	nodes := make([]*ControlledNode, 0, len(businesses))
	chainIsConjunction := false
	withRpcServer := false
	for i := range businesses {
		//create businesschain
		inputs, isConjunction := businesses[i].Inputs()
		list_inputs = append(list_inputs, inputs)
		list_outputs = append(list_outputs, businesses[i].Outputs())

		if isConjunction && len(businesses) > 1 {
			panic("All elements in the pipeline must be DisConjunction.")
		}

		if i == 0 {
			chainIsConjunction = isConjunction
		} else {
			chainIsConjunction = false
		}

		// ControlledNodes
		controllers := make([]Controller, 0, 2)
		if !isConjunction {
			if _, ok := businesses[i].(HeightSensitive); ok {
				controllers = append(controllers, NewHeightController(businesses[i].(HeightSensitive)))
			}
			if _, ok := businesses[i].(FSMCompatible); ok {
				controllers = append(controllers, NewFSMController(businesses[i].(FSMCompatible)))
			}
		}
		node := NewControlledNode(businesses[i], controllers, businessNames[i], i)
		nodes = append(nodes, node)

		//rpc server
		if _, ok := businesses[i].(RpcConfigurable); ok {
			serviceName, bufferSize := businesses[i].(RpcConfigurable).RpcConfig()
			serviceName = strings.Trim(serviceName, " ")
			if len(serviceName) > 0 {
				rpc.NewRPCService("rpc."+serviceName, actor, broker, bufferSize)
				// Only checks for duplicate service names.
				if err := rpc.GlobalRPCFactory.Register("rpc."+serviceName, businesses[i]); err != nil {
					panic(err)
				}
				actor.rpcIndex["rpc."+serviceName] = node
				withRpcServer = true
			}
		}
	}
	actor.chain = NewBusinessChain(nodes, filters)

	//RegisterProducer
	if list_outputs != nil || len(list_outputs) > 0 {
		outputs := DeduplicationOutputs(list_outputs)
		publishTo := make([]string, 0, len(outputs))
		bufferLen := make([]int, 0, len(outputs))
		for k, v := range outputs {
			publishTo = append(publishTo, k)
			bufferLen = append(bufferLen, v)
		}
		if withRpcServer {
			publishTo = append(publishTo, rpccomplete)
			bufferLen = append(bufferLen, 1)
		}
		actor.broker.RegisterProducer(brokerpk.NewDefaultProducer(name+"-producer", publishTo, bufferLen))
	}

	//RegisterConsumer
	inputs := DeduplicationInputs(list_inputs)
	if withRpcServer {
		inputs = append(inputs, rpccomplete)
	}
	var controller brokerpk.StreamController
	if chainIsConjunction {
		controller = brokerpk.NewConjunctions(actor)
	} else {
		controller = brokerpk.NewDisjunctions(actor, 1)
	}
	actor.broker.RegisterConsumer(brokerpk.NewDefaultConsumer(actor.name+"-consumer", inputs, controller))

	actor.execCtx.actor = actor

	go actor.Serve()
	return actor
}

func (a *Actor) RegisterContinuation(id string, c Continuation) {
	a.continuations.Register(id, c, rpc.RpcTimeout)
}

func (a *Actor) Consume(evt brokerpk.ActorEvent) {
	a.inbox <- evt
}
func (a *Actor) onExit() {
	a.continuations.CancelAll(
		fmt.Errorf("actor inbox closed"),
	)
}
func (a *Actor) Serve() {
	defer a.onExit()

	for evt := range a.inbox {
		switch e := evt.(type) {

		case *ContinuationNext:
			a.onContinuationNext(e)

		case *brokerpk.RPCCompletion:
			a.onCompletion(e)

		case *brokerpk.RPCInvocation:
			a.onInvocation(e)

		case []interface{}:
			a.onMessage(a.parseParams(e))

		case brokerpk.Aggregated:
			a.parseAggregated(e)

		default:
			// unknown event
		}
	}
}

func (a *Actor) onMessage(msgs []*scommon.Message) {
	a.log(msgs)

	if err := a.chain.OnMessages(msgs, a.execCtx); err != nil {
		logger.Log.Error(
			context.Background(),
			a.name+" execute msg Err",
			logger.F("err", err),
		)
	}
}

func (a *Actor) ParseStep(step string) (string, *ControlledNode) {
	spans := strings.Split(step, ".")
	if len(spans) != 2 {
		return "", nil
	}
	idx, err := strconv.Atoi(spans[0])
	if err != nil {
		logger.Log.Error(context.Background(), "ContinuationNext step err", logger.F("step", step))
		return "", nil
	}
	return spans[1], a.chain.GetNode(idx)
}

func (a *Actor) onContinuationNext(inv *ContinuationNext) {
	rpcmsg := inv.Payload.(*scommon.Message)
	// node := a.rpcIndex[rpcmsg.Service]

	step, node := a.ParseStep(inv.Step)

	if node == nil {
		logger.Log.Error(context.Background(), "rpc request node not found")
		return
	}

	rpcCtx := &RPCContext{
		Request: rpcmsg.Data,
	}

	ctx := scommon.BindLoggerContextFromMessageSafe(context.Background(), rpcmsg)
	logger.Log.Debug(ctx, a.name+" start rpc Request next")

	if step == "" {
		logger.Log.Warn(ctx, "rpc continuation without next step")
		return
	}

	if inv.Error != "" {
		node.StartRPC("onRPCError", []*scommon.Message{rpcmsg}, rpcCtx, a.execCtx)
		return
	}

	node.StartRPC(
		step,
		[]*scommon.Message{rpcmsg},
		rpcCtx,
		a.execCtx,
	)
}

func (a *Actor) onInvocation(inv *brokerpk.RPCInvocation) {
	node := a.rpcIndex[inv.ServerName]
	if node == nil {
		logger.Log.Error(context.Background(), inv.ServerName+" not configure")
		return
	}

	reqID := inv.ID

	rpcmsg := inv.Payload.(*scommon.Message)
	rpcmsg.ReplyTo = inv.ReplyTo

	ctx := scommon.BindLoggerContextFromMessageSafe(context.Background(), rpcmsg)
	logger.Log.Debug(ctx, a.name+" received rpc Request")

	a.RegisterContinuation(reqID, NewDefaultContinuation(a.broker, a.name, rpcmsg))

	rpcCtx := &RPCContext{
		Request: rpcmsg.Data,
	}

	node.StartRPC(
		inv.Method,
		[]*scommon.Message{rpcmsg},
		rpcCtx,
		a.execCtx,
	)
}

func (a *Actor) onCompletion(comp *brokerpk.RPCCompletion) {
	defer func() {
		if r := recover(); r != nil {
			logger.Log.Error(context.Background(), "panic in continuation", logger.F("id", comp.ID), logger.F("err", r))
		}
	}()

	if !a.continuations.Resume(comp.ID, comp) {
		logger.Log.Warn(
			context.Background(), "RPCCompletion dropped: no continuation",
			logger.F("id", comp.ID),
		)
	}
}

func (actor *Actor) log(msgs []*scommon.Message) {
	for _, v := range msgs {
		ctx := scommon.BindLoggerContextFromMessageSafe(context.Background(), v)
		logger.Log.Debug(ctx, actor.name+" received Msg")
	}
}

func (actor *Actor) parseParams(data []interface{}) []*scommon.Message {
	msgs := make([]*scommon.Message, len(data))
	for _, pp := range data {
		msgs = append(msgs, pp.(*scommon.Message))
	}
	return msgs
}

func (a *Actor) parseAggregated(data brokerpk.Aggregated) {
	switch d := data.Data.(type) {
	case *broker.RPCCompletion:
		a.onCompletion(d)
	case *scommon.Message:
		a.onMessage([]*scommon.Message{d})
	}
}

func DeduplicationInputs(inputs [][]string) []string {
	totalSize := 0
	for i := range inputs {
		totalSize += len(inputs[i])
	}
	pools := make(map[string]int, totalSize)
	for i := range inputs {
		for j := range inputs[i] {
			pools[inputs[i][j]] = 0
		}
	}
	outputs := make([]string, 0, len(pools))
	for input := range pools {
		outputs = append(outputs, input)
	}
	return outputs
}

func DeduplicationOutputs(inputs []map[string]int) map[string]int {
	totalSize := 0
	for i := range inputs {
		totalSize += len(inputs[i])
	}
	pools := make(map[string]int, totalSize)
	for i := range inputs {
		for k, v := range inputs[i] {
			pools[k] = v
		}
	}
	return pools
}
