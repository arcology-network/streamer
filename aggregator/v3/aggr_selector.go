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
	evmCommon "github.com/ethereum/go-ethereum/common"
)

type AggrOperation interface {
	GetData(msg *scommon.Message) ([]evmCommon.Hash, []interface{})
	GetList(msg *scommon.Message) []evmCommon.Hash
	OnListFulfilled(data []interface{}, broker *actor.ExecutionContext)
	Outputs() map[string]int
	Config(params map[string]interface{})
}

type AggrSelector struct {
	dataMsg  string
	listMsg  string
	clearMsg string
	ds       *DataSet
	op       AggrOperation
	name     string
}

func NewAggrSelector(name string, dataMsg string, listMsg string, clearMsg string, op AggrOperation) *AggrSelector {
	aggr := &AggrSelector{
		dataMsg:  dataMsg,
		listMsg:  listMsg,
		clearMsg: clearMsg,
		ds:       NewDataSet(),
		op:       op,
		name:     name,
	}
	return aggr
}

func (aggr *AggrSelector) Inputs() ([]string, bool) {
	return []string{aggr.dataMsg, aggr.listMsg, aggr.clearMsg}, false
}

func (aggr *AggrSelector) Outputs() map[string]int {
	return aggr.op.Outputs()
}
func (aggr *AggrSelector) Config(params map[string]interface{}) {
	aggr.op.Config(params)
}
func (aggr *AggrSelector) RegisterActions(reg actor.ActionRegistrar) {
	reg.Register(aggr.dataMsg, aggr.ReceivedData)
	reg.Register(aggr.listMsg, aggr.ReceivedList)
	reg.Register(aggr.clearMsg, aggr.ReceivedClearCommand)
}

func (aggr *AggrSelector) ReceivedData(ctx *actor.ActionContext) error {
	msg := ctx.Messages[0]
	hashes, data := aggr.op.GetData(msg)
	for i, hash := range hashes {
		lists := aggr.ds.Add(hash, data[i], msg.Height)
		for _, list := range lists {
			aggr.op.OnListFulfilled(list, ctx.ExecCtx)
		}
	}
	return nil
}
func (aggr *AggrSelector) ReceivedList(ctx *actor.ActionContext) error {
	msg := ctx.Messages[0]
	list := aggr.op.GetList(msg)
	data := aggr.ds.Get(list, msg.Height)
	if data != nil {
		aggr.op.OnListFulfilled(data, ctx.ExecCtx)
	}
	return nil
}
func (aggr *AggrSelector) ReceivedClearCommand(ctx *actor.ActionContext) error {
	aggr.ds.Clear(ctx.Messages[0].Height)
	return nil
}
