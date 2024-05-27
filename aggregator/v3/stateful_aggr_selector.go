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
	"math"

	"github.com/arcology-network/streamer/actor"
	"github.com/arcology-network/streamer/log"
)

const (
	aggrStateInit = iota
	aggrStateCollecting
	aggrStateDone
)

type StatefulAggrSelector struct {
	actor.WorkerThread

	dataMsg  string
	listMsg  string
	clearMsg string
	ds       *DataSet
	op       AggrOperation
	state    int
	height   uint64
}

func NewStatefulAggrSelector(concurrency int, groupId string, dataMsg string, listMsg string, clearMsg string, op AggrOperation) actor.IWorkerEx {
	aggr := &StatefulAggrSelector{
		dataMsg:  dataMsg,
		listMsg:  listMsg,
		clearMsg: clearMsg,
		ds:       NewDataSet(),
		op:       op,
		state:    aggrStateDone,
	}
	aggr.Set(concurrency, groupId)
	return aggr
}

func (aggr *StatefulAggrSelector) Inputs() ([]string, bool) {
	return []string{aggr.dataMsg, aggr.listMsg, aggr.clearMsg}, false
}

func (aggr *StatefulAggrSelector) Outputs() map[string]int {
	return aggr.op.Outputs()
}

func (aggr *StatefulAggrSelector) Config(params map[string]interface{}) {
	aggr.op.Config(params)
}

func (aggr *StatefulAggrSelector) OnStart() {}

func (aggr *StatefulAggrSelector) OnMessageArrived(msgs []*actor.Message) error {
	msg := msgs[0]
	switch aggr.state {
	case aggrStateInit:
		if msg.Name == aggr.dataMsg {
			aggr.onDataReceived(msg)
		} else if msg.Name == aggr.listMsg {
			if aggr.onListReceived(msg) {
				aggr.state = aggrStateDone
				aggr.AddLog(log.LogLevel_Info, "[StatefulAggrSelector] list received, switch to aggrStateDone")
			} else {
				aggr.state = aggrStateCollecting
				aggr.AddLog(log.LogLevel_Info, "[StatefulAggrSelector] list received, switch to aggrStateCollecting")
			}
		}
	case aggrStateCollecting:
		if msg.Name == aggr.dataMsg {
			if aggr.onDataReceived(msg) {
				aggr.state = aggrStateDone
				aggr.AddLog(log.LogLevel_Info, "[StatefulAggrSelector] data received, switch to aggrStateDone")
			}
		}
	case aggrStateDone:
		if msg.Name == aggr.clearMsg {
			aggr.ds.Clear(msg.Height)
			aggr.state = aggrStateInit
			aggr.height = msg.Height + 1
			aggr.AddLog(log.LogLevel_Info, fmt.Sprintf("[StatefulAggrSelector] clear on height %d", msg.Height))
		}
		// } else {
		// 	// FIXME
		// 	// MsgInitDB walk throught aggr selector, height inited before first MsgBlockCompleted by mistake.
		// 	aggr.height = 0
		// }
	}
	return nil
}

func (aggr *StatefulAggrSelector) GetStateDefinitions() map[int][]string {
	return map[int][]string{
		aggrStateInit:       {aggr.dataMsg, aggr.listMsg},
		aggrStateCollecting: {aggr.dataMsg},
		aggrStateDone:       {aggr.clearMsg},
	}
}

func (aggr *StatefulAggrSelector) GetCurrentState() int {
	return aggr.state
}

func (aggr *StatefulAggrSelector) Height() uint64 {
	if aggr.height == 0 {
		// var na int
		// var state state.State
		// if err := intf.Router.Call("tmstatestore", "Load", &na, &state); err != nil {
		// 	panic(err)
		// }
		// aggr.height = uint64(state.LastBlockHeight + 1)
		// fmt.Printf("[StatefulAggrSelector.Height] Init height to %d\n", aggr.height)

		// Before first MsgBlockCompleted, we accept all the messages.
		return math.MaxUint64
	}
	return aggr.height
}

func (aggr *StatefulAggrSelector) onDataReceived(msg *actor.Message) (fulfilled bool) {
	hashes, data := aggr.op.GetData(msg)
	for i, hash := range hashes {
		lists := aggr.ds.Add(hash, data[i], msg.Height)
		for _, list := range lists {
			fulfilled = true
			aggr.op.OnListFulfilled(list, aggr.MsgBroker)
		}
	}
	return
}

func (aggr *StatefulAggrSelector) onListReceived(msg *actor.Message) bool {
	list := aggr.op.GetList(msg)
	data := aggr.ds.Get(list, msg.Height)
	if data != nil {
		aggr.op.OnListFulfilled(data, aggr.MsgBroker)
		return true
	}
	return false
}
