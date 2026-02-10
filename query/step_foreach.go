package query

type ForEachStep struct {
	Items   func(ctx *QueryContext) []interface{}
	Body    func(item interface{}) Step
	Collect func(parent *QueryContext, item any, sub *QueryContext)
}

func (s *ForEachStep) Start(
	ctx *QueryContext,
	cont Continuation,
) {
	items := s.Items(ctx)
	if len(items) == 0 {
		cont(nil, nil)
		return
	}

	var run func(i int)
	run = func(i int) {
		if i >= len(items) {
			cont(nil, nil)
			return
		}

		item := items[i]

		subCtx := ctx.Fork()

		subCtx.OnReturn(func(v any, err error) {
			if err != nil {
				cont(nil, err)
				return
			}

			if s.Collect != nil {
				s.Collect(ctx, item, subCtx)
			}

			run(i + 1)
		})

		step := s.Body(item)

		StartStep(subCtx, step, func(_ any, err error) {
			if err != nil {
				cont(nil, err)
				return
			}

			if subCtx.IsReturned() {
				return
			}

			if s.Collect != nil {
				s.Collect(ctx, item, subCtx)
			}

			run(i + 1)
		})
	}
	run(0)
}
