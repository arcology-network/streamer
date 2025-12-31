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
	"context"
	"fmt"
	"math"

	"github.com/arcology-network/streamer/actor"
	scommon "github.com/arcology-network/streamer/common"
	"github.com/arcology-network/streamer/logger"
)

const (
	aggrStateInit = iota
	aggrStateCollecting
	aggrStateDone
)

type StatefulAggrSelector struct {
	dataMsg  string
	listMsg  string
	clearMsg string
	ds       *DataSet
	op       AggrOperation
	state    int
	height   uint64
	name     string
}

func NewStatefulAggrSelector(name string, dataMsg string, listMsg string, clearMsg string, op AggrOperation) *StatefulAggrSelector {
	aggr := &StatefulAggrSelector{
		dataMsg:  dataMsg,
		listMsg:  listMsg,
		clearMsg: clearMsg,
		ds:       NewDataSet(),
		op:       op,
		state:    aggrStateInit,
		name:     name,
	}
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

func (aggr *StatefulAggrSelector) RegisterActions(reg actor.ActionRegistrar) {
	reg.Register(aggr.dataMsg, aggr.ReceivedData)
	reg.Register(aggr.listMsg, aggr.ReceivedList)
	reg.Register(aggr.clearMsg, aggr.ReceivedClearCommand)
}
func (aggr *StatefulAggrSelector) ReceivedData(ctx *actor.ActionContext) error {
	msg := ctx.Messages[0]
	switch aggr.state {
	case aggrStateInit:
		aggr.onDataReceived(msg, ctx.ExecCtx)
	case aggrStateCollecting:
		if aggr.onDataReceived(msg, ctx.ExecCtx) {
			aggr.state = aggrStateDone
			logger.Log.Debug(context.Background(), "[StatefulAggrSelector] data received, switch to aggrStateDone")
		}
	}
	return nil
}

func (aggr *StatefulAggrSelector) ReceivedList(ctx *actor.ActionContext) error {
	msg := ctx.Messages[0]
	if aggr.onListReceived(msg, ctx.ExecCtx) {
		aggr.state = aggrStateDone
		logger.Log.Debug(context.Background(), "[StatefulAggrSelector] list received, switch to aggrStateDone")
	} else {
		aggr.state = aggrStateCollecting
		logger.Log.Debug(context.Background(), "[StatefulAggrSelector] list received, switch to aggrStateCollecting")
	}
	return nil
}

func (aggr *StatefulAggrSelector) ReceivedClearCommand(ctx *actor.ActionContext) error {
	msg := ctx.Messages[0]
	aggr.ds.Clear(msg.Height)
	aggr.state = aggrStateInit
	aggr.height = msg.Height + 1
	logger.Log.Debug(context.Background(), fmt.Sprintf("[StatefulAggrSelector] clear on height %d", msg.Height))
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
		return math.MaxUint64
	}
	return aggr.height
}

func (aggr *StatefulAggrSelector) onDataReceived(msg *scommon.Message, ctx *actor.ExecutionContext) (fulfilled bool) {
	hashes, data := aggr.op.GetData(msg)
	for i, hash := range hashes {
		lists := aggr.ds.Add(hash, data[i], msg.Height)
		for _, list := range lists {
			fulfilled = true
			aggr.op.OnListFulfilled(list, ctx)
		}
	}
	return
}

func (aggr *StatefulAggrSelector) onListReceived(msg *scommon.Message, ctx *actor.ExecutionContext) bool {
	list := aggr.op.GetList(msg)
	data := aggr.ds.Get(list, msg.Height)
	if data != nil {
		aggr.op.OnListFulfilled(data, ctx)
		return true
	}
	return false
}
