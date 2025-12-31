package actor

import (
	"fmt"

	scommon "github.com/arcology-network/streamer/common"
)

type Controller interface {
	// Process a single message, return nil to indicate caching, non-nil to indicate release
	OnMessage(msg *scommon.Message) ([]*scommon.Message, error)
	// Triggered after business execution completes, release the cache
	OnAfterExecute() ([]*scommon.Message, error)
}

type ControlledNode struct {
	business      Business
	controllers   []Controller
	router        *ActionRouter
	name          string
	isConjunction bool
}

func NewControlledNode(
	business Business,
	controllers []Controller,
	businessname string,
	nodexIdx int,
) *ControlledNode {
	_, iscon := business.Inputs()
	n := &ControlledNode{
		business:      business,
		controllers:   controllers,
		isConjunction: iscon,
		name:          businessname,
	}
	n.router = NewActionRouter(n.name, nodexIdx)
	business.RegisterActions(n.router)
	//check
	err := n.Validate()
	if err != nil {
		panic(err)
	}
	//
	return n
}

// Handle receives the message and executes the closed loop
func (n *ControlledNode) Handle(
	msgs []*scommon.Message,
	execCtx *ExecutionContext,
) error {

	pending := msgs
	round := 0

	for len(pending) > 0 {
		round++
		if round > 100 {
			return fmt.Errorf("node %s exceeded max rounds", n.name)
		}

		ready, err := n.control(pending)
		if err != nil {
			return err
		}

		if len(ready) > 0 {
			if err := n.execute(ready, execCtx); err != nil {
				return err
			}
		}

		pending, err = n.AfterExecute()
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *ControlledNode) control(msgs []*scommon.Message) ([]*scommon.Message, error) {
	ready := msgs

	// The control chain processes sequentially
	for i := 0; i < len(n.controllers); i++ {
		ctrl := n.controllers[i]
		out := []*scommon.Message{}
		for _, msg := range ready {
			res, err := ctrl.OnMessage(msg)
			if err != nil {
				return nil, err
			}
			if res != nil {
				out = append(out, res...)
			}
		}
		ready = out
		if len(ready) == 0 {
			// The control chain intercepted all messages
			return nil, nil
		}
	}
	return ready, nil
}
func (n *ControlledNode) AfterExecute() ([]*scommon.Message, error) {
	if len(n.controllers) == 0 {
		return nil, nil
	}

	// Disjunction: The control chain releases the message
	var released []*scommon.Message
	for i := len(n.controllers) - 1; i >= 0; i-- {
		ctrl := n.controllers[i]
		out, err := ctrl.OnAfterExecute()
		if err != nil {
			return nil, err
		}
		if out != nil {
			released = append(released, out...)
		}
	}

	return released, nil
}

// execute Invoke the router to execute the business
func (n *ControlledNode) execute(msgs []*scommon.Message, execCtx *ExecutionContext) error {
	if len(msgs) == 0 {
		return nil
	}
	if n.isConjunction {
		return n.router.Dispatch(msgs[0].Name, msgs, nil, execCtx)
	}
	for i := range msgs {
		err := n.router.Dispatch(msgs[i].Name, []*scommon.Message{msgs[i]}, nil, execCtx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *ControlledNode) StartRPC(
	method string,
	msgs []*scommon.Message,
	rpcCtx *RPCContext,
	execCtx *ExecutionContext,
) error {
	// not controller
	return n.router.Dispatch(
		method,
		msgs,
		rpcCtx,
		execCtx,
	)
}

func (n *ControlledNode) Validate() error {
	inputs, _ := n.business.Inputs()
	for _, in := range inputs {
		if !n.router.HasAction(in) {
			return fmt.Errorf(
				"business %s subscribes %s but no action registered",
				n.name, in,
			)
		}
	}
	return nil
}
