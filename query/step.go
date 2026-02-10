package query

import "github.com/arcology-network/streamer/actor"

type QueryContinuation interface {
	QueryContinuationAction(ctx *actor.ActionContext) error
}

type Step interface {
	Start(ctx *QueryContext, cont Continuation)
}

func StartStep(ctx *QueryContext, step Step, cont Continuation) {
	if ctx.finished {
		return
	}

	if ctx.Observer != nil {
		ctx.Observer.OnStepStart(ctx, step)
	}

	step.Start(ctx, func(resp interface{}, err error) {
		if ctx.finished {
			return
		}

		if ctx.Observer != nil {
			ctx.Observer.OnStepFinish(ctx, step, err)
		}
		cont(resp, err)
	})
}
