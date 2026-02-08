package actor

import (
	"context"
	"fmt"
	"sync"

	"github.com/arcology-network/streamer/actor/rpc"
	"github.com/arcology-network/streamer/broker"
	brokerpk "github.com/arcology-network/streamer/broker"
	scommon "github.com/arcology-network/streamer/common"
	"github.com/arcology-network/streamer/logger"

	// "github.com/arcology-network/streamer/query"
	"github.com/google/uuid"
)

type Meta struct {
	key string
	val interface{}
}

func NMeta(key string, val interface{}) Meta {
	return Meta{key, val}
}

type WorkContext struct {
	ActorName    string
	BusinassName string
	NodeIdx      int
	// RpcCompleteTopic string
	Concurrency int

	Height uint64
}

type CallFrame struct {
	ReqID  string
	Parent *CallFrame

	ReqMsg  *scommon.Message
	ReplyTo string

	Once sync.Once
}

type ExecutionContext struct {
	// ==== Message-related ====
	WorkCtx  *WorkContext
	traceCtx *logger.TraceContext

	Current *CallFrame

	// ==== Capability Proxy ====
	msgSender *broker.StatefulStreamer
	rpcCaller *rpc.RPCClient
	actor     *Actor

	// QueryExec *query.QueryExecution
	vals map[string]interface{}
}

func NewExecutionContext(ActorName string, broker *broker.StatefulStreamer, concurrency int) *ExecutionContext {
	return &ExecutionContext{
		msgSender: broker,
		rpcCaller: rpc.GlobalRPCClient,
		WorkCtx: &WorkContext{
			ActorName: ActorName,
			// RpcCompleteTopic: rpcCompleteTopic,
			Concurrency: concurrency,
		},
		traceCtx: &logger.TraceContext{},
		vals:     make(map[string]interface{}),
	}
}

func (ctx *ExecutionContext) Fork() *ExecutionContext {
	child := &ExecutionContext{
		Current: ctx.Current,
		WorkCtx: ctx.WorkCtx,
		traceCtx: &logger.TraceContext{
			TraceID:  ctx.traceCtx.TraceID,
			ParentID: ctx.traceCtx.SpanID,
			SpanID:   uuid.NewString(),
			SpanName: ctx.traceCtx.SpanName,
		},
		msgSender: ctx.msgSender,
		rpcCaller: ctx.rpcCaller,
		actor:     ctx.actor,
		// vals:      make(map[string]interface{}),
	}
	return child
}

func (ctx *ExecutionContext) WithBusinessName(businessName string) *ExecutionContext {
	ctx.WorkCtx.BusinassName = businessName
	return ctx
}
func (ctx *ExecutionContext) Concurrency() int {
	return ctx.WorkCtx.Concurrency
}

func (ctx *ExecutionContext) InvokeRPC(
	target string,
	method string,
	payload any,
	nextStep string,
	metas ...Meta,
) {
	allServerName := "rpc." + target
	msg := scommon.NewMessageForRPCREQ(allServerName, method, payload)
	msg.NextStep = fmt.Sprintf("%v.%v", ctx.WorkCtx.NodeIdx, nextStep)
	msg = ctx.AddTrace(msg)

	var cf *CallFrame
	if ctx.Current == nil {
		// root RPC
		cf = &CallFrame{ReqID: msg.ID, ReqMsg: msg}
	} else {
		// sub RPC
		cf = &CallFrame{ReqID: msg.ID, ReqMsg: msg, Parent: ctx.Current}
	}

	// ctx.Current = cf

	msg.WithRpcTrace(cf.ReqID)

	for i := range metas {
		switch metas[i].key {
		case "contId":
			msg.ContId = metas[i].val.(string)
		}
	}

	inv := &brokerpk.RPCInvocation{
		ID:         msg.ID,
		ServerName: allServerName,
		Method:     method,
		Payload:    msg,
		ReplyTo:    ctx.actor.rpcCompleteTopic,
	}

	ctx.actor.RegisterContinuation(msg.ID, NewBusinessContinuation(ctx.actor, cf))

	ctx.msgSender.Send(allServerName, inv)

	ctxlog := scommon.BindLoggerContextFromMessageSafe(context.Background(), msg)
	logger.Log.Debug(ctxlog, ctx.WorkCtx.BusinassName, "Send rpc Msg")

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
	return msg
}

func (ctx *ExecutionContext) SendRpcResponse(err string, data interface{}, args ...uint64) {
	frame := ctx.Current
	if frame == nil {
		panic("SendRpcResponse without CallFrame")
	}

	frame.Once.Do(func() {
		msg := scommon.NewMessageForStream(ctx.actor.rpcCompleteTopic, data)
		msg = ctx.AddTrace(msg)
		msg.WithRpcTrace(frame.ReqID)
		msg.Error = err

		ret := &brokerpk.RPCCompletion{
			ID:      frame.ReqID,
			Payload: msg,
			Error:   err,
		}

		ctx.msgSender.Send(ctx.actor.rpcCompleteTopic, ret)

		ctxlog := scommon.BindLoggerContextFromMessageSafe(context.Background(), msg)
		logger.Log.Debug(ctxlog, ctx.WorkCtx.BusinassName, "Business send rpc response")

		//for next send
		ctx.traceCtx.ParentID = msg.SpanID

		// if frame.Parent == nil {
		ctx.actor.removeFrame(frame.ReqID)
		// }
	})
}

func (ctx *ExecutionContext) Send(name string, data interface{}, args ...uint64) {
	if len(args) > 0 {
		ctx.WorkCtx.Height = args[0]
	}

	msg := scommon.NewMessageForStream(name, data)
	msg = ctx.AddTrace(msg)

	if ctx.Current != nil {
		msg.WithRpcTrace(ctx.Current.ReqID)
	}

	ctx.msgSender.Send(name, msg)

	ctxlog := scommon.BindLoggerContextFromMessageSafe(context.Background(), msg)
	logger.Log.Debug(ctxlog, ctx.WorkCtx.BusinassName, "Send Msg")

	//for next send
	ctx.traceCtx.ParentID = msg.SpanID
}

func (ctx *ExecutionContext) GetCtx() context.Context {
	ctx.traceCtx.SpanID = uuid.NewString()
	return logger.WithTrace(context.Background(), *ctx.traceCtx)
}

// -------log
func (ctx *ExecutionContext) LogDebug(event string, fs ...logger.Field) {
	logger.Log.Debug(ctx.GetCtx(), ctx.WorkCtx.BusinassName, event, fs...)
}

func (ctx *ExecutionContext) LogInfo(event string, fs ...logger.Field) {
	logger.Log.Info(ctx.GetCtx(), ctx.WorkCtx.BusinassName, event, fs...)
}

func (ctx *ExecutionContext) LogWarn(event string, fs ...logger.Field) {
	logger.Log.Warn(ctx.GetCtx(), ctx.WorkCtx.BusinassName, event, fs...)
}
func (ctx *ExecutionContext) LogErr(event string, fs ...logger.Field) {
	logger.Log.Error(ctx.GetCtx(), ctx.WorkCtx.BusinassName, event, fs...)
}
func (ctx *ExecutionContext) LogFatal(event string, fs ...logger.Field) {
	logger.Log.Fatal(ctx.GetCtx(), ctx.WorkCtx.BusinassName, event, fs...)
}
