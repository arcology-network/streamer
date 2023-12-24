package aggregator

import evmCommon "github.com/arcology-network/evm/common"

type AggregateCollection struct {
	accumulated []interface{}
}

func NewAggregateCollection() *AggregateCollection {
	return &AggregateCollection{
		accumulated: []interface{}{},
	}
}

// action when clear received
func (ac *AggregateCollection) OnReset() {
	ac.accumulated = []interface{}{}
}

// action when new batch received
func (ac *AggregateCollection) OnNewBatchReceived(addBatch *[]interface{}) *[]interface{} {
	for _, v := range *addBatch {
		ac.accumulated = append(ac.accumulated, v)
	}
	return &ac.accumulated
}

// get all items
func (ac *AggregateCollection) GetAll() *[]interface{} {
	return &ac.accumulated
}

func (ac *AggregateCollection) ConvertToHashList(raws *[]interface{}) *[]*evmCommon.Hash {
	if raws == nil || len(*raws) == 0 {
		return &[]*evmCommon.Hash{}
	}
	rets := make([]*evmCommon.Hash, len(*raws))
	for i, v := range *raws {
		rets[i] = v.(*evmCommon.Hash)
	}
	return &rets
}
