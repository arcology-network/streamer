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
	"github.com/arcology-network/streamer/actor"
)

func init() {
	actor.Factory.Register("stateless_euresult_aggr_selector", func(concurrency int, groupId string) actor.IWorkerEx {
		return NewAggrSelector(
			concurrency,
			groupId,
			actor.MsgEuResults,
			actor.MsgGenerationReapingList,
			actor.MsgBlockEnd,
			&EuResultOperation{},
		)
	})
	actor.Factory.Register("stateful_euresult_aggr_selector", func(concurrency int, groupId string) actor.IWorkerEx {
		return NewStatefulAggrSelector(
			concurrency,
			groupId,
			actor.MsgEuResults,
			actor.MsgGenerationReapingList,
			actor.MsgBlockEnd,
			&EuResultOperation{},
		)
	})
	actor.Factory.Register("stateful_receipt_aggr_selector", func(concurrency int, groupId string) actor.IWorkerEx {
		return NewStatefulAggrSelector(
			concurrency,
			groupId,
			actor.MsgReceipts,
			actor.MsgInclusive,
			actor.MsgBlockCompleted,
			&ReceiptOperation{},
		)
	})
	actor.Factory.Register("stateful_receipt_hash_aggr_selector", func(concurrency int, groupId string) actor.IWorkerEx {
		return NewStatefulAggrSelector(
			concurrency,
			groupId,
			actor.MsgReceiptHashList,
			actor.MsgInclusive,
			actor.MsgBlockCompleted,
			&ReceiptHashOperation{},
		)
	})
}
