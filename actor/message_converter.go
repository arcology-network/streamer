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
	"sync"

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

	primaryMsg string

	callframes map[string]*CallFrame
	cfLock     sync.Mutex
}

func CreateActor(name string, broker *brokerpk.StatefulStreamer, businesses []Business, businessNames []string, concurrency int, rpcServers []string) *Actor {
	rpccomplete := name + ".rpc.complete"
	actor := &Actor{
		inbox:            make(chan brokerpk.ActorEvent, 1024),
		name:             name,
		broker:           broker,
		execCtx:          NewExecutionContext(name, broker, concurrency),
		rpcIndex:         make(map[string]*ControlledNode, len(businesses)),
		rpcCompleteTopic: rpccomplete,

		continuations: NewContinuationManager(),
		primaryMsg:    "",
		callframes:    make(map[string]*CallFrame),
	}

	//create chain
	list_inputs := make([][]string, 0, len(businesses))
	list_outputs := make([]map[string]int, len(businesses))
	nodes := make([]*ControlledNode, 0, len(businesses))
	chainIsConjunction := false
	// withRpcServer := false
	for i := range businesses {
		//create businesschain
		inputs, isConjunction := businesses[i].Inputs()
		list_inputs = append(list_inputs, inputs)
		list_outputs = append(list_outputs, businesses[i].Outputs())

		if isConjunction && len(businesses) > 1 {
			panic("All elements in the pipeline must be DisConjunction.")
		}

		if isConjunction {
			if _, ok := businesses[i].(PickPrimary); ok {
				actor.primaryMsg = businesses[i].(PickPrimary).PrimaryMsg()
			} else {
				panic(businessNames[i] + " is conjunction , that must implement PickPrimary interface.")
			}
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
				controllers = append(controllers, NewHeightController(businesses[i].(HeightSensitive), businessNames[i]))
			}
			if _, ok := businesses[i].(FSMCompatible); ok {
				controllers = append(controllers, NewFSMController(businesses[i].(FSMCompatible), businessNames[i]))
			}
		}
		node := NewControlledNode(businesses[i], controllers, businessNames[i], i)
		nodes = append(nodes, node)

		//rpc server
		if _, ok := businesses[i].(RpcConfigurable); ok {
			serviceName, bufferSize := businesses[i].(RpcConfigurable).RpcConfig()
			if len(rpcServers) > 0 && len(rpcServers[i]) > 0 {
				serviceName = rpcServers[i] //When the RPC server is part of a cluster, enable this configuration.
			}
			serviceName = strings.Trim(serviceName, " ")
			if len(serviceName) > 0 {
				rpc.NewRPCService("rpc."+serviceName, actor, broker, bufferSize)
				// Only checks for duplicate service names.
				if err := rpc.GlobalRPCFactory.Register("rpc."+serviceName, businesses[i]); err != nil {
					panic(err)
				}
				actor.rpcIndex["rpc."+serviceName] = node

			}
		}
	}
	actor.chain = NewBusinessChain(nodes)

	actor.broker.RegisterProducer(brokerpk.NewDefaultProducer(actor.name+"-rpc-producer", []string{rpccomplete}, []int{1}))
	actor.broker.RegisterConsumer(
		brokerpk.NewDefaultConsumer(
			actor.name+"-rpc-consumer",
			[]string{rpccomplete},
			brokerpk.NewDisjunctions(actor, 1),
		))

	//RegisterProducer
	if list_outputs != nil || len(list_outputs) > 0 {
		outputs := DeduplicationOutputs(list_outputs)
		publishTo := make([]string, 0, len(outputs))
		bufferLen := make([]int, 0, len(outputs))
		for k, v := range outputs {
			publishTo = append(publishTo, k)
			bufferLen = append(bufferLen, v)
		}

		actor.broker.RegisterProducer(brokerpk.NewDefaultProducer(name+"-producer", publishTo, bufferLen))
	}

	//RegisterConsumer
	inputs := DeduplicationInputs(list_inputs)

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

func (a *Actor) registerFrame(cf *CallFrame) {
	a.cfLock.Lock()
	a.callframes[cf.ReqID] = cf
	a.cfLock.Unlock()
}

func (a *Actor) getFrame(reqID string) *CallFrame {
	a.cfLock.Lock()
	defer a.cfLock.Unlock()
	return a.callframes[reqID]
}

func (a *Actor) removeFrame(reqID string) {
	a.cfLock.Lock()
	delete(a.callframes, reqID)
	a.cfLock.Unlock()
}

func PickPrimaryMsg(msgs []*scommon.Message, primaryMsg string) *scommon.Message {
	if len(primaryMsg) == 0 {
		return msgs[0]
	} else {
		for i := range msgs {
			if msgs[i] != nil && msgs[i].Name == primaryMsg {
				return msgs[i]
			}
		}
	}

	return nil
}

