package actor

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"reflect"
	"runtime"
	"sort"
	"strings"

	scommon "github.com/arcology-network/streamer/common"
	"github.com/arcology-network/streamer/logger"
	"github.com/google/uuid"
)

type Action func(ctx *ActionContext) error

type ActionContext struct {
	Messages []*scommon.Message // disjunction: len=1, conjunction: len>1
	RPC      *RPCContext        //  RPC

	ExecCtx *ExecutionContext
}

type ActionRegistrar interface {
	Register(actionName string, action Action)
}

type ActionRouter struct {
	actions      map[string]Action
	bussnessname string
	nodeIdx      int
}

func NewActionRouter(name string, idx int) *ActionRouter {
	return &ActionRouter{
		actions:      make(map[string]Action),
		bussnessname: name,
		nodeIdx:      idx,
	}
}

func (r *ActionRouter) Register(actionName string, action Action) {
	r.actions[actionName] = action
}

func (r *ActionRouter) HasAction(actionName string) bool {
	_, ok := r.actions[actionName]
	return ok
}

func (r *ActionRouter) Dispatch(
	actionName string,
	msgs []*scommon.Message,
	rpc *RPCContext,
	execctx *ExecutionContext,
) error {
	action, ok := r.actions[actionName]
	if !ok {
		return nil //
	}

	execctx = r.UpdateCtx(msgs, execctx, action)
	r.log(msgs)
	ctx := &ActionContext{
		ExecCtx:  execctx,
		Messages: msgs,
		RPC:      rpc,
	}

	return action(ctx)
}

func (r *ActionRouter) log(msgs []*scommon.Message) {
	for _, msg := range msgs {
		ctx := scommon.BindLoggerContextFromMessageSafe(context.Background(), msg)
		logger.Log.Debug(ctx, "Start apply Msg")
	}
}

func (r *ActionRouter) UpdateCtx(msgs []*scommon.Message, ctx *ExecutionContext, action Action) *ExecutionContext {
	ctx.traceCtx.SpanName = MethodName(action)
	ctx.WorkCtx.BusinassName = r.bussnessname

	maxHeight := uint64(0)
	foundIdx := 0
	for i := range msgs {
		if msgs[i].Height > maxHeight {
			foundIdx = i
			maxHeight = msgs[i].Height
		}
	}

	ctx.traceCtx.TraceID = msgs[foundIdx].TraceID
	ctx.traceCtx.ParentID = NewAggregationParent(msgs)
	ctx.WorkCtx.Height = msgs[foundIdx].Height
	ctx.traceCtx.ReqID = msgs[foundIdx].ReqID
	ctx.WorkCtx.NodeIdx = r.nodeIdx
	return ctx
}

func MethodName(fn interface{}) string {
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		return "<not-func>"
	}

	full := runtime.FuncForPC(v.Pointer()).Name()

	if i := strings.LastIndex(full, "/"); i >= 0 {
		full = full[i+1:]
	}
	tmpfull := full

	if i := strings.LastIndex(full, "."); i >= 0 {
		tmpfull = full[i+1:]
	}

	if strings.HasSuffix(tmpfull, "-fm") {
		return tmpfull[:len(tmpfull)-3]
	}
	return tmpfull
}

func NewAggregationParent(msgs []*scommon.Message) string {
	if len(msgs) == 0 {
		return ""
	}

	spanIDs := make([]string, 0, len(msgs))

	if len(msgs) == 1 {
		return msgs[0].SpanID
	}

	for _, m := range msgs {
		if m == nil {
			continue
		}
		if m.SpanID != "" {
			spanIDs = append(spanIDs, m.SpanID)
		}
	}

	sort.Strings(spanIDs)

	h := sha256.New()
	for _, s := range spanIDs {
		h.Write([]byte(s))
	}
	hash := hex.EncodeToString(h.Sum(nil))[:16]

	return "agg-" + hash + "-" + uuid.NewString()[:8]

}
