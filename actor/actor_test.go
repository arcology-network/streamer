package actor

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/arcology-network/streamer/actor/rpc"
	"github.com/arcology-network/streamer/broker"
	scommon "github.com/arcology-network/streamer/common"
	"github.com/arcology-network/streamer/logger"
	"github.com/google/uuid"
)

type BusinessConjunction struct {
	state       int
	execCounter int
}

func NewBusinessConjunction() *BusinessConjunction {
	a := &BusinessConjunction{}

	a.execCounter = 0
	return a
}

func (a *BusinessConjunction) exec_msgA_msgB(ctx *ActionContext) error {
	fmt.Println("Action conjunction msgA/B executed with msgs:", len(ctx.Messages))
	a.execCounter++
	return nil
}

func (a *BusinessConjunction) RPC_Test(ctx *ActionContext) error {
	fmt.Println("RPC Action executed, request:", ctx.RPC.Request)
	ctx.RPC.Response = "RPC result"
	return nil
}

func (a *BusinessConjunction) RegisterActions(reg ActionRegistrar) {
	reg.Register("msgA", a.exec_msgA_msgB)
	reg.Register("msgB", a.exec_msgA_msgB)
	reg.Register("RPC_Test", a.RPC_Test)
}

func (a *BusinessConjunction) Inputs() ([]string, bool) {
	return []string{"msgA", "msgB"}, true
}
func (a *BusinessConjunction) Outputs() map[string]int {
	return map[string]int{}
}
func (a *BusinessConjunction) PrimaryMsg() string {
	return "msgA"
}

func TestHeightFsmConjunction(t *testing.T) {
	logger.InitLog("./log.toml", "")
	rpc.InitGlobalRPCFactory()
	ss := broker.NewStatefulStreamer()
	rpc.InitGlobalRPCClient(ss, 10, 60)

	concurrency := runtime.NumCPU()
	businessname := "businessConjunction"
	a := NewBusinessConjunction()
	act := CreateActor(businessname, ss, []Business{a}, []string{businessname}, concurrency, []string{""})

	ss.Serve()

	msgs := []*scommon.Message{
		{Name: "msgA", Height: 2},
		{Name: "msgB", Height: 2},
	}

	err := act.chain.OnMessages(msgs, act.execCtx)
	if err != nil {
		t.Error(err)
	}

	if a.execCounter != 1 {
		t.Errorf("controller test err,a.execCounter:%v", a.execCounter)
	}

}

//-----------------------------------------------------------

type BusinessDisConjunction struct {
	state  int
	height uint64

	execCounter int
}

func NewBusinessDisConjunction() *BusinessDisConjunction {
	a := &BusinessDisConjunction{}
	a.height = 1
	a.state = 0
	a.execCounter = 0
	return a
}

// FSMCompatible interface
func (a *BusinessDisConjunction) GetFSMRules() map[int]FSMRule {
	return map[int]FSMRule{
		0: {Accept: []string{"msg1", "msg2"}},
		1: {Accept: []string{"msgA"}},
	}
}

func (a *BusinessDisConjunction) GetCurrentState() int {
	return a.state
}

func (a *BusinessDisConjunction) exec_msg1(ctx *ActionContext) error {
	ctx.ExecCtx.LogDebug("Action msg1 executed", logger.F("height", ctx.Messages[0].Height))
	a.height = ctx.Messages[0].Height + 1
	a.execCounter++
	return nil
}

func (a *BusinessDisConjunction) exec_msg2(ctx *ActionContext) error {
	ctx.ExecCtx.LogDebug("Action msg2 executed", logger.F("height", ctx.Messages[0].Height))
	a.state = 1
	a.height = ctx.Messages[0].Height + 1
	a.execCounter++
	return nil
}

func (a *BusinessDisConjunction) exec_msgA_msgB(ctx *ActionContext) error {
	ctx.ExecCtx.LogDebug("Action conjunction msgA/B executed", logger.F("msgs", len(ctx.Messages)))
	a.execCounter++
	a.state = 0
	return nil
}

