package actor

import (
	"context"
	"fmt"

	"github.com/arcology-network/streamer/actor/rpc"
	"github.com/arcology-network/streamer/broker"
	brokerpk "github.com/arcology-network/streamer/broker"
	scommon "github.com/arcology-network/streamer/common"
	"github.com/arcology-network/streamer/logger"
	"github.com/google/uuid"
)

type WorkContext struct {
	ActorName        string
	BusinassName     string
	NodeIdx          int
	RpcCompleteTopic string
	Concurrency      int

	Height          uint64
	ParentRequestId string
}

type ExecutionContext struct {
	// ==== Message-related ====
	WorkCtx  *WorkContext
	traceCtx *logger.TraceContext

	// ==== Capability Proxy ====
	msgSender *broker.StatefulStreamer
	rpcCaller *rpc.RPCClient
	actor     *Actor
}

func NewExecutionContext(ActorName string, rpcCompleteTopic string, broker *broker.StatefulStreamer, concurrency int) *ExecutionContext {
	return &ExecutionContext{
		msgSender: broker,
		rpcCaller: rpc.GlobalRPCClient,
		WorkCtx: &WorkContext{
			ActorName:        ActorName,
			RpcCompleteTopic: rpcCompleteTopic,
			Concurrency:      concurrency,
		},
		traceCtx: &logger.TraceContext{},
	}
}

func (ctx *ExecutionContext) Copy() *ExecutionContext {
	return &ExecutionContext{
		msgSender: ctx.msgSender,
		rpcCaller: ctx.rpcCaller,
		WorkCtx: &WorkContext{
			ActorName:        ctx.WorkCtx.ActorName,
			BusinassName:     ctx.WorkCtx.BusinassName,
			Height:           ctx.WorkCtx.Height,
			RpcCompleteTopic: ctx.WorkCtx.RpcCompleteTopic,
			Concurrency:      ctx.WorkCtx.Concurrency,
		},
		traceCtx: &logger.TraceContext{
			TraceID:  ctx.traceCtx.TraceID,
			SpanID:   ctx.traceCtx.SpanID,
			ParentID: ctx.traceCtx.ParentID,
			SpanName: ctx.traceCtx.SpanName,
			ReqID:    ctx.traceCtx.ReqID,
		},
	}
}

func (ctx *ExecutionContext) WithBusinessName(businessName string) *ExecutionContext {
	ctx.WorkCtx.BusinassName = businessName
	return ctx
}
func (ctx *ExecutionContext) Concurrency() int {
	return ctx.WorkCtx.Concurrency
}

func (ctx *ExecutionContext) StartCasecade() {
	ctx.WorkCtx.ParentRequestId = ctx.traceCtx.ReqID
}
func (ctx *ExecutionContext) EndCasecade() {
	ctx.traceCtx.ReqID = ctx.WorkCtx.ParentRequestId
}

func (ctx *ExecutionContext) InvokeRPC(
	target string,
	method string,
	payload any,
	nextStep string,
	args ...uint64,
) {
	if len(args) > 0 {
		ctx.WorkCtx.Height = args[0]
	}

	allServerName := "rpc." + target
	msg := scommon.NewMessageForRPCREQ(allServerName, method, payload)
	msg.NextStep = fmt.Sprintf("%v.%v", ctx.WorkCtx.NodeIdx, nextStep)
	msg = ctx.AddTrace(msg)
	msg.StartRpcTrace(msg.ID)

	ctxlog := scommon.BindLoggerContextFromMessageSafe(context.Background(), msg)
	logger.Log.Debug(ctxlog, "Send rpc Msg")

	inv := &brokerpk.RPCInvocation{
		ID:         msg.ID,
		ServerName: allServerName,
		Method:     method,
		Payload:    msg,
		ReplyTo:    ctx.WorkCtx.RpcCompleteTopic,
	}

	ctx.actor.RegisterContinuation(msg.ID, NewBusinessContinuation(ctx.actor))

	ctx.msgSender.Send(allServerName, inv)
	ctx.traceCtx.ParentID = msg.SpanID
}

func (ctx *ExecutionContext) StartTrace() {
	ctx.traceCtx.TraceID = uuid.NewString()
}

func (ctx *ExecutionContext) AddTrace(msg *scommon.Message) *scommon.Message {
	msg.ParentID = ctx.traceCtx.ParentID
	msg.SpanID = uuid.NewString()
	msg.From = ctx.WorkCtx.BusinassName
	msg.Height = ctx.WorkCtx.Height
	msg.TraceID = ctx.traceCtx.TraceID
	msg.SpanName = ctx.traceCtx.SpanName
	msg.ReqID = ctx.traceCtx.ReqID
	return msg
}

func (ctx *ExecutionContext) SendRpcResponse(err string, data interface{}, args ...uint64) {
	msg := scommon.NewMessageForStream(ctx.WorkCtx.RpcCompleteTopic, data)
	msg = ctx.AddTrace(msg)
	msg.Error = err

	ret := &brokerpk.RPCCompletion{
		ID:      msg.ReqID,
		Payload: msg,
		Error:   "",
	}
	ctx.msgSender.Send(ctx.WorkCtx.RpcCompleteTopic, ret)

	ctxlog := scommon.BindLoggerContextFromMessageSafe(context.Background(), msg)
	logger.Log.Debug(ctxlog, "Send Msg")

	//for next send
	ctx.traceCtx.ParentID = msg.SpanID
}

func (ctx *ExecutionContext) Send(name string, data interface{}, args ...uint64) {
	if len(args) > 0 {
		ctx.WorkCtx.Height = args[0]
	}

	msg := scommon.NewMessageForStream(name, data)
	msg = ctx.AddTrace(msg)
	ctx.msgSender.Send(name, msg)

	ctxlog := scommon.BindLoggerContextFromMessageSafe(context.Background(), msg)
	logger.Log.Debug(ctxlog, "Send Msg")

	//for next send
	ctx.traceCtx.ParentID = msg.SpanID
}

func (ctx *ExecutionContext) GetReqID() string {
	return ctx.traceCtx.ReqID
}

func (ctx *ExecutionContext) AddRpcReqID(reqId string) {
	ctx.traceCtx.ReqID = reqId
}

func (ctx *ExecutionContext) GetCtx() context.Context {
	ctx.traceCtx.SpanID = uuid.NewString()
	return logger.WithTrace(context.Background(), *ctx.traceCtx)
}

// -------log
func (ctx *ExecutionContext) LogDebug(event string, fs ...logger.Field) {
	logger.Log.Debug(ctx.GetCtx(), event, fs...)
}

func (ctx *ExecutionContext) LogInfo(event string, fs ...logger.Field) {
	logger.Log.Info(ctx.GetCtx(), event, fs...)
}

func (ctx *ExecutionContext) LogWarn(event string, fs ...logger.Field) {
	logger.Log.Warn(ctx.GetCtx(), event, fs...)
}
func (ctx *ExecutionContext) LogErr(event string, fs ...logger.Field) {
	logger.Log.Error(ctx.GetCtx(), event, fs...)
}
func (ctx *ExecutionContext) LogFatal(event string, fs ...logger.Field) {
	logger.Log.Fatal(ctx.GetCtx(), event, fs...)
}
