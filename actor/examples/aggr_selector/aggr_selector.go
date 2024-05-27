package aggrselector

import (
	"fmt"

	"github.com/arcology-network/streamer/actor"
)

const (
	stateInit = iota
	stateCollecting
	stateDone
)

const (
	msgData  = "data"
	msgList  = "list"
	msgClear = "clear"
)

type AggrSelector struct {
	actor.WorkerThread

	state    int
	height   uint64
	buf      []string
	waitings map[string]struct{}
}

func NewAggrSelector() *AggrSelector {
	return &AggrSelector{
		state: stateInit,
	}
}

func (aggr *AggrSelector) Inputs() ([]string, bool) {
	return []string{msgData, msgList, msgClear}, false
}

func (aggr *AggrSelector) Outputs() map[string]int {
	return map[string]int{}
}

func (aggr *AggrSelector) OnStart() {

}

func (aggr *AggrSelector) OnMessageArrived(msgs []*actor.Message) error {
	msg := msgs[0]
	switch aggr.state {
	case stateInit:
		if msg.Name == msgData {
			aggr.buf = append(aggr.buf, msg.Data.(string))
		} else { // msg.Name == msgList
			list := msg.Data.([]string)
			aggr.waitings = make(map[string]struct{})
			for _, elem := range list {
				aggr.waitings[elem] = struct{}{}
			}

			for _, elem := range aggr.buf {
				delete(aggr.waitings, elem)
			}

			if len(aggr.waitings) == 0 {
				aggr.state = stateDone
				fmt.Println("done.")
			} else {
				aggr.state = stateCollecting
				fmt.Println("collecting.")
			}
		}
	case stateCollecting:
		delete(aggr.waitings, msg.Data.(string))
		if len(aggr.waitings) == 0 {
			aggr.state = stateDone
			fmt.Println("done.")
		}
	case stateDone:
		if msg.Name == msgClear {
			aggr.state = stateInit
			aggr.height = msg.Height + 1
			fmt.Println("cleared.")
		}
	}
	return nil
}

func (aggr *AggrSelector) GetStateDefinitions() map[int][]string {
	return map[int][]string{
		stateInit:       {msgData, msgList},
		stateCollecting: {msgData},
		stateDone:       {msgData, msgClear},
	}
}

func (aggr *AggrSelector) GetCurrentState() int {
	return aggr.state
}

func (aggr *AggrSelector) Height() uint64 {
	return aggr.height
}