func (a *BusinessDisConjunction) RPC_internal_Test(ctx *ActionContext) error {
	ctx.ExecCtx.LogDebug("RPC_internal_Test Action executed", logger.F("request", ctx.RPC.Request))
	ctx.ExecCtx.SendRpcResponse("", fmt.Sprintf("%v Tom,I am internal rpc server", ctx.RPC.Request))
	logger.Log.Debug(context.Background(), "actorTest", "startRpc", logger.F("reqID", ctx.Messages[0].ReqID))
	a.execCounter++
	return nil
}

func (a *BusinessDisConjunction) externalCasecadeLine1(ctx *ActionContext) error {
	ctx.ExecCtx.LogDebug("externalCasecadeLine1RPC_internal_Test Action executed", logger.F("request", ctx.RPC.Request))
	// ctx.ExecCtx.StartCasecade()
	ctx.ExecCtx.InvokeRPC("collection", "externalCasecadeLine1", fmt.Sprintf("%v -- collection", ctx.RPC.Request), "externalCasecadeLine1_resp")
	return nil
}

func (a *BusinessDisConjunction) externalCasecadeLine1_resp(ctx *ActionContext) error {
	// ctx.ExecCtx.EndCasecade()
	ctx.ExecCtx.SendRpcResponse("", fmt.Sprintf("%v -- sender", ctx.RPC.Request))
	return nil
}

func (a *BusinessDisConjunction) externalCasecadeLine2(ctx *ActionContext) error {
	ctx.ExecCtx.SendRpcResponse("", fmt.Sprintf("%v -- sender", ctx.RPC.Request))
	return nil
}

func (a *BusinessDisConjunction) RpcConfig() (string, int) {
	return "business", 20
}

func (a *BusinessDisConjunction) Height() uint64 {
	return a.height
}

func (a *BusinessDisConjunction) Inputs() ([]string, bool) {
	return []string{"msg1", "msg2", "msgA"}, false
}

func (a *BusinessDisConjunction) Outputs() map[string]int {
	return map[string]int{}
}

// Action rigister
func (a *BusinessDisConjunction) RegisterActions(reg ActionRegistrar) {
	reg.Register("msg1", a.exec_msg1)
	reg.Register("msg2", a.exec_msg2)
	reg.Register("msgA", a.exec_msgA_msgB)
	reg.Register("RPC_internal_Test", a.RPC_internal_Test)
	reg.Register("externalCasecadeLine1", a.externalCasecadeLine1)
	reg.Register("externalCasecadeLine1_resp", a.externalCasecadeLine1_resp)

	reg.Register("externalCasecadeLine2", a.externalCasecadeLine2)
}

// -----------------------------------------------------------
type BusinessCollect struct {
}

func NewBusinessCollect() *BusinessCollect {
	a := &BusinessCollect{}
	return a
}

func (a *BusinessCollect) Inputs() ([]string, bool) {
	return []string{"msgD"}, false
}

func (a *BusinessCollect) Outputs() map[string]int {
	return map[string]int{
		"msgE": 10,
	}
}
func (a *BusinessCollect) RpcConfig() (string, int) {
	return "collection", 20
}

// Action
func (a *BusinessCollect) RegisterActions(reg ActionRegistrar) {
	reg.Register("msgD", a.exec_msgD)
	reg.Register("externalCasecadeLine1", a.externalCasecadeLine1)
}

func (a *BusinessCollect) exec_msgD(ctx *ActionContext) error {
	logger.Log.Debug(context.Background(), "actorTest", "startRpc", logger.F("data", ctx.Messages[0].Data), logger.F("reqID", ctx.Messages[0].ReqID))
	ctx.ExecCtx.Send("msgE", fmt.Sprintf("%v clol ", ctx.Messages[0].Data))
	return nil
}
func (a *BusinessCollect) externalCasecadeLine1(ctx *ActionContext) error {
	ctx.ExecCtx.SendRpcResponse("", fmt.Sprintf("%v -- business", ctx.RPC.Request))
	return nil
}

