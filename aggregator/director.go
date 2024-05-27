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
)

type Director struct {
	actor.WorkerThread
	msgRelation map[string]string
}

// return a Subscriber struct
func NewDirector(concurrency int, groupid string, msgRelation map[string]string) *Director {
	d := Director{}
	d.Set(concurrency, groupid)
	d.msgRelation = msgRelation
	return &d
}

func (d *Director) OnStart() {
}

func (d *Director) Stop() {

}

func (d *Director) OnMessageArrived(msgs []*actor.Message) error {
	for _, v := range msgs {
		if sendMsg, ok := d.msgRelation[v.Name]; ok {
			d.LatestMessage = v
			d.MsgBroker.LatestMessage = v
			d.MsgBroker.Send(sendMsg, v.Data)
		}
	}
	return nil
}
