package logger

import (
	"context"
)

type ctxKey string

type TraceContext struct {
	TraceID  string
	SpanID   string
	ParentID string
	SpanName string
	// ReqID    string
}

type MsgContext struct {
	MsgType   string
	Name      string
	Key       string
	Sequence  uint64
	Redeliver int
	Height    uint64
	From      string
	Method    string
	ReqID     string
}

const (
	traceKey ctxKey = "trace"
	msgKey   ctxKey = "msg"
)

func WithTrace(ctx context.Context, t TraceContext) context.Context {
	return context.WithValue(ctx, traceKey, t)
}

func WithMsg(ctx context.Context, m MsgContext) context.Context {
	return context.WithValue(ctx, msgKey, m)
}

func GetTrace(ctx context.Context) (TraceContext, bool) {
	v := ctx.Value(traceKey)
	if v == nil {
		return TraceContext{}, false
	}
	t, ok := v.(TraceContext)
	return t, ok
}

func GetMsg(ctx context.Context) (MsgContext, bool) {
	v := ctx.Value(msgKey)
	if v == nil {
		return MsgContext{}, false
	}
	m, ok := v.(MsgContext)
	return m, ok
}
