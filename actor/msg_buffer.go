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

package actor

import (
	"container/list"
)

type MsgBuffer struct {
	bufs    map[string]*list.List
	isMatch MatcherFunc
}

type MatcherFunc func(msg *Message, args ...interface{}) bool

func NewMsgBuffer(isMatch MatcherFunc) *MsgBuffer {
	return &MsgBuffer{
		bufs:    make(map[string]*list.List),
		isMatch: isMatch,
	}
}

func (mb *MsgBuffer) Put(msg *Message) {
	if buf, ok := mb.bufs[msg.Name]; ok {
		buf.PushBack(msg)
	} else {
		mb.bufs[msg.Name] = list.New()
		mb.bufs[msg.Name].PushBack(msg)
	}
}

func (mb *MsgBuffer) Get(args ...interface{}) *Message {
	for _, buf := range mb.bufs {
		if buf.Len() == 0 {
			continue
		}

		front := buf.Front()
		msg := front.Value.(*Message)
		if mb.isMatch(msg, args...) {
			buf.Remove(front)
			return msg
		}
	}
	return nil
}
