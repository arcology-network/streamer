package actor

import (
	"context"

	scommon "github.com/arcology-network/streamer/common"
	"github.com/arcology-network/streamer/logger"
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

	businessName string
}

func NewFSMController(client FSMCompatible, businessName string) *FSMController {
	ctrl := &FSMController{
		client:       client,
		rules:        client.GetFSMRules(),
		buf:          NewMsgBuffer(),
		businessName: businessName,
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

	logger.Log.Debug(context.Background(), c.businessName, "FSM Push", logger.F("msg", msg.Name), logger.F("state", state))
	return nil, nil
}
func (c *FSMController) OnAfterExecute() ([]*scommon.Message, error) {
	state := c.client.GetCurrentState()
	rule := c.rules[state]

	logger.Log.Debug(context.Background(), c.businessName, "FSM Status", logger.F("state", state), logger.F("accept", rule))

	if msg := c.buf.PopByNames(rule.Accept); msg != nil {

		logger.Log.Debug(context.Background(), c.businessName, "FSM Pop", logger.F("msg", msg.Name), logger.F("state", state))
		return []*scommon.Message{msg}, nil
	}
	return nil, nil
}
