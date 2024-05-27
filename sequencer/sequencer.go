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

import (
	"sort"
)

type Sequencer struct {
	buffer []interface{}
	filter FilterInterface
}

func NewSequencer(filter FilterInterface) *Sequencer {
	return &Sequencer{
		buffer: make([]interface{}, 0, 1024),
		filter: filter,
	}
}

/*
newMsg == nil && passed == true  && changed == true: the msg triggerd state change, pass the msg(nil) an
d Retry all cached, INVALID ?
newMsg == nil && passed == false && changed == true: the msg triggerd state change and consumed by the filter              => Retry all cached
newMsg != nil && passed == true  && changed == true: the msg triggerd state change, pass the msg                           => Retry all cached
newMsg != nil && passed == false && changed == true: Seems INVALID

newMsg == nil && passed == true  && changed == false: Seems INVALID
newMsg == nil && passed == false && changed == false: Seems INVALID
newMsg != nil && passed == true  && changed == false: Forward the msg to the downstream actor
newMsg != nil && passed == false && changed == false: Buffer the msg
*/

func (this *Sequencer) Try(msg interface{}) []interface{} {
	if this.filter == nil || !this.filter.Check(msg) {
		return []interface{}{msg}
	}
	newMsg, passed, changed := this.filter.Filter(msg)

	outMsgs := []interface{}{}
	if newMsg != nil && passed && !changed {
		outMsgs = append(outMsgs, newMsg) // Send to the downstream consumers
		return outMsgs
	}

	if newMsg != nil && !passed && !changed {
		this.buffer = append(this.buffer, newMsg) // Buffer the message
		return outMsgs
	}

	if (newMsg == nil && !passed && changed) ||
		(newMsg != nil && passed && changed) {
		if newMsg != nil {
			outMsgs = append(outMsgs, newMsg)
		}
		this.retryCached(&outMsgs)
		return outMsgs
	}

	panic("Error: Invalid combinations !!!")
}

func (this *Sequencer) retryCached(outMsgs *[]interface{}) {
	for j := range this.buffer {
		if this.buffer[j] == nil {
			continue
		}

		if nMsg, passed, _ := this.filter.Filter(this.buffer[j]); nMsg != nil && passed {
			*outMsgs = append(*outMsgs, nMsg)
			this.buffer[j] = nil
		}
	}
	RemoveNils(&this.buffer)
}

func (this *Sequencer) ToBuffer(msg interface{}) {
	this.buffer = append(this.buffer, msg)

	sort.SliceStable(this.buffer, func(i, j int) bool {
		return this.filter.Less(this.buffer[i], this.buffer[j])
	})
}

type Sequencers []*Sequencer

func (this Sequencers) Try(msg interface{}) []interface{} {
	outMsgs := this[0].Try(msg)
	for i := range outMsgs {
		outMsgs = this[1].Try(outMsgs[i])
	}
	return outMsgs
}

func (this Sequencers) BatchTry(msgs []interface{}) []interface{} {
	outMsgs := []interface{}{}
	for i := range msgs {
		outMsgs = append(outMsgs, this.Try(msgs[i])...)
	}
	return outMsgs
}

func RemoveNils(values *[]interface{}) {
	pos := int64(-1)
	for i := 0; i < len((*values)); i++ {
		if pos < 0 && (*values)[i] == nil {
			pos = int64(i)
			continue
		}

		if pos < 0 && (*values)[i] != nil {
			continue
		}

		if pos >= 0 && (*values)[i] == nil {
			(*values)[pos] = (*values)[i]
			continue
		}

		(*values)[pos] = (*values)[i]
		pos++
	}

	if pos >= 0 {
		(*values) = (*values)[:pos]
	}
}
