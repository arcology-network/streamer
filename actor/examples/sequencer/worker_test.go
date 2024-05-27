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

package sequencer

import (
	"testing"
	"time"

	"github.com/arcology-network/streamer/actor"
	brokerpk "github.com/arcology-network/streamer/broker"
	"github.com/arcology-network/streamer/log"
)

func TestSequencer(t *testing.T) {
	log.InitLog("testing.log", "./log.toml", "testing", "testing", 0)

	broker := brokerpk.NewStatefulStreamer()
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
	workerActor.Connect(brokerpk.NewDisjunctions(workerActor, 1))

	sender := brokerpk.NewDefaultProducer(
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

	broker := brokerpk.NewStatefulStreamer()
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
	workerActor.Connect(brokerpk.NewDisjunctions(workerActor, 1))

	sender := brokerpk.NewDefaultProducer(
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
