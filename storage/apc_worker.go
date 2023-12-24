package storage

import (
	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/types"
	"github.com/arcology-network/concurrenturl/interfaces"
	univaluepk "github.com/arcology-network/concurrenturl/univalue"
)

func GetTransitionIds(euresults []*types.EuResult) []uint32 {
	txIds := make([]uint32, len(euresults))
	for i, euresult := range euresults {
		txIds[i] = euresult.ID
	}
	return txIds
}

func GetTransitions(euresults []*types.EuResult) ([]uint32, []interfaces.Univalue) {
	txIds := make([]uint32, len(euresults))
	transitionsize := 0
	for i, euresult := range euresults {
		transitionsize = transitionsize + len(euresult.Transitions)
		txIds[i] = euresult.ID
	}
	threadNum := 6
	transitionses := make([][]interfaces.Univalue, threadNum)
	worker := func(start, end, index int, args ...interface{}) {
		for i := start; i < end; i++ {
			transitionses[index] = append(transitionses[index], univaluepk.UnivaluesDecode(euresults[i].Transitions, func() interface{} { return &univaluepk.Univalue{} }, nil)...)
		}
	}
	common.ParallelWorker(len(euresults), threadNum, worker)

	transitions := make([]interfaces.Univalue, 0, transitionsize)
	for _, trans := range transitionses {
		transitions = append(transitions, trans...)
	}
	return txIds, transitions
}