//----------------------------------------------------------------------------------

type BusinessSender struct {
	sender       OutboundSender
	line1, line2 string
}

func NewBusinessSender() *BusinessSender {
	a := &BusinessSender{}
	return a
}

func (a *BusinessSender) SetSender(sender OutboundSender) {
	a.sender = sender
}

func (a *BusinessSender) Inputs() ([]string, bool) {
	return []string{"internalrpc", "msgE"}, false
}

func (a *BusinessSender) Outputs() map[string]int {
	return map[string]int{
		"msgA":        10,
		"msgB":        20,
		"msg1":        10,
		"msg2":        20,
		"internalrpc": 10,
		"msgD":        10,
	}
}
func (a *BusinessSender) RpcConfig() (string, int) {
	return "sender", 20
}

// Action register
func (a *BusinessSender) RegisterActions(reg ActionRegistrar) {
	reg.Register("externalrpc", a.externalrpc)
	reg.Register("internalrpc", a.internalrpc)
	reg.Register("internalrpc_resp", a.internalrpc_resp)

	reg.Register("externalrpcToAsync", a.externalrpcToAsync)
	reg.Register("msgE", a.externalrpcToAsync_resp)

	reg.Register("externalrpccasecade", a.externalrpccasecade)
	reg.Register("externalrpccasecade_next", a.externalrpccasecade_next)
	reg.Register("externalrpccasecade_resp", a.externalrpccasecade_resp)
}

func (a *BusinessSender) externalrpccasecade(ctx *ActionContext) error {
	// ctx.ExecCtx.StartCasecade()
	ctx.ExecCtx.InvokeRPC("business", "externalCasecadeLine2", "line2:sender -- business", "externalrpccasecade_next")
	return nil
}

func (a *BusinessSender) externalrpccasecade_next(ctx *ActionContext) error {
	logger.Log.Debug(context.Background(), "actorTest", "externalrpccasecade line 2", logger.F("ret", ctx.RPC.Request))
	a.line2 = ctx.RPC.Request.(string)
	ctx.ExecCtx.InvokeRPC("business", "externalCasecadeLine1", "line1:sender -- business", "externalrpccasecade_resp")
	return nil
}

func (a *BusinessSender) externalrpccasecade_resp(ctx *ActionContext) error {
	logger.Log.Debug(context.Background(), "actorTest", "externalrpccasecade line 1", logger.F("ret", ctx.RPC.Request))
	a.line1 = ctx.RPC.Request.(string)
	// ctx.ExecCtx.EndCasecade()
	ctx.ExecCtx.SendRpcResponse("", fmt.Sprintf("%v     %v ", a.line1, a.line2))
	return nil
}

func (a *BusinessSender) externalrpc(ctx *ActionContext) error {
	ctx.ExecCtx.SendRpcResponse("", fmt.Sprintf("%v Tom,I am external rpc server", ctx.RPC.Request))
	return nil
}

func (a *BusinessSender) externalrpcToAsync(ctx *ActionContext) error {
	ctx.ExecCtx.Send("msgD", ctx.RPC.Request)
	return nil
}
func (a *BusinessSender) externalrpcToAsync_resp(ctx *ActionContext) error {
	ctx.ExecCtx.SendRpcResponse("", fmt.Sprintf("%v Tom,I am external rpc server from collect", ctx.Messages[0].Data))
	return nil
}

func (a *BusinessSender) internalrpc(ctx *ActionContext) error {
	ctx.ExecCtx.InvokeRPC("business", "RPC_internal_Test", "hello", "internalrpc_resp")
	return nil
}

func (a *BusinessSender) internalrpc_resp(ctx *ActionContext) error {
	logger.Log.Debug(context.Background(), "actorTest", "internalrpc_resp return", logger.F("ret", ctx.RPC.Request))
	return nil
}

