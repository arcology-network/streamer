package broker

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

type mockState struct {
	v string
}

func (ms *mockState) Equals(rhs Comparable) bool {
	return ms.v == rhs.(*mockState).v
}

type mockActor1 struct {
	c chan interface{}
}

func (mw1 *mockActor1) Consume(data interface{}) {
	mw1.c <- data
}

func (mw1 *mockActor1) start() {
	for v := range mw1.c {
		fmt.Printf("e0: %v, e1: %v\n", v.([]interface{})[0].(*mockState), v.([]interface{})[1].(*mockState))
	}
}

type mockActor2 struct {
	c chan interface{}
}

func (mw2 *mockActor2) Consume(data interface{}) {
	mw2.c <- data
}

func (mw2 *mockActor2) serve() {
	for v := range mw2.c {
		fmt.Printf("%v: %v\n", v.(Aggregated).Name, v.(Aggregated).Data)
	}
}

func TestStateAggregator(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.JSONFormatter{})

	broker := NewStatefulBroker()
	p1 := NewDefaultProducer("A", []string{"e0"}, []int{0})
	p2 := NewDefaultProducer("B", []string{"e1"}, []int{0})

	broker.RegisterProducer(p1)
	broker.RegisterProducer(p2)

	actor := mockActor1{
		c: make(chan interface{}),
	}
	c := NewDefaultConsumer("C",
		[]string{"e0", "e1"},
		NewConjunctions(&actor))

	broker.RegisterConsumer(c)

	broker.Serve()
	go actor.start()

	broker.Send("e0", &mockState{"1"})
	broker.Send("e1", &mockState{"2"})
	time.Sleep(time.Second)

	broker.Send("e0", &mockState{"1"})
	broker.Send("e0", &mockState{"2"})
	broker.Send("e1", &mockState{"1"})
	time.Sleep(time.Second)
}

func TestStreamAggregator(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.JSONFormatter{})

	broker := NewStatefulBroker()

	actorA := NewDefaultProducer("A", []string{"Stream1"}, []int{1})
	broker.RegisterProducer(actorA)

	actorB := NewDefaultProducer("B", []string{"Stream2"}, []int{1})
	broker.RegisterProducer(actorB)

	actorC := mockActor1{
		c: make(chan interface{}),
	}
	c := NewDefaultConsumer("C", []string{"Stream1", "Stream2"}, NewConjunctions(&actorC))
	broker.RegisterConsumer(c)

	broker.Serve()
	go actorC.start()

	broker.Send("Stream1", &mockState{"1"})
	broker.Send("Stream2", &mockState{"2"})
	time.Sleep(time.Second)

	broker.Send("Stream1", &mockState{"1"})
	broker.Send("Stream2", &mockState{"2"})
	time.Sleep(time.Second)
}

func TestDisjunctions(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.JSONFormatter{})

	broker := NewStatefulBroker()
	p1 := NewDefaultProducer("A", []string{"Stream1"}, []int{1})
	p2 := NewDefaultProducer("A", []string{"Stream2"}, []int{1})
	actor := mockActor2{
		c: make(chan interface{}),
	}
	c := NewDefaultConsumer("C", []string{"Stream1", "Stream2"}, NewDisjunctions(&actor, 3))

	broker.RegisterProducer(p1)
	broker.RegisterProducer(p2)
	broker.RegisterConsumer(c)
	broker.Serve()
	fmt.Println(broker.GenerateDot())
	go actor.serve()

	broker.Send("Stream1", &mockState{"1"})
	broker.Send("Stream2", &mockState{"2"})
	time.Sleep(time.Second)

	broker.Send("Stream1", &mockState{"1"})
	broker.Send("Stream1", &mockState{"1"})
	broker.Send("Stream2", &mockState{"2"})
	broker.Send("Stream1", &mockState{"1"})
	broker.Send("Stream1", &mockState{"1"})
	time.Sleep(time.Second)
}

