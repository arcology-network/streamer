package aggregator

import (
	"github.com/arcology-network/common-lib/types"
	"github.com/arcology-network/component-lib/actor"
	evmCommon "github.com/arcology-network/evm/common"
	ethTypes "github.com/arcology-network/evm/core/types"
)

type ReceiptOperation struct{}

func (op *ReceiptOperation) GetData(msg *actor.Message) (hashes []evmCommon.Hash, data []interface{}) {
	receipts := msg.Data.(*[]*ethTypes.Receipt)
	if receipts == nil {
		return
	}

	for _, receipt := range *receipts {
		hashes = append(hashes, receipt.TxHash)
		data = append(data, receipt)
	}
	return
}

func (op *ReceiptOperation) GetList(msg *actor.Message) (hashes []evmCommon.Hash) {
	list := msg.Data.(*types.InclusiveList).HashList
	for _, hash := range list {
		hashes = append(hashes, *hash)
	}
	return
}

func (op *ReceiptOperation) OnListFulfilled(data []interface{}, broker *actor.MessageWrapper) {
	broker.Send(actor.MsgSelectedReceipts, data)
}

func (op *ReceiptOperation) Outputs() map[string]int {
	return map[string]int{actor.MsgSelectedReceipts: 1}
}

func (op *ReceiptOperation) Config(params map[string]interface{}) {}
