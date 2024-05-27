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

type FSMCompatible interface {
	IWorker

	GetStateDefinitions() map[int][]string
	GetCurrentState() int
}

type FSMController struct {
	BaseLinkedActor

	client       FSMCompatible
	defs         map[int][]string
	universalSet map[string]struct{}
	buf          *MsgBuffer
}

func NewFSMController() *FSMController {
	controller := &FSMController{}
	controller.BaseLinkedActor.SetDerived(controller)
	return controller
}

func (ctrl *FSMController) Preprocess(msgs []*Message) ([]*Message, error) {
	if len(msgs) != 1 {
		panic("cannot handle more than one message.")
	}
	msg := msgs[0]

	// No need to control.
	if _, ok := ctrl.universalSet[msg.Name]; !ok {
		return msgs, nil
	}

	state := ctrl.client.GetCurrentState()
	if acceptables, ok := ctrl.defs[state]; !ok {
		panic(fmt.Sprintf("unknown state (%v) given.", state))
	} else {
		isAcceptable := false
		for _, typ := range acceptables {
			if typ == msg.Name {
				isAcceptable = true
				break
			}
		}

		if isAcceptable {
			return msgs, nil
		} else {
			ctrl.buf.Put(msg)
			fmt.Printf("FSM: Push %v\n", msg)
		}
	}
	return nil, nil
}

func (ctrl *FSMController) Postprocess(msgs []*Message) error {
	state := ctrl.client.GetCurrentState()
	if state == -1 {
		fmt.Printf("FSM: State definitions updated.\n")
		ctrl.defs = ctrl.client.GetStateDefinitions()
		state = ctrl.client.GetCurrentState()
	}
	if acceptables, ok := ctrl.defs[state]; !ok {
		panic(fmt.Sprintf("unknown state (%v) given.", state))
	} else {
		msg := ctrl.buf.Get(acceptables)
		if msg != nil {
			fmt.Printf("FSM: Pop %v\n", msg)
			ctrl.ChangeEnvironment(msg)
			ctrl.OnMessageArrived([]*Message{msg})
		}
	}
	return nil
}

func (ctrl *FSMController) OnStart() {
	ctrl.BaseLinkedActor.SetDerived(ctrl)
	client := ctrl.GetClient(reflect.TypeOf((*FSMCompatible)(nil)).Elem())
	if client == nil {
		panic("a non-FSMCompatible actor given.")
	}
	ctrl.client = client.(FSMCompatible)
	ctrl.defs = ctrl.client.GetStateDefinitions()

	ctrl.universalSet = make(map[string]struct{})
	for _, def := range ctrl.defs {
		for _, typ := range def {
			ctrl.universalSet[typ] = struct{}{}
		}
	}

	ctrl.buf = NewMsgBuffer(func(msg *Message, args ...interface{}) bool {
		acceptables := args[0].([]string)
		for _, acceptable := range acceptables {
			if msg.Name == acceptable {
				return true
			}
		}
		return false
	})
	ctrl.BaseLinkedActor.OnStart()
}
