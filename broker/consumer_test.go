package broker

// type Actor interface {
// 	Consume(data interface{})
// }

// type mockActor1 struct {
// 	// c chan interface{}
// }

// func (mw1 *mockActor1) Consume(data interface{}) {
// 	mw1.c <- data
// }

// func (mw1 *mockActor1) start() {
// 	for v := range mw1.c {
// 		fmt.Printf("e0: %v, e1: %v\n", v.([]interface{})[0].(*mockState), v.([]interface{})[1].(*mockState))
// 	}
// }

// func TestConsumer(t *testing.T) {
// 	broker := NewStatefulBroker()
// 	p1 := NewDefaultProducer("A", []string{"e0"}, []int{0})
// 	p2 := NewDefaultProducer("B", []string{"e1"}, []int{0})

// 	broker.RegisterProducer(p1)
// 	broker.RegisterProducer(p2)

// 	actor := mockActor1{
// 		// c: make(chan interface{}),
// 	}

// 	c := NewDefaultConsumer("C",
// 		[]string{"e0", "e1"},
// 		NewConjunctions(&actor))

// 	broker.RegisterConsumer(c)

// 	broker.Serve()
// 	go actor.start()

// 	broker.Send("e0", &mockState{"1"})
// 	broker.Send("e1", &mockState{"2"})
// 	time.Sleep(time.Second)
// }
