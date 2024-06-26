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
)

type Aggregator struct {
	actor.WorkerThread

	aggrMsgs []string
	pickMsg  string
	clearMsg string
	outMsg   string
	dataList []interface{}
	dataType reflect.Type
}

func NewAggregator(
	concurrency int, groupID string,
	aggrMsgs []string, pickMsg, clearMsg, outMsg string,
	dataType reflect.Type,
) *Aggregator {
	aggr := &Aggregator{
		aggrMsgs: aggrMsgs,
		pickMsg:  pickMsg,
		clearMsg: clearMsg,
		outMsg:   outMsg,
		dataType: dataType,
	}
	aggr.Set(concurrency, groupID)
	return aggr
}

func (aggr *Aggregator) OnStart() {}

func (aggr *Aggregator) OnMessageArrived(msgs []*actor.Message) error {
	msg := msgs[0]
	switch msg.Name {
	case aggr.pickMsg:
		aggr.sendData()
		break
	case aggr.clearMsg:
		aggr.clear()
		break
	default:
		found := false
		for _, name := range aggr.aggrMsgs {
			if name == msg.Name {
				found = true
				break
			}
		}
		if !found {
			// Log error.
		} else {
			aggr.addData(msg)
		}
		break
	}
	return nil
}

func (aggr *Aggregator) addData(msg *actor.Message) {
	newList := reflect.ValueOf(msg.Data)
	for i := 0; i < newList.Len(); i++ {
		data := newList.Index(i).Interface()
		aggr.dataList = append(aggr.dataList, data)
	}
}

func (aggr *Aggregator) sendData() {
	dataToSend := reflect.MakeSlice(reflect.SliceOf(aggr.dataType), 0, len(aggr.dataList))
	for _, data := range aggr.dataList {
		dataToSend = reflect.Append(dataToSend, reflect.ValueOf(data))
	}
	aggr.MsgBroker.Send(aggr.outMsg, dataToSend.Interface())
}

func (aggr *Aggregator) clear() {
	aggr.dataList = aggr.dataList[:0]
}
