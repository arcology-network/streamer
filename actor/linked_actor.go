package actor

import (
	"reflect"

	"github.com/arcology-network/streamer/broker"
)

type LinkedActor interface {
	IWorkerEx
	Initializer

	Preprocess(msgs []*Message) ([]*Message, error)
	Postprocess(msgs []*Message) error
	Next(actor LinkedActor) LinkedActor
	EndWith(actor IWorkerEx) IWorkerEx
	GetNext() LinkedActor
	GetEnd() IWorkerEx
}

type BaseLinkedActor struct {
	WorkerThread

	next    LinkedActor
	end     IWorkerEx
	derived LinkedActor
}

func (base *BaseLinkedActor) Preprocess(msgs []*Message) ([]*Message, error) {
	return base.derived.Preprocess(msgs)
}

func (base *BaseLinkedActor) Postprocess(msgs []*Message) error {
	return base.derived.Postprocess(msgs)
}

func (base *BaseLinkedActor) Next(actor LinkedActor) LinkedActor {
	if base.end != nil {
		panic("next actor and end actor cannot coexist.")
	}
	base.next = actor
	return actor
}

func (base *BaseLinkedActor) EndWith(actor IWorkerEx) IWorkerEx {
	if base.next != nil {
		panic("next actor and end actor cannot coexist.")
	}
	base.end = actor
	return actor
}

func (base *BaseLinkedActor) GetNext() LinkedActor {
	return base.next
}

func (base *BaseLinkedActor) GetEnd() IWorkerEx {
	return base.end
}

func (base *BaseLinkedActor) Inputs() ([]string, bool) {
	if base.next != nil {
		return base.next.Inputs()
	}
	if base.end != nil {
		return base.end.Inputs()
	}
	return []string{}, false
}

func (base *BaseLinkedActor) Outputs() map[string]int {
	outputs := make(map[string]int)
	if base.next != nil {
		for output, bufferSize := range base.next.Outputs() {
			outputs[output] = bufferSize
		}
	}
	if base.end != nil {
		for output, bufferSize := range base.end.Outputs() {
			outputs[output] = bufferSize
		}
	}
	return outputs
}

func (base *BaseLinkedActor) InitMsgs() []*Message {
	if base.next != nil {
		return base.next.InitMsgs()
	}
	if base.end != nil {
		if _, ok := base.end.(Initializer); ok {
			return base.end.(Initializer).InitMsgs()
		}
	}
	return []*Message{}
}

func (base *BaseLinkedActor) Init(name string, broker *broker.StatefulStreamer) {
	base.WorkerThread.Init(name, broker)
	if base.next != nil {
		base.next.Init(name, broker)
	}

	if base.end != nil {
		base.end.Init(name, broker)
	}
}

func (base *BaseLinkedActor) OnStart() {
	if base.next != nil {
		base.next.OnStart()
	}

	if base.end != nil {
		base.end.OnStart()
	}
}
func (base *BaseLinkedActor) OnMessageArrived(msgs []*Message) error {
	msgs, err := base.Preprocess(msgs)
	if err != nil {
		return err
	}

	if len(msgs) == 0 {
		return nil
	}

	if base.next != nil {
		base.next.ChangeEnvironment(msgs[0])
		err := base.next.OnMessageArrived(msgs)
		if err != nil {
			return err
		}
	}

	if base.end != nil {
		base.end.ChangeEnvironment(msgs[0])
		err := base.end.OnMessageArrived(msgs)
		if err != nil {
			return err
		}
	}

	err = base.Postprocess(msgs)
	if err != nil {
		return err
	}

	return nil
}

func (base *BaseLinkedActor) SetDerived(derived LinkedActor) {
	base.derived = derived
}

func (base *BaseLinkedActor) GetClient(typ reflect.Type) interface{} {
	curr := LinkedActor(base)
	for {
		if curr.GetNext() != nil {
			curr = curr.GetNext()
			if reflect.TypeOf(curr).Implements(typ) {
				return curr
			}

			if l, ok := curr.(*linkable); ok {
				if reflect.TypeOf(l.worker).Implements(typ) {
					return l.worker
				}
			}
		} else {
			if curr.GetEnd() != nil {
				if reflect.TypeOf(curr.GetEnd()).Implements(typ) {
					return curr.GetEnd()
				}

				if l, ok := curr.GetEnd().(*linkable); ok {
					if reflect.TypeOf(l.worker).Implements(typ) {
						return l.worker
					}
				}
			}
			return nil
		}
	}
}

type linkable struct {
	BaseLinkedActor

	worker IWorkerEx
}

func (l *linkable) Preprocess(msgs []*Message) ([]*Message, error) {
	l.worker.ChangeEnvironment(msgs[0])
	l.worker.OnMessageArrived(msgs)
	return msgs, nil
}

func (l *linkable) Postprocess(msgs []*Message) error {
	return nil
}

func (ctrl *linkable) Inputs() ([]string, bool) {
	i1, c1 := ctrl.worker.Inputs()
	i2, c2 := ctrl.BaseLinkedActor.Inputs()
	if c1 || c2 {
		panic("cannot use conjunction workers in a chain")
	}
	return MergeInputs(i1, i2), false
}

func (ctrl *linkable) Outputs() map[string]int {
	return MergeOutputs(ctrl.worker.Outputs(), ctrl.BaseLinkedActor.Outputs())
}

func (ctrl *linkable) InitMsgs() []*Message {
	if _, ok := ctrl.worker.(Initializer); ok {
		return append(ctrl.worker.(Initializer).InitMsgs(), ctrl.BaseLinkedActor.InitMsgs()...)
	}
	return ctrl.BaseLinkedActor.InitMsgs()
}

func (l *linkable) OnStart() {
	l.worker.Init(l.Name, l.MsgBroker.MsgBroker)
	l.worker.OnStart()
	l.BaseLinkedActor.SetDerived(l)
	l.BaseLinkedActor.OnStart()
}

func MakeLinkable(worker IWorkerEx) LinkedActor {
	l := &linkable{
		worker: worker,
	}
	l.BaseLinkedActor.SetDerived(l)
	return l
}

func MergeInputs(i1, i2 []string) []string {
	inputsMap := make(map[string]struct{})
	for _, input := range i1 {
		inputsMap[input] = struct{}{}
	}
	for _, input := range i2 {
		inputsMap[input] = struct{}{}
	}

	inputs := make([]string, 0, len(inputsMap))
	for input := range inputsMap {
		inputs = append(inputs, input)
	}
	return inputs
}

func MergeOutputs(o1, o2 map[string]int) map[string]int {
	outputs := make(map[string]int)
	for output, bufferSize := range o1 {
		outputs[output] = bufferSize
	}
	for output, bufferSize := range o2 {
		outputs[output] = bufferSize
	}
	return outputs
}
