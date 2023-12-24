package aggregator

import (
	"fmt"

	"github.com/arcology-network/component-lib/actor"
	evmCommon "github.com/arcology-network/evm/common"
)

type AggrOperation interface {
	GetData(msg *actor.Message) ([]evmCommon.Hash, []interface{})
	GetList(msg *actor.Message) []evmCommon.Hash
	OnListFulfilled(data []interface{}, broker *actor.MessageWrapper)
	Outputs() map[string]int
	Config(params map[string]interface{})
}

type AggrSelector struct {
	actor.WorkerThread

	dataMsg  string
	listMsg  string
	clearMsg string
	ds       *DataSet
	op       AggrOperation
}

func NewAggrSelector(concurrency int, groupId string, dataMsg string, listMsg string, clearMsg string, op AggrOperation) *AggrSelector {
	aggr := &AggrSelector{
		dataMsg:  dataMsg,
		listMsg:  listMsg,
		clearMsg: clearMsg,
		ds:       NewDataSet(),
		op:       op,
	}
	aggr.Set(concurrency, groupId)
	return aggr
}

func (aggr *AggrSelector) Inputs() ([]string, bool) {
	return []string{aggr.dataMsg, aggr.listMsg, aggr.clearMsg}, false
}

func (aggr *AggrSelector) Outputs() map[string]int {
	return aggr.op.Outputs()
}

func (aggr *AggrSelector) OnStart() {}

func (aggr *AggrSelector) OnMessageArrived(msgs []*actor.Message) error {
	if len(msgs) > 1 {
		panic("too many messages received")
	}

	msg := msgs[0]
	switch msg.Name {
	case aggr.dataMsg:
		hashes, data := aggr.op.GetData(msg)
		for i, hash := range hashes {
			lists := aggr.ds.Add(hash, data[i], msg.Height)
			for _, list := range lists {
				aggr.op.OnListFulfilled(list, aggr.MsgBroker)
			}
		}
	case aggr.listMsg:
		list := aggr.op.GetList(msg)
		data := aggr.ds.Get(list, msg.Height)
		if data != nil {
			aggr.op.OnListFulfilled(data, aggr.MsgBroker)
		}
	case aggr.clearMsg:
		aggr.ds.Clear(msg.Height)
	default:
		panic(fmt.Sprintf("unexpected message type: %v", msg.Name))
	}
	return nil
}
