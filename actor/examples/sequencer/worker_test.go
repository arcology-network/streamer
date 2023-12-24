package sequencer

import (
	"testing"
	"time"

	"github.com/arcology-network/component-lib/actor"
	streamer "github.com/arcology-network/component-lib/broker"
	"github.com/arcology-network/component-lib/log"
)

func TestSequencer(t *testing.T) {
	log.InitLog("testing.log", "./log.toml", "testing", "testing", 0)

	broker := streamer.NewStatefulBroker()
	workerBase := &actor.FSMController{}
	workerBase.EndWith(actor.AsSequencer(NewWorker(), []string{msgA, msgB, msgC}))
	workerActor := actor.NewActor(
		"worker",
		broker,
		[]string{
			msgA,
			msgB,
			msgC,
		},
		[]string{},
		[]int{},
		workerBase,
	)
	workerActor.Connect(streamer.NewDisjunctions(workerActor, 1))

	sender := streamer.NewDefaultProducer(
		"sender",
		[]string{
			msgA,
			msgB,
			msgC,
		},
		[]int{1, 1, 1},
	)
	broker.RegisterProducer(sender)

	broker.Serve()

	workerBase.OnStart()

	broker.Send(msgC, &actor.Message{
		Name: msgC,
		Data: 1,
	})
	broker.Send(msgB, &actor.Message{
		Name: msgB,
		Data: 2,
	})
	broker.Send(msgA, &actor.Message{
		Name: msgA,
		Data: 3,
	})
	broker.Send(msgB, &actor.Message{
		Name: msgB,
		Data: 4,
	})
	broker.Send(msgC, &actor.Message{
		Name: msgC,
		Data: 5,
	})
	broker.Send(msgA, &actor.Message{
		Name: msgA,
		Data: 6,
	})
	time.Sleep(time.Second * 3)
}

func TestSequencer2(t *testing.T) {
	log.InitLog("testing.log", "./log.toml", "testing", "testing", 0)

	broker := streamer.NewStatefulBroker()
	workerBase := &actor.FSMController{}
	workerBase.EndWith(actor.AsSequencer(NewWorker(), []string{msgA, msgB, msgC}))
	workerActor := actor.NewActor(
		"worker",
		broker,
		[]string{
			msgA,
			msgB,
			msgC,
		},
		[]string{},
		[]int{},
		workerBase,
	)
	workerActor.Connect(streamer.NewDisjunctions(workerActor, 1))

	sender := streamer.NewDefaultProducer(
		"sender",
		[]string{
			msgA,
			msgB,
			msgC,
		},
		[]int{1, 1, 1},
	)
	broker.RegisterProducer(sender)

	broker.Serve()

	workerBase.OnStart()

	broker.Send(msgC, &actor.Message{Name: msgC, Data: 1})
	broker.Send(msgC, &actor.Message{Name: msgC, Data: 2})
	broker.Send(msgC, &actor.Message{Name: msgC, Data: 3})
	time.Sleep(time.Second)
	broker.Send(msgB, &actor.Message{Name: msgB, Data: 4})
	broker.Send(msgB, &actor.Message{Name: msgB, Data: 5})
	broker.Send(msgA, &actor.Message{Name: msgA, Data: 6})
	time.Sleep(time.Second)
	broker.Send(msgB, &actor.Message{Name: msgB, Data: 7})
	broker.Send(msgA, &actor.Message{Name: msgA, Data: 8})
	broker.Send(msgA, &actor.Message{Name: msgA, Data: 9})
	time.Sleep(time.Second * 3)
}
