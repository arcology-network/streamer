package aggregator

import (
	evmCommon "github.com/arcology-network/evm/common"
)

type Hashable interface {
	GetList() (selectList []*evmCommon.Hash, clearList []*evmCommon.Hash)
}

type Aggregator struct {
	selector *Selector
}

func NewAggregator() *Aggregator {
	return &Aggregator{
		selector: NewSelector(),
	}
}
func (a *Aggregator) GetClearInfos() []*evmCommon.Hash {
	return a.selector.clearance
}
func (a *Aggregator) GetRemaining() int {
	return a.selector.Remaining()
}

// action when clear received
func (a *Aggregator) OnClearInfoReceived() int {
	a.selector.Clear()
	return a.selector.Remaining()
}

// action when a data received
func (a *Aggregator) OnDataReceived(h evmCommon.Hash, data interface{}) *[]*interface{} {
	a.selector.OnDataReceived(h, data)
	return a.packSelected(a.selector.GetSelected())
}

// action when list received
func (a *Aggregator) OnListReceived(hashs Hashable) (*[]*interface{}, int) {
	missingSize := a.selector.GenerateMissing(hashs)
	return a.packSelected(a.selector.GetSelected()), missingSize
}

// send these raws, objs to pub
func (a *Aggregator) packSelected(completed bool, objs *[]*interface{}) *[]*interface{} {
	if completed {
		return objs
	}
	return nil
}

// action when clear list received
func (a *Aggregator) OnClearListReceived(hashs Hashable) {
	a.selector.SetClearance(hashs)
}

// reap txs
func (a *Aggregator) Reap(max int) (*[]*evmCommon.Hash, *[]*interface{}) {
	return a.selector.Reap(max)
}