func (a *Actor) SetFilters(filters []*Filter) {
	a.chain.SetFilters(filters)
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
	nmsgs := make([]*scommon.Message, 0, len(msgs))
	for i := range msgs {
		if msgs[i] != nil {
			nmsgs = append(nmsgs, msgs[i])
		}
	}
	if len(nmsgs) == 0 {
		logger.Log.Error(
			context.Background(),
			a.name, "Received msg is nil",
		)
		return
	}
	a.log(nmsgs)

	primaryMsg := PickPrimaryMsg(msgs, a.primaryMsg)
	if primaryMsg != nil && primaryMsg.ReqID != "" {
		cf := a.getFrame(primaryMsg.ReqID)
		if cf == nil {
			// create temp frame(only use for trace,not add into global map )
			cf = &CallFrame{ReqID: primaryMsg.ReqID, ReqMsg: primaryMsg, Parent: a.execCtx.Current}
		}
		a.execCtx.Current = cf
	} else {
		a.execCtx.Current = nil
	}

	if err := a.chain.OnMessages(nmsgs, a.execCtx); err != nil {
		logger.Log.Error(
			context.Background(),
			a.name, "Execute msg Err",
			logger.F("err", err),
		)
	}
}

func CallFrameFrom(msg *scommon.Message) *CallFrame {
	return &CallFrame{
		ReqID:   msg.ReqID,
		ReqMsg:  msg,
		ReplyTo: msg.ReplyTo,
	}
}

func (a *Actor) ParseStep(step string) (string, *ControlledNode) {
	nextStep, idx := NextStep(a.name, step)
	return nextStep, a.chain.GetNode(idx)
}

func NextStep(name string, step string) (string, int) {
	spans := strings.Split(step, ".")
	if len(spans) != 2 {
		return "", -1
	}
	idx, err := strconv.Atoi(spans[0])
	if err != nil {
		logger.Log.Error(context.Background(), name, "ContinuationNext step err", logger.F("step", step))
		return "", -1
	}
	return spans[1], idx
}

func (a *Actor) onContinuationNext(inv *ContinuationNext) {
	rpcmsg := inv.Payload.(*scommon.Message)
	// node := a.rpcIndex[rpcmsg.Service]

	step, node := a.ParseStep(inv.Step)

	if node == nil {
		logger.Log.Error(context.Background(), a.name, "Rpc request node not found")
		return
	}

	rpcCtx := &RPCContext{
		Request: rpcmsg.Data,
	}

	ctx := scommon.BindLoggerContextFromMessageSafe(context.Background(), rpcmsg)
	logger.Log.Debug(ctx, a.name, "Start rpc Request next")

	a.execCtx.Current = inv.Frame

	if step == "" {
		logger.Log.Warn(ctx, a.name, "Rpc continuation without next step")
		return
	}

	if inv.Error != "" {
		node.StartRPC("onRPCError", []*scommon.Message{rpcmsg}, rpcCtx, a.execCtx)
		return
	}

	if inv.Frame != nil {
		a.execCtx.Current = inv.Frame.Parent
	}

	err := node.StartRPC(
		step,
		[]*scommon.Message{rpcmsg},
		rpcCtx,
		a.execCtx,
	)
	if err != nil {
		logger.Log.Error(context.Background(), a.name, err.Error())
	}
}

func (a *Actor) onInvocation(inv *brokerpk.RPCInvocation) {
	node := a.rpcIndex[inv.ServerName]
	if node == nil {
		logger.Log.Error(context.Background(), a.name, inv.ServerName+" not configure")
		return
	}

	rpcmsg := inv.Payload.(*scommon.Message)

	rpcmsg.ReplyTo = inv.ReplyTo
	cf := &CallFrame{
		ReqID:   rpcmsg.ReqID,
		ReqMsg:  rpcmsg,
		ReplyTo: rpcmsg.ReplyTo,
	}

	a.registerFrame(cf)

	execCtx := a.execCtx.Fork()
	execCtx.Current = cf
	ctx := scommon.BindLoggerContextFromMessageSafe(context.Background(), rpcmsg)
	logger.Log.Debug(ctx, a.name, "Received rpc Request")

	a.RegisterContinuation(cf.ReqID, NewDefaultContinuation(a.broker, a.name, cf))

	rpcCtx := &RPCContext{
		Request: rpcmsg.Data,
	}

	err := node.StartRPC(
		inv.Method,
		[]*scommon.Message{rpcmsg},
		rpcCtx,
		execCtx,
	)
	if err != nil {
		logger.Log.Error(context.Background(), a.name, err.Error())
	}

}

func (a *Actor) onCompletion(comp *brokerpk.RPCCompletion) {
	defer func() {
		if r := recover(); r != nil {
			logger.Log.Error(context.Background(), a.name, "panic in continuation", logger.F("id", comp.ID), logger.F("err", r))
		}
	}()

	// log.Printf("[Rpc] Actor onCompletion -- MsdId:%v", comp.ID)
	if !a.continuations.Resume(comp.ID, comp) {
		logger.Log.Warn(
			context.Background(), a.name, "RPCCompletion dropped: no continuation",
			logger.F("id", comp.ID),
		)
	}
}

func (a *Actor) log(msgs []*scommon.Message) {
	for _, v := range msgs {
		ctx := scommon.BindLoggerContextFromMessageSafe(context.Background(), v)
		logger.Log.Debug(ctx, a.name, "Received Msg")
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
