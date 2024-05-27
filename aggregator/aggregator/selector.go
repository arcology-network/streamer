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
	"github.com/arcology-network/common-lib/common"
	evmCommon "github.com/ethereum/go-ethereum/common"
)

type Selector struct {
	clearance    []evmCommon.Hash
	selectedList []*interface{}
	missingList  *MissingList
	pool         *DataPool
}

// NewDataPool returns a new DataPool structure.
func NewSelector() *Selector {
	return &Selector{
		clearance:    []evmCommon.Hash{},
		selectedList: []*interface{}{},
		missingList:  NewMissingList(),
		pool:         NewDataPool(),
	}
}

// load data into missing when list received
func (s *Selector) GenerateMissing(hashs Hashable) int {
	s.missingList = NewMissingList()
	selectList, clearance := hashs.GetList()
	s.clearance = clearance
	s.selectedList = make([]*interface{}, len(selectList))

	worker := func(start, end, idx int, args ...interface{}) {
		selectingList := args[0].([]interface{})[0].([]evmCommon.Hash)
		selectedList := args[0].([]interface{})[1].([]*interface{})
		missingist := args[0].([]interface{})[2].(*MissingList)
		for i := start; i < end; i++ {
			hash := selectingList[i]
			obj := s.pool.get(hash)
			if obj != nil {
				selectedList[i] = &obj
			} else {
				missingist.Put(hash, i)
			}
		}
	}
	common.ParallelWorker(len(selectList), 4, worker, selectList, s.selectedList, s.missingList)
	s.missingList.CompleteGnereation()
	return s.missingList.Size()
}

// received a new data
func (s *Selector) OnDataReceived(h evmCommon.Hash, data interface{}) {
	s.pool.add(h, data)
	if s.missingList.IsGenerated() {
		if ok, idx := s.missingList.RemoveFromMissing(h); ok {
			s.selectedList[idx] = &data
		}
	}
}

// clear from pool
func (s *Selector) Clear() {
	for _, h := range s.clearance {
		s.pool.remove(h)
	}
}

// remain quantity in pool
func (s *Selector) Remaining() int {
	return s.pool.count()
}

// get selected data if gather terminated
func (s *Selector) GetSelected() (bool, *[]*interface{}) {
	if s.missingList.IsReceiveCompleted() {
		retobjs := s.selectedList
		s.missingList.ClearList()
		s.selectedList = []*interface{}{}
		return true, &retobjs
	} else {
		return false, nil
	}

}

// reap txs with no list
func (s *Selector) Reap(max int) (*[]*evmCommon.Hash, *[]*interface{}) {
	hashs := []*evmCommon.Hash{}
	objs := []*interface{}{}
	selected := 0
	f := func(hash evmCommon.Hash, obj interface{}) bool {
		hashs = append(hashs, &hash)
		objs = append(objs, &obj)
		selected++

		if max > 0 && selected >= max {
			return false
		}
		return true
	}
	s.pool.Range(f)
	return &hashs, &objs
}

// set clear list
func (s *Selector) SetClearance(hashs Hashable) {
	_, s.clearance = hashs.GetList()
}
