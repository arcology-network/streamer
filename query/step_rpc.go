package query

type RpcStep struct {
	Call func(ctx *QueryContext, cont Continuation)
}

func (s *RpcStep) Start(ctx *QueryContext, cont Continuation) {
	if ctx.Observer != nil {
		ctx.Observer.OnRPCSend(ctx, s, nil)
	}

	s.Call(ctx, func(resp interface{}, err error) {
		if ctx.Observer != nil {
			ctx.Observer.OnRPCRecv(ctx, s, resp, err)
		}

		cont(resp, err)
	})
}
