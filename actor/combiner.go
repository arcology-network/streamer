package actor

import (
	"sort"
	"strings"

	brokerpk "github.com/arcology-network/streamer/broker"
)

const (
	CombinerPrefix = "__combine__"
)

func CombinedName(inputs ...string) string {
	sort.Strings(inputs)
	return CombinerPrefix + strings.Join(inputs, "-")
}

type CombinerElements struct {
	Msgs map[string]*Message
}

func (cl *CombinerElements) Get(msgname string) *Message {
	if msg, ok := cl.Msgs[msgname]; ok {
		return msg
	} else {
		return nil
	}
}

type Combiner struct {
	WorkerThread
	inputs []string
	outMsg string
	els    *CombinerElements
}

func Combine(inputs ...string) *Combiner {
	return &Combiner{
		inputs: inputs,
		outMsg: CombinedName(inputs...),
		els: &CombinerElements{
			Msgs: map[string]*Message{},
		},
	}
}

func (c *Combiner) On(broker *brokerpk.StatefulStreamer) *Combiner {
	combiner := NewActorEx(
		c.outMsg,
		broker,
		c,
	)
	combiner.Connect(brokerpk.NewConjunctions(combiner))
	return c
}

func (c *Combiner) Inputs() ([]string, bool) {
	return c.inputs, true
}

func (c *Combiner) Outputs() map[string]int {
	return map[string]int{
		c.outMsg: 1,
	}
}

func (c *Combiner) OnStart() {
}

func (c *Combiner) Stop() {

}

func (c *Combiner) OnMessageArrived(msgs []*Message) error {
	for _, v := range msgs {
		c.els.Msgs[v.Name] = v
	}
	c.MsgBroker.Send(c.outMsg, c.els)
	c.els = &CombinerElements{
		Msgs: map[string]*Message{},
	}
	return nil
}
