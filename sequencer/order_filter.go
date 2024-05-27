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

package sequencer

type OrderFilter struct {
	height    uint64
	cursor    uint32
	whitelist []string
	slots     []interface{}
}

func NewOrderFilter(names []string, height uint64) *OrderFilter {
	return &OrderFilter{
		height:    height,
		cursor:    0,
		whitelist: names,
		slots:     make([]interface{}, len(names)),
	}
}

func (this *OrderFilter) Less(lhs interface{}, rhs interface{}) bool {
	return lhs.(MessageInterface).Height() < rhs.(MessageInterface).Height()
}

func (this *OrderFilter) Check(msg interface{}) bool {
	if len(this.whitelist) == 0 {
		return true
	}

	if this.height > msg.(MessageInterface).Height() {
		panic("Error: Wrong message height")
	}

	for _, name := range this.whitelist {
		if msg.(MessageInterface).Name() == name {
			return true
		}
	}
	return false
}

func (this *OrderFilter) Filter(msg interface{}) (interface{}, bool, bool) {
	this.slots[this.find(msg)] = msg

	var outMsg interface{}
	if this.slots[this.cursor] != nil {
		outMsg = this.slots[this.cursor]
		this.cursor++
		this.cursor = this.cursor % uint32(len(this.slots))
		return outMsg, true, true
	}
	return msg, false, false
}

func (this *OrderFilter) find(msg interface{}) int {
	for i := 0; i < len(this.whitelist); i++ {
		if msg.(MessageInterface).Name() == this.whitelist[i] {
			return i
		}
	}
	return -1
}
