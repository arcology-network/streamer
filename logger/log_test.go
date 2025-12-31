package logger

import (
	"context"
	"testing"
)

func TestLogger(t *testing.T) {
	log := Log

	ctx := context.Background()

	ctx = WithTrace(ctx, TraceContext{
		TraceID:  "trace-1",
		SpanID:   "span-2",
		ParentID: "span-1",
	})

	ctx = WithMsg(ctx, MsgContext{
		MsgType:   "STREAM",
		Name:      "order.created",
		Key:       "order:8899",
		Sequence:  1024,
		Redeliver: 1,
	})

	log.Debug(ctx, "order_failed",
		F("order_id", 8899),
		F("cost_ms", 32),
		F("error", "out of stock"),
	)

	log.Info(ctx, "order_failed",
		F("order_id", 8899),
		F("cost_ms", 32),
		F("error", "out of stock"),
	)

	log.Warn(ctx, "order_failed",
		F("order_id", 8899),
		F("cost_ms", 32),
		F("error", "out of stock"),
	)

	log.Error(ctx, "order_failed",
		F("order_id", 8899),
		F("cost_ms", 32),
		F("error", "out of stock"),
	)

	// ✅✅✅ Key: Ensure logs are 100% written to both stdout and files
	log.Sync()
}