func (a *BusinessSender) send(ss *broker.StatefulStreamer) {
	traceId := uuid.NewString()
	spanId := uuid.NewString()
	msgs := []*scommon.Message{
		{Name: "msg1", Height: 1, From: "sender", TraceID: traceId, SpanID: spanId},
		{Name: "msgB", Height: 1, From: "sender", TraceID: traceId, SpanID: spanId},
		{Name: "msg1", Height: 3, From: "sender", TraceID: traceId, SpanID: spanId},
		{Name: "msg1", Height: 2, From: "sender", TraceID: traceId, SpanID: spanId},
		{Name: "msg2", Height: 3, From: "sender", TraceID: traceId, SpanID: spanId},
		{Name: "msgA", Height: 4, From: "sender", TraceID: traceId, SpanID: spanId},
		{Name: "internalrpc", Height: 1, From: "sender", TraceID: traceId, SpanID: spanId}, //async
	}

	for i := range msgs {
		ss.Send(msgs[i].Name, msgs[i])
		time.Sleep(1 * time.Second)
	}

	//sync
	ret, _ := a.sender.SendSync("sender", "externalrpc", "hello", 10)
	logger.Log.Debug(context.Background(), "****actorTest", "externalrpc_resp return", logger.F("ret", ret))
	time.Sleep(1 * time.Second)

	// //sync to async
	ret, _ = a.sender.SendSync("sender", "externalrpcToAsync", "hello", 10)
	logger.Log.Debug(context.Background(), "****actorTest", "externalrpcToAsync return", logger.F("ret", ret))
	time.Sleep(1 * time.Second)

	// //casecade
	ret, _ = a.sender.SendSync("sender", "externalrpccasecade", "hello", 10)
	logger.Log.Debug(context.Background(), "****actorTest", "externalrpccasecade return", logger.F("ret", ret))
	time.Sleep(1 * time.Second)
}

/*
stream message FSM & height

	*external -->msg1       -->DisConjunction
	           -->msgB
			   -->msg1		 -->DisConjunction
			   -->msg1       -->DisConjunction
			   -->msg2       -->DisConjunction
			   -->msgA		 -->DisConjunction

rpc single

	external   -->sender.externalrpc
	external   -->sender.internalrpc   -->  business.RPC_internal_Test

sync to async

	external   -->sender.externalrpcToAsync -->msgD -->Collect  --> msgE  --> sender

rpc casecade

	external   -->sender.externalrpccasecade   -->business.externalCasecadeLine2
	                                           -->business.externalCasecadeLine1   -->Collect.externalCasecadeLine1
*/
func TestActorRpcController(t *testing.T) {
	logger.InitLog("./log.toml", "")
	rpc.InitGlobalRPCFactory()
	ss := broker.NewStatefulStreamer()
	rpc.InitGlobalRPCClient(ss, 10, 60)

	filter := NewOriginFilter(
		"origin-filter",
		MsgsOnlyFrom([]string{"msg1", "msg2"}, "sender"),
	)

	concurrency := runtime.NumCPU()

	a := NewBusinessDisConjunction()
	act := CreateActor("BusinessDisConjunction", ss, []Business{a}, []string{"businessDisConjunction"}, concurrency, []string{""})
	act.SetFilters([]*Filter{filter})

	b := NewBusinessSender()
	b.SetSender(NewSendAdaptor(ss, rpc.GlobalRPCClient))
	CreateActor("sender", ss, []Business{b}, []string{"businessSender"}, concurrency, []string{""})

	c := NewBusinessCollect()
	CreateActor("collect", ss, []Business{c}, []string{"businessCollect"}, concurrency, []string{""})

	ss.Serve()

	b.send(ss)

	if a.execCounter != 6 {
		t.Errorf("controller test err,a.execCounter:%v", a.execCounter)
	}
}
