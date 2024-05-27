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

package aggrselector

import (
	"testing"
	"time"

	"github.com/arcology-network/streamer/actor"
	brokerpk "github.com/arcology-network/streamer/broker"
	"github.com/arcology-network/streamer/log"
)

func TestAggrSelector(t *testing.T) {
	log.InitLog("testing.log", "./log.toml", "testing", "testing", 0)

	broker := brokerpk.NewStatefulStreamer()
	aggrSelectorBase := &actor.HeightController{}
	aggrSelectorBase.Next(&actor.FSMController{}).Next(actor.MakeLinkable(NewDataPreprocessor())).EndWith(NewAggrSelector())
	aggrSelectorActor := actor.NewActor(
		"aggrselector",
		broker,
		[]string{
			msgData,
			msgList,
			msgClear,
		},
		[]string{},
		[]int{},
		aggrSelectorBase,
	)
	aggrSelectorActor.Connect(brokerpk.NewDisjunctions(aggrSelectorActor, 1))

	sender := brokerpk.NewDefaultProducer(
		"sender",
		[]string{
			msgData,
			msgList,
			msgClear,
		},
		[]int{1, 1, 1},
	)
	broker.RegisterProducer(sender)

	broker.Serve()

	aggrSelectorBase.OnStart()

	broker.Send(msgData, &actor.Message{
		Name: msgData,
		Data: "item1",
	})
	broker.Send(msgClear, &actor.Message{
		Name: msgClear,
	})
	broker.Send(msgList, &actor.Message{
		Name: msgList,
		Data: []string{"item1", "item2"},
	})
	broker.Send(msgData, &actor.Message{
		Name:   msgData,
		Data:   "item2",
		Height: 1,
	})
	broker.Send(msgData, &actor.Message{
		Name:   msgData,
		Data:   "item2",
		Height: 2,
	})
	broker.Send(msgData, &actor.Message{
		Name: msgData,
		Data: "item2",
	})
	time.Sleep(time.Second * 3)
}

func TestAggrSelector2(t *testing.T) {
	log.InitLog("testing.log", "./log.toml", "testing", "testing", 0)

	broker := brokerpk.NewStatefulStreamer()
	aggrSelectorBase := &actor.HeightController{}
	aggrSelectorBase.Next(&actor.FSMController{}).Next(actor.MakeLinkable(NewDataPreprocessor())).EndWith(NewAggrSelector())
	aggrSelectorActor := actor.NewActor(
		"aggrselector",
		broker,
		[]string{
			msgData,
			msgList,
			msgClear,
		},
		[]string{},
		[]int{},
		aggrSelectorBase,
	)
	aggrSelectorActor.Connect(brokerpk.NewDisjunctions(aggrSelectorActor, 1))

	sender := brokerpk.NewDefaultProducer(
		"sender",
		[]string{
			msgData,
			msgList,
			msgClear,
		},
		[]int{1, 1, 1},
	)
	broker.RegisterProducer(sender)

	broker.Serve()

	aggrSelectorBase.OnStart()

	broker.Send(msgData, &actor.Message{Name: msgData, Data: "item1", Height: 0})
	broker.Send(msgList, &actor.Message{Name: msgList, Data: []string{"item1", "item2"}, Height: 0})
	broker.Send(msgClear, &actor.Message{Name: msgClear, Height: 0})
	broker.Send(msgClear, &actor.Message{Name: msgClear, Height: 1})
	broker.Send(msgClear, &actor.Message{Name: msgClear, Height: 2})
	time.Sleep(time.Second)
	broker.Send(msgList, &actor.Message{Name: msgList, Data: []string{"item3", "item4"}, Height: 1})
	broker.Send(msgData, &actor.Message{Name: msgData, Data: "item2", Height: 0})
	broker.Send(msgData, &actor.Message{Name: msgData, Data: "item3", Height: 1})
	broker.Send(msgData, &actor.Message{Name: msgData, Data: "item4", Height: 1})
	broker.Send(msgData, &actor.Message{Name: msgData, Data: "garbage", Height: 1})
	broker.Send(msgData, &actor.Message{Name: msgData, Data: "garbage", Height: 2})
	time.Sleep(time.Second)
	broker.Send(msgData, &actor.Message{Name: msgData, Data: "item5", Height: 2})
	broker.Send(msgData, &actor.Message{Name: msgData, Data: "item6", Height: 2})
	broker.Send(msgData, &actor.Message{Name: msgList, Data: []string{"item5", "item6"}, Height: 2})
	time.Sleep(time.Second * 3)
}

