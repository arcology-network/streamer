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
	eushared "github.com/arcology-network/eu/shared"
	"github.com/arcology-network/streamer/actor"
	evmCommon "github.com/ethereum/go-ethereum/common"
)

type PrecedingListOperation struct{}

func (op *PrecedingListOperation) GetData(msg *actor.Message) (hashes []evmCommon.Hash, data []interface{}) {
	results := msg.Data.(*eushared.Euresults)
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
