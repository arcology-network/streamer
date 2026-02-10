package query

type ResultStep struct {
	Build func(ctx *QueryContext) interface{}
}

func (s *ResultStep) Start(ctx *QueryContext, cont Continuation) {
	cont(s.Build(ctx), nil)
}

type ReturnFromSubPlanStep struct {
	Cond  func(ctx *QueryContext) bool
	Value func(ctx *QueryContext) any
	Err   error
}

func (s *ReturnFromSubPlanStep) Start(
	ctx *QueryContext,
	cont Continuation,
) {
	if s.Cond != nil && !s.Cond(ctx) {
		cont(nil, nil)
		return
	}

	var v any
	if s.Value != nil {
		v = s.Value(ctx)
	}

	ctx.Return(v, s.Err)
}

type CallStep struct {
	Plan Step
	Bind func(ctx *QueryContext, v any)
}

func (s *CallStep) Start(
	ctx *QueryContext,
	cont Continuation,
) {
	subCtx := ctx.Fork()

	subCtx.OnReturn(func(v any, err error) {
		if err != nil {
			cont(nil, err)
			return
		}

		if s.Bind != nil {
			s.Bind(ctx, v)
		}

		cont(nil, nil)
	})

	StartStep(subCtx, s.Plan, func(v any, err error) {
		if err != nil {
			cont(nil, err)
			return
		}

		if subCtx.IsReturned() {
			return
		}

		if s.Bind != nil {
			s.Bind(ctx, v)
		}

		cont(nil, nil)
	})
}
