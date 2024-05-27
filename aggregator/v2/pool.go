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

package v2

import (
	"reflect"

	"github.com/arcology-network/streamer/actor"
	"github.com/arcology-network/streamer/aggregator/v2/aggregator"
)

type AggreSelector struct {
	actor.WorkerThread
	impl *aggregator.Aggregator
}

type brokerAdaptor struct {
	wt *actor.WorkerThread
}

func (adaptor *brokerAdaptor) Send(name string, msg *actor.Message) error {
	adaptor.wt.MsgBroker.Send(name, msg)
	return nil
}

func NewAggreSelector(
	concurrency int,
	groupID string,
	clearMsg, dataMsg, pickListMsg, clearListMsg, outMsg string,
	dataType reflect.Type) *AggreSelector {
	aggreSelector := &AggreSelector{}
	aggreSelector.Set(concurrency, groupID)
	aggreSelector.impl = aggregator.NewAggregator(
		&brokerAdaptor{wt: &aggreSelector.WorkerThread},
		clearMsg, dataMsg, pickListMsg, clearListMsg, outMsg,
		dataType)
	return aggreSelector
}

func (as *AggreSelector) OnStart() {}

func (as *AggreSelector) OnMessageArrived(msgs []*actor.Message) error {
	as.impl.HandleMsg(msgs[0])
	return nil
}
