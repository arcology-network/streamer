package aggregator

import (
	types "github.com/arcology-network/common-lib/types"
	"github.com/arcology-network/streamer/actor"
	evmCommon "github.com/ethereum/go-ethereum/common"
)

type ReceiptHashOperation struct{}

func (op *ReceiptHashOperation) GetData(msg *actor.Message) ([]evmCommon.Hash, []interface{}) {
	receiptHashes := msg.Data.(*types.ReceiptHashList)

	if receiptHashes == nil {
		return []evmCommon.Hash{}, []interface{}{}
	}
	size := len(receiptHashes.TxHashList)
	hashes := make([]evmCommon.Hash, 0, size)
	data := make([]interface{}, 0, size)
	for i := range receiptHashes.TxHashList {
		hashes = append(hashes, receiptHashes.TxHashList[i])
		data = append(data, &types.ReceiptHash{
			Txhash:      &receiptHashes.TxHashList[i],
			Receipthash: &receiptHashes.ReceiptHashList[i],
			GasUsed:     receiptHashes.GasUsedList[i],
		})
	}
	return hashes, data
}

func (op *ReceiptHashOperation) GetList(msg *actor.Message) []evmCommon.Hash {
	// list := msg.Data.(*types.InclusiveList).HashList
	// hashes := make([]evmCommon.Hash, 0, len(list))
	// for i := range list {
	// 	hashes = append(hashes, list[i])
	// }
	// return hashes
	return msg.Data.(*types.InclusiveList).HashList
}

func (op *ReceiptHashOperation) OnListFulfilled(data []interface{}, broker *actor.MessageWrapper) {
	receipts := make(map[evmCommon.Hash]*types.ReceiptHash)
	for i := range data {
		rh := data[i].(*types.ReceiptHash)
		receipts[*rh.Txhash] = rh
	}
	broker.Send(actor.MsgSelectedReceiptsHash, &receipts)
}

func (op *ReceiptHashOperation) Outputs() map[string]int {
	return map[string]int{actor.MsgSelectedReceiptsHash: 1}
}

func (op *ReceiptHashOperation) Config(params map[string]interface{}) {}
