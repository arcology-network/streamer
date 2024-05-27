/*
 *   Copyright (c) 2024 Arcology Network

 *   This program is free software: you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation, either version 3 of the License, or
 *   (at your option) any later version.

 *   This program is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *   GNU General Public License for more details.

 *   You should have received a copy of the GNU General Public License
 *   along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

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
