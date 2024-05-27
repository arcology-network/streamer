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

	evmCommon "github.com/ethereum/go-ethereum/common"
)

type pendingTask struct {
	waitingList map[evmCommon.Hash]struct{}
	fulfilled   []interface{}
	done        bool
}

type subSet struct {
	all   map[evmCommon.Hash]interface{}
	tasks []*pendingTask
}

func newSubSet() *subSet {
	return &subSet{
		all: make(map[evmCommon.Hash]interface{}),
	}
}

func (ss *subSet) add(key evmCommon.Hash, value interface{}) (fulfilledLists [][]interface{}) {
	ss.all[key] = value
	for _, task := range ss.tasks {
		if task.done {
			continue
		}

		if _, ok := task.waitingList[key]; ok {
			delete(task.waitingList, key)
			task.fulfilled = append(task.fulfilled, value)
			if len(task.waitingList) == 0 {
				fulfilledLists = append(fulfilledLists, task.fulfilled)
				task.done = true
			}
		}
	}
	return
}

func (ss *subSet) get(list []evmCommon.Hash) []interface{} {
	waitingList := make(map[evmCommon.Hash]struct{})
	fulfilled := []interface{}{}
	for _, key := range list {
		if value, ok := ss.all[key]; ok {
			fulfilled = append(fulfilled, value)
		} else {
			waitingList[key] = struct{}{}
		}
	}

	if len(waitingList) == 0 {
		return fulfilled
	}

	ss.tasks = append(ss.tasks, &pendingTask{
		waitingList: waitingList,
		fulfilled:   fulfilled,
	})
	return nil
}

func (ss *subSet) String() string {
	return fmt.Sprintf("%v", ss.all)
}

type DataSet struct {
	subSets map[uint64]*subSet
}

func NewDataSet() *DataSet {
	return &DataSet{
		subSets: make(map[uint64]*subSet),
	}
}

func (ds *DataSet) Add(key evmCommon.Hash, value interface{}, height uint64) [][]interface{} {
	if _, ok := ds.subSets[height]; !ok {
		ds.subSets[height] = newSubSet()
	}
	return ds.subSets[height].add(key, value)
}

func (ds *DataSet) Get(list []evmCommon.Hash, height uint64) []interface{} {
	if _, ok := ds.subSets[height]; !ok {
		ds.subSets[height] = newSubSet()
	}
	return ds.subSets[height].get(list)
}

func (ds *DataSet) Clear(height uint64) {
	for h := range ds.subSets {
		if h <= height {
			delete(ds.subSets, h)
		}
	}
}

func (ds *DataSet) String() string {
	return fmt.Sprintf("%v", ds.subSets)
}
