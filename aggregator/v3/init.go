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
	scommon "github.com/arcology-network/streamer/common"
)

func init() {
	actor.Factory.Register("stateless_euresult_aggr_selector", func() actor.Business {
		return NewAggrSelector(
			"stateless_euresult_aggr_selector",
			scommon.MsgEuResults,
			scommon.MsgGenerationReapingList,
			scommon.MsgBlockEnd,
			&EuResultOperation{},
		)
	})
	actor.Factory.Register("stateful_euresult_aggr_selector", func() actor.Business {
		return NewStatefulAggrSelector(
			"stateful_euresult_aggr_selector",
			scommon.MsgEuResults,
			scommon.MsgGenerationReapingList,
			scommon.MsgBlockEnd,
			&EuResultOperation{},
		)
	})
	actor.Factory.Register("stateful_receipt_aggr_selector", func() actor.Business {
		return NewStatefulAggrSelector(
			"stateful_receipt_aggr_selector",
			scommon.MsgReceipts,
			scommon.MsgInclusive,
			scommon.MsgBlockCompleted,
			&ReceiptOperation{},
		)
	})
	actor.Factory.Register("stateful_receipt_hash_aggr_selector", func() actor.Business {
		return NewStatefulAggrSelector(
			"stateful_receipt_hash_aggr_selector",
			scommon.MsgReceiptHashList,
			scommon.MsgInclusive,
			scommon.MsgBlockCompleted,
			&ReceiptHashOperation{},
		)
	})
}