func TestCustomizableController(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.JSONFormatter{})

	broker := NewStatefulBroker()
	p1 := NewDefaultProducer("A", []string{"Stream1"}, []int{5})
	p2 := NewDefaultProducer("B", []string{"Stream2"}, []int{5})
	actor := mockActor1{
		c: make(chan interface{}),
	}
	c := NewDefaultConsumer("C", []string{"Stream1", "Stream2"}, NewCustomizableController(
		&actor, func(inputs interface{}) bool {
			return inputs.([]interface{})[0] != nil && inputs.([]interface{})[1] != nil &&
				inputs.([]interface{})[0].(*mockState).Equals(inputs.([]interface{})[1].(*mockState))
		},
	))

	broker.RegisterProducer(p1)
	broker.RegisterProducer(p2)
	broker.RegisterConsumer(c)
	broker.Serve()
	fmt.Println(broker.GenerateDot())
	go actor.start()

	broker.Send("Stream1", &mockState{"1"})
	broker.Send("Stream2", &mockState{"1"})
	time.Sleep(time.Second)

	broker.Send("Stream1", &mockState{"2"})
	broker.Send("Stream2", &mockState{"2"})
	broker.Send("Stream2", &mockState{"2"})
	broker.Send("Stream2", &mockState{"2"})
	time.Sleep(time.Second)

	broker.Send("Stream1", &mockState{"3"})
	broker.Send("Stream2", &mockState{"1"})
	broker.Send("Stream2", &mockState{"3"})
	time.Sleep(time.Second)
}

func TestGenerateDot(t *testing.T) {
	var p1, p2, p3 Producer
	var c1, c2 Consumer
	broker := NewStatefulBroker()
	p1 = NewDefaultProducer("A", []string{"Stream1"}, []int{1})
	p2 = NewDefaultProducer("B", []string{"Stream2"}, []int{1})
	p3 = NewDefaultProducer("C", []string{"Stream3"}, []int{1})

	actor2 := mockActor2{
		c: make(chan interface{}),
	}
	c2 = NewDefaultConsumer("E", []string{"#D", "Stream3"}, NewDisjunctions(&actor2, 3))
	actor1 := NewShortCircuitActor(c2, "#D")
	c1 = NewDefaultConsumer("D", []string{"Stream1", "Stream2"}, NewConjunctions(actor1))
	actor3 := mockActor1{
		c: make(chan interface{}),
	}
	c3 := NewDefaultConsumer("H", []string{"Stream2", "Stream3"}, NewDisjunctions(&actor3, 3))

	broker.RegisterProducer(p1)
	broker.RegisterProducer(p2)
	broker.RegisterProducer(p3)
	broker.RegisterConsumer(c1)
	broker.RegisterConsumer(c2)
	broker.RegisterConsumer(c3)
	broker.Serve()
	fmt.Println(broker.GenerateDot())
}

func TestPromMetrics(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.JSONFormatter{})

	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":19002", nil)

	broker := NewStatefulBroker()
	p1 := NewDefaultProducer("A", []string{"Stream1"}, []int{1})
	p2 := NewDefaultProducer("B", []string{"Stream2"}, []int{1})
	actor := mockActor2{
		c: make(chan interface{}),
	}
	c := NewDefaultConsumer("C", []string{"Stream1", "Stream2"}, NewDisjunctions(&actor, 3))

	broker.RegisterProducer(p1)
	broker.RegisterProducer(p2)
	broker.RegisterConsumer(c)
	broker.Serve()
	go actor.serve()

	go func() {
		for {
			broker.Send("Stream1", &mockState{"1"})
			time.Sleep(time.Millisecond * 500)
		}
	}()
	go func() {
		for {
			broker.Send("Stream2", &mockState{"2"})
			time.Sleep(time.Second)
		}
	}()
	time.Sleep(time.Second * 10)
}

func TestCompoundController(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.JSONFormatter{})

	var p1, p2, p3 Producer
	var c1, c2 Consumer
	broker := NewStatefulBroker()
	p1 = NewDefaultProducer("A", []string{"Stream1"}, []int{1})
	p2 = NewDefaultProducer("B", []string{"Stream2"}, []int{1})
	p3 = NewDefaultProducer("C", []string{"Stream3"}, []int{1})

	actor2 := mockActor2{
		c: make(chan interface{}),
	}
	c2 = NewDefaultConsumer("E", []string{"#D", "Stream3"}, NewDisjunctions(&actor2, 3))
	actor1 := NewShortCircuitActor(c2, "#D")
	c1 = NewDefaultConsumer("D", []string{"Stream1", "Stream2"}, NewConjunctions(actor1))

	broker.RegisterProducer(p1)
	broker.RegisterProducer(p2)
	broker.RegisterProducer(p3)
	broker.RegisterConsumer(c1)
	broker.RegisterConsumer(c2)
	broker.Serve()
	fmt.Println(broker.GenerateDot())
	go actor1.Serve()
	go actor2.serve()

	broker.Send("Stream1", &mockState{"1"})
	broker.Send("Stream2", &mockState{"2"})
	broker.Send("Stream3", &mockState{"3"})
	time.Sleep(time.Second)
}
