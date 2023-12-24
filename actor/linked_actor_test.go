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