func TestAggrSelectorWithMsgCleaner(t *testing.T) {
	log.InitLog("testing.log", "./log.toml", "testing", "testing", 0)

	broker := brokerpk.NewStatefulStreamer()
	aggrSelectorBase := actor.NewMsgCleaner(func(msg *actor.Message) bool {
		return msg.Name != msgData || msg.From != "sender"
	})
	aggrSelectorBase.Next(&actor.HeightController{}).Next(&actor.FSMController{}).EndWith(NewAggrSelector())
	aggrSelectorActor := actor.NewActor(
		"aggrselector",
		broker,
		[]string{
			msgData,
			msgList,
			msgClear,
		},
		[]string{},
		[]int{},
		aggrSelectorBase,
	)
	aggrSelectorActor.Connect(brokerpk.NewDisjunctions(aggrSelectorActor, 1))

	dataPreprocessorBase := actor.NewMsgCleaner(actor.OnlyFrom("sender"))
	dataPreprocessorBase.EndWith(NewDataPreprocessorV2())
	dataPreprocessorActor := actor.NewActor(
		"datapreprocessor",
		broker,
		[]string{msgData},
		[]string{msgData},
		[]int{1},
		dataPreprocessorBase,
	)
	dataPreprocessorActor.Connect(brokerpk.NewDisjunctions(dataPreprocessorActor, 1))

	sender := brokerpk.NewDefaultProducer(
		"sender",
		[]string{
			msgData,
			msgList,
			msgClear,
		},
		[]int{1, 1, 1},
	)
	broker.RegisterProducer(sender)

	broker.Serve()

	aggrSelectorBase.OnStart()
	dataPreprocessorBase.OnStart()

	broker.Send(msgData, &actor.Message{From: "sender", Name: msgData, Data: "item1", Height: 0})
	broker.Send(msgList, &actor.Message{From: "sender", Name: msgList, Data: []string{"item1", "item2"}, Height: 0})
	broker.Send(msgClear, &actor.Message{From: "sender", Name: msgClear, Height: 0})
	broker.Send(msgClear, &actor.Message{From: "sender", Name: msgClear, Height: 1})
	broker.Send(msgClear, &actor.Message{From: "sender", Name: msgClear, Height: 2})
	time.Sleep(time.Second)
	broker.Send(msgList, &actor.Message{From: "sender", Name: msgList, Data: []string{"item3", "item4"}, Height: 1})
	broker.Send(msgData, &actor.Message{From: "sender", Name: msgData, Data: "item2", Height: 0})
	broker.Send(msgData, &actor.Message{From: "sender", Name: msgData, Data: "item3", Height: 1})
	broker.Send(msgData, &actor.Message{From: "sender", Name: msgData, Data: "item4", Height: 1})
	broker.Send(msgData, &actor.Message{From: "sender", Name: msgData, Data: "garbage", Height: 1})
	broker.Send(msgData, &actor.Message{From: "sender", Name: msgData, Data: "garbage", Height: 2})
	time.Sleep(time.Second)
	broker.Send(msgData, &actor.Message{From: "sender", Name: msgData, Data: "item5", Height: 2})
	broker.Send(msgData, &actor.Message{From: "sender", Name: msgData, Data: "item6", Height: 2})
	broker.Send(msgData, &actor.Message{From: "sender", Name: msgList, Data: []string{"item5", "item6"}, Height: 2})
	time.Sleep(time.Second * 3)
}
