package query

type SequentialStep struct {
	Steps []Step
}

func (s *SequentialStep) Start(ctx *QueryContext, cont Continuation) {
	if ctx.IsReturned() {
		return
	}

	if len(s.Steps) == 0 {
		cont(nil, nil)
		return
	}

	var run func(i int)
	run = func(i int) {
		if ctx.IsReturned() {
			return
		}

		if i >= len(s.Steps) {
			cont(nil, nil)
			return
		}

		StartStep(ctx, s.Steps[i], func(resp any, err error) {
			if ctx.IsReturned() {
				return
			}

			if err != nil {
				cont(nil, err)
				return
			}

			if i == len(s.Steps)-1 {
				cont(resp, nil)
				return
			}

			run(i + 1)
		})
	}

	run(0)
}
