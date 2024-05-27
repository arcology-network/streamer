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
	"github.com/arcology-network/common-lib/types"
	"github.com/arcology-network/streamer/actor"
	evmCommon "github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
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

func (op *ReceiptOperation) GetList(msg *actor.Message) []evmCommon.Hash {
	// list := msg.Data.(*types.InclusiveList).HashList
	// for _, hash := range list {
	// 	hashes = append(hashes, *hash)
	// }
	// return
	return msg.Data.(*types.InclusiveList).HashList
}

func (op *ReceiptOperation) OnListFulfilled(data []interface{}, broker *actor.MessageWrapper) {
	broker.Send(actor.MsgSelectedReceipts, data)
}

func (op *ReceiptOperation) Outputs() map[string]int {
	return map[string]int{actor.MsgSelectedReceipts: 1}
}

func (op *ReceiptOperation) Config(params map[string]interface{}) {}
