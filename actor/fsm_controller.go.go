package actor

import (
	"fmt"

	scommon "github.com/arcology-network/streamer/common"
)

type FSMRule struct {
	Accept []string
}

type FSMCompatible interface {
	GetFSMRules() map[int]FSMRule
	GetCurrentState() int
}

type FSMController struct {
	client       FSMCompatible
	rules        map[int]FSMRule
	universalSet map[string]struct{}
	buf          *MsgBuffer
}

func NewFSMController(client FSMCompatible) *FSMController {
	ctrl := &FSMController{
		client: client,
		rules:  client.GetFSMRules(),
		buf:    NewMsgBuffer(),
	}

	ctrl.universalSet = make(map[string]struct{})
	for _, rule := range ctrl.rules {
		for _, name := range rule.Accept {
			ctrl.universalSet[name] = struct{}{}
		}
	}
	return ctrl
}

func (c *FSMController) OnMessage(msg *scommon.Message) ([]*scommon.Message, error) {
	if _, ok := c.universalSet[msg.Name]; !ok {
		return []*scommon.Message{msg}, nil
	}

	state := c.client.GetCurrentState()
	rule := c.rules[state]

	for _, accept := range rule.Accept {
		if msg.Name == accept {
			return []*scommon.Message{msg}, nil
		}
	}

	c.buf.Put(msg)
	fmt.Printf("FSM: Push %s\n", msg.Name)
	return nil, nil
}
func (c *FSMController) OnAfterExecute() ([]*scommon.Message, error) {
	state := c.client.GetCurrentState()
	rule := c.rules[state]

	if msg := c.buf.PopByNames(rule.Accept); msg != nil {
		fmt.Printf("FSM: Pop %s (state=%d)\n", msg.Name, state)
		return []*scommon.Message{msg}, nil
	}
	return nil, nil
}
