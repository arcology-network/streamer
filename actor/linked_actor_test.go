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
	"reflect"
	"testing"
)

type model1 interface {
	M1()
}

type model2 interface {
	M2()
}

type model3 interface {
	M3()
}

type worker1 struct {
	WorkerThread
}

func (w *worker1) Inputs() ([]string, bool) {
	return []string{}, false
}

func (w *worker1) Outputs() map[string]int {
	return map[string]int{}
}

func (w *worker1) OnStart() {}

func (w *worker1) OnMessageArrived(msgs []*Message) error {
	return nil
}

func (w *worker1) M1() {}

type worker2 struct {
	WorkerThread
}

func (w *worker2) Inputs() ([]string, bool) {
	return []string{}, false
}

func (w *worker2) Outputs() map[string]int {
	return map[string]int{}
}

func (w *worker2) OnStart() {}

func (w *worker2) OnMessageArrived(msgs []*Message) error {
	return nil
}

func (w *worker2) M2() {}

func TestGetClient(t *testing.T) {
	base := BaseLinkedActor{}
	base.Next(MakeLinkable(&worker1{})).EndWith(&worker2{})

	c1 := base.GetClient(reflect.TypeOf((*model1)(nil)).Elem())
	if c1 == nil {
		t.Fail()
	}

	c2 := base.GetClient(reflect.TypeOf((*model2)(nil)).Elem())
	if c2 == nil {
		t.Fail()
	}

	c3 := base.GetClient(reflect.TypeOf((*model3)(nil)).Elem())
	if c3 != nil {
		t.Fail()
	}
}
