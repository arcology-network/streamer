package aggregator

import (
	"github.com/arcology-network/component-lib/actor"
)

func init() {
	actor.Factory.Register("stateless_euresult_aggr_selector", func(concurrency int, groupId string) actor.IWorkerEx {
		return NewAggrSelector(
			concurrency,
			groupId,
			actor.MsgEuResults,
			actor.MsgPrecedingList,
			actor.MsgBlockCompleted,
			&PrecedingListOperation{},
		)
	})
	actor.Factory.Register("stateful_euresult_aggr_selector", func(concurrency int, groupId string) actor.IWorkerEx {
		return NewStatefulAggrSelector(
			concurrency,
			groupId,
			actor.MsgEuResults,
			actor.MsgInclusive,
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
