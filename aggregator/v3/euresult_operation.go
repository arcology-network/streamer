package aggregator

import (
	types "github.com/arcology-network/common-lib/types"
	eushared "github.com/arcology-network/eu/shared"
	"github.com/arcology-network/streamer/actor"
	evmCommon "github.com/ethereum/go-ethereum/common"
)

type EuResultOperation struct {
	output string
}

func (op *EuResultOperation) GetData(msg *actor.Message) (hashes []evmCommon.Hash, data []interface{}) {
	results := msg.Data.(*eushared.Euresults)
	if results == nil {
		return
	}

	for _, result := range *results {
		hashes = append(hashes, evmCommon.BytesToHash([]byte(result.H)))
		data = append(data, result)
	}
	return
}

func (op *EuResultOperation) GetList(msg *actor.Message) (hashes []evmCommon.Hash) {
	list := msg.Data.(*types.InclusiveList)
	for i, hash := range list.HashList {
		if list.Successful[i] {
			hashes = append(hashes, hash)
		}
	}
	return
}

func (op *EuResultOperation) OnListFulfilled(data []interface{}, broker *actor.MessageWrapper) {
	broker.Send(op.output, data)
}

func (op *EuResultOperation) Outputs() map[string]int {
	return map[string]int{op.output: 1}
}

func (op *EuResultOperation) Config(params map[string]interface{}) {
	op.output = params["output"].(string)
}
