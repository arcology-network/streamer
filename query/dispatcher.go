package query

import "github.com/arcology-network/streamer/actor"

type Dispatcher struct {
	plans map[string]QueryPlan
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		plans: make(map[string]QueryPlan),
	}
}

func (d *Dispatcher) Register(name string, plan QueryPlan) {
	d.plans[name] = plan
}

func (d *Dispatcher) GetQueryPlan(name string) QueryPlan {
	return d.plans[name]
}
func (d *Dispatcher) Query(
	name string,
	req interface{},
	execCtx *actor.ExecutionContext,
	obs StepObserver,
	scheduler Scheduler,
	cont Continuation,
) {
	plan := d.plans[name]
	if plan == nil {
		cont(nil, ErrPlanNotFound)
		return
	}

	ctx := NewQueryContext(req, execCtx, obs, scheduler, func(resp interface{}, err error) {
		cont(resp, err)
	})

	plan.Start(ctx, func(resp interface{}, err error) {
		if ctx.finished {
			return
		}
		ctx.Return(resp, err)
	})
}
