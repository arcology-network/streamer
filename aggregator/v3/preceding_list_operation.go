package aggregator

import (
	types "github.com/arcology-network/common-lib/types"
	"github.com/arcology-network/component-lib/actor"
	evmCommon "github.com/arcology-network/evm/common"
)

type PrecedingListOperation struct{}

func (op *PrecedingListOperation) GetData(msg *actor.Message) (hashes []evmCommon.Hash, data []interface{}) {
	results := msg.Data.(*types.Euresults)
	if results == nil || len(*results) == 0 {
		return
	}

	for _, result := range *results {
		hashes = append(hashes, evmCommon.BytesToHash([]byte(result.H)))
		data = append(data, result)
	}
	return
}

func (op *PrecedingListOperation) GetList(msg *actor.Message) (hashes []evmCommon.Hash) {
	precedings := msg.Data.(*[]*evmCommon.Hash)
	for _, hash := range *precedings {
		hashes = append(hashes, *hash)
	}
	return
}

func (op *PrecedingListOperation) OnListFulfilled(data []interface{}, broker *actor.MessageWrapper) {
	broker.Send(actor.MsgPrecedingsEuresult, data)
}

func (op *PrecedingListOperation) Outputs() map[string]int {
	return map[string]int{actor.MsgPrecedingsEuresult: 1}
}

func (op *PrecedingListOperation) Config(params map[string]interface{}) {}
