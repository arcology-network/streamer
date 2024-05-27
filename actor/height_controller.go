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
	"fmt"
	"reflect"
)

type HeightSensitive interface {
	Height() uint64
}

type HeightController struct {
	BaseLinkedActor

	client HeightSensitive
	buf    *MsgBuffer
}

func NewHeightController() *HeightController {
	controller := &HeightController{}
	controller.BaseLinkedActor.SetDerived(controller)
	return controller
}

func (ctrl *HeightController) Preprocess(msgs []*Message) ([]*Message, error) {
	if len(msgs) != 1 {
		panic("cannot handle more than one message.")
	}
	msg := msgs[0]

	if msg.Height <= ctrl.client.Height() {
		return msgs, nil
	} else {
		ctrl.buf.Put(msg)
		fmt.Printf("HC: Push %v\n", msg)
	}
	return nil, nil
}

func (ctrl *HeightController) Postprocess(msgs []*Message) error {
	msg := ctrl.buf.Get(ctrl.client.Height())
	if msg != nil {
		fmt.Printf("HC: Pop %v\n", msg)
		ctrl.ChangeEnvironment(msg)
		ctrl.OnMessageArrived([]*Message{msg})
	}
	return nil
}

func (ctrl *HeightController) OnStart() {
	ctrl.BaseLinkedActor.SetDerived(ctrl)
	client := ctrl.GetClient(reflect.TypeOf((*HeightSensitive)(nil)).Elem())
	if client == nil {
		panic("no height sensitive consumers found.")
	}
	ctrl.client = client.(HeightSensitive)
	ctrl.buf = NewMsgBuffer(func(msg *Message, args ...interface{}) bool {
		return msg.Height <= args[0].(uint64)
	})
	ctrl.BaseLinkedActor.OnStart()
}
