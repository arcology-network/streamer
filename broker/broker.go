package broker

import (
	"reflect"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	promProduceCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sstreamer_received_msgs_total",
			Help: "The total number of received messages.",
		},
		[]string{"channel"},
	)
	promConsumeCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sstreamer_sent_msgs_total",
			Help: "The total number of sent messages.",
		},
		[]string{"channel"},
	)
)

type StatefulBroker struct {
	producers []Producer
	consumers []Consumer
	buffers   map[string]streamBuffer
}

type Producer interface {
	Name() string
	Outputs() []string
	BufferLengths() []int
}

type Consumer interface {
	Name() string
	Inputs() []string
	StreamController() StreamController
}

func NewStatefulBroker() *StatefulBroker {
	return &StatefulBroker{
		buffers: make(map[string]streamBuffer),
	}
}

func (ss *StatefulBroker) RegisterProducer(p Producer) {
	ss.producers = append(ss.producers, p)
}

func (ss *StatefulBroker) RegisterConsumer(c Consumer) {
	ss.consumers = append(ss.consumers, c)
}

func (ss *StatefulBroker) Send(name string, data interface{}) {
	ss.buffers[name].Add(data)
}

func (ss *StatefulBroker) Serve() {
	for _, p := range ss.producers {
		outputs := p.Outputs()
		lengths := p.BufferLengths()
		for i := range outputs {
			if lengths[i] != 0 {
				ss.buffers[outputs[i]] = newDefaultStreamBuffer(outputs[i], lengths[i])
			} else {
				ss.buffers[outputs[i]] = newStaticStreamBuffer(outputs[i])
			}
		}
	}

	for _, c := range ss.consumers {
		inputs := c.Inputs()
		for i := range inputs {
			if inputs[i][0] == '#' {
				c.StreamController().GetListener(inputs[i])
				continue
			}
			ss.buffers[inputs[i]].RegisterListener(c.StreamController().GetListener(inputs[i]))
		}
		go c.StreamController().Serve()
	}

	for _, buf := range ss.buffers {
		go buf.Serve()
	}
}

func (ss *StatefulBroker) GenerateDot() string {
	str := "digraph g {\n"
	str += "\tStreamer [shape=record]\n"
	for _, p := range ss.producers {
		outputs := p.Outputs()
		for _, o := range outputs {
			str += "\t" + p.Name() + " -> Streamer [label=\"" + o + "\", shape=record]\n"
			str += "\t" + p.Name() + " [shape=record]\n"
		}
	}
	intermediate := make(map[string]struct{})
	for _, c := range ss.consumers {
		inputs := c.Inputs()
		for _, i := range inputs {
			if i[0] == '#' {
				intermediate[string(i[1:])] = struct{}{}
				continue
			}
			str += "\tStreamer -> " + c.Name() + " [label=\"" + i + "\"]\n"
		}
		worker := c.StreamController().GetActor()
		switch w := worker.(type) {
		case *ShortCircuitActor:
			str += "\t" + c.Name() + " -> " + w.consumer.Name() + " [label=\"" + w.name + "\"]\n"
		default:
			str += "\t" + c.Name() + " -> " + reflect.TypeOf(w).Elem().Name() + " [shape=circle]\n"
		}
	}
	rank := "\t{ rank=same "
	for _, c := range ss.consumers {
		str += "\t" + c.Name() + " [label=<<b>" + c.Name() + "</b><br/>"
		str += c.StreamController().GenerateDotLabel() + ">,"
		if _, ok := intermediate[c.Name()]; ok {
			str += " shape=record style=rounded color=grey bgcolor=grey"
		} else {
			str += " shape=record"
			rank += c.Name() + " "
		}
		str += "]\n"
	}
	str += rank + "}\n"
	str += "}"
	return str
}
