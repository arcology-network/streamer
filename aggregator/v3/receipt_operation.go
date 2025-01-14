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
	"fmt"

	"github.com/arcology-network/common-lib/types"
	"github.com/arcology-network/streamer/actor"
	evmCommon "github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	EventSigner = evmCommon.BytesToHash(crypto.Keccak256([]byte("WriteBackEvent(bytes32,bytes32,bytes)")))
)

func PrintReceipts(receipts []*ethTypes.Receipt) {
	for i := range receipts {
		fmt.Printf(">>>>>>>>>>>>>>>>>>>>>receipts.txhash:%x,logs:%v\n", receipts[i].TxHash.Bytes(), len(receipts[i].Logs))
		for j := range receipts[i].Logs {
			fmt.Printf("*************************receipts.logs.blockNumber:%v\n", receipts[i].Logs[j].BlockNumber)
			fmt.Printf("-------------------------receipts.logs.data:%x\n", receipts[i].Logs[j].Data)
			fmt.Printf("-------------------------receipts.logs.txhash:%x\n", receipts[i].Logs[j].TxHash.Bytes())
			fmt.Printf("-------------------------receipts.logs.TxIndex:%v\n", receipts[i].Logs[j].TxIndex)
			fmt.Printf("-------------------------receipts.logs.Index:%v\n", receipts[i].Logs[j].Index)
			for k := range receipts[i].Logs[j].Topics {
				fmt.Printf("-------------------------receipts.logs.topics:%x\n", receipts[i].Logs[j].Topics[k].Bytes())
			}
		}
	}
}

func PostProcess(receipts []*ethTypes.Receipt) []*ethTypes.Receipt {
	logs := make([]*ethTypes.Log, 0, 50000)
	mpReceipts := make(map[evmCommon.Hash]*ethTypes.Receipt, len(receipts))

	//filter defer logs
	for i := range receipts {
		tmpLogs := make([]*ethTypes.Log, 0, len(receipts[i].Logs))
		for j := range receipts[i].Logs {
			if receipts[i].Logs[j].Topics[0] == EventSigner {
				//collect defer logs
				logs = append(logs, receipts[i].Logs[j])
			} else {
				//remove current log from src
				receipts[i].Logs[j].Index = uint(len(tmpLogs))
				tmpLogs = append(tmpLogs, receipts[i].Logs[j])
			}
		}
		receipts[i].Logs = tmpLogs
		mpReceipts[receipts[i].TxHash] = receipts[i]
	}

	if len(logs) == 0 {
		return receipts
	}

	//update ldefer log into parallel transactions
	for i := range logs {
		txhash := logs[i].Topics[1]
		destReceipt := mpReceipts[txhash]

		topics, data := ParseData(logs[i].Data[64:])

		logs[i].TxHash = txhash
		logs[i].TxIndex = destReceipt.TransactionIndex
		logs[i].Index = uint(len(destReceipt.Logs))
		logs[i].Topics = AppendTopic(topics, logs[i].Topics[2]) //[]evmCommon.Hash{logs[i].Topics[2]}
		logs[i].Data = data
		destReceipt.Logs = append(destReceipt.Logs, logs[i])
		mpReceipts[txhash] = destReceipt
	}

	//setup dest receipts
	destReceipts := make([]*ethTypes.Receipt, 0, len(mpReceipts))
	for i := range receipts {
		destReceipts = append(destReceipts, mpReceipts[receipts[i].TxHash])
	}
	return destReceipts
}
func AppendTopic(topics []evmCommon.Hash, firstTopic evmCommon.Hash) []evmCommon.Hash {
	finalTopics := make([]evmCommon.Hash, 0, 3)
	finalTopics = append(finalTopics, firstTopic)
	for i := range topics {
		finalTopics = append(finalTopics, topics[i])
	}
	if len(finalTopics) > 3 {
		return finalTopics[:3]
	} else {
		return finalTopics
	}
}
func ParseData(datas []byte) ([]evmCommon.Hash, []byte) {
	Indexed_Splite := "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	findIdx := -1
	topics := make([]evmCommon.Hash, 0, 3)
	for i := 0; i < len(datas)/32; i++ {
		if fmt.Sprintf("%x", datas[i*32:(i+1)*32]) == Indexed_Splite {
			findIdx = i
		}
	}
	if findIdx < 0 {
		return topics, datas
	}
	for i := 0; i < len(datas)/32; i++ {
		if i < findIdx {
			topics = append(topics, evmCommon.BytesToHash(datas[i*32:(i+1)*32]))
		} else if i > findIdx {
			datas = datas[i*32:]
			break
		}
	}
	return topics, datas
}

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
	var receipts []*ethTypes.Receipt
	for _, item := range data {
		receipts = append(receipts, item.(*ethTypes.Receipt))
	}
	broker.Send(actor.MsgSelectedReceipts, PostProcess(receipts))
	// broker.Send(actor.MsgSelectedReceipts, receipts)
}

func (op *ReceiptOperation) Outputs() map[string]int {
	return map[string]int{actor.MsgSelectedReceipts: 1}
}

func (op *ReceiptOperation) Config(params map[string]interface{}) {}
