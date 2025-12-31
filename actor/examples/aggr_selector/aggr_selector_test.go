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
	scommon "github.com/arcology-network/streamer/common"
)

func TestAggrSelector(t *testing.T) {
	// log.InitLog("testing.log", "./log.toml", "testing", "testing", 0)

	broker := brokerpk.NewStatefulStreamer()
	dataProcessor := NewDataPreprocessor()
	aggr := NewAggrSelector()
	actor.CreateActor("aggrselector", broker, []actor.Business{dataProcessor, aggr}, []string{"dataProcessor", "aggr"}, []*actor.Filter{}, 8)

	broker.Serve()

	broker.Send(msgData, &scommon.Message{
		Name: msgData,
		Data: "item1",
	})
	broker.Send(msgClear, &scommon.Message{
		Name: msgClear,
	})
	broker.Send(msgList, &scommon.Message{
		Name: msgList,
		Data: []string{"item1", "item2"},
	})
	broker.Send(msgData, &scommon.Message{
		Name:   msgData,
		Data:   "item2",
		Height: 1,
	})
	broker.Send(msgData, &scommon.Message{
		Name:   msgData,
		Data:   "item2",
		Height: 2,
	})
	broker.Send(msgData, &scommon.Message{
		Name: msgData,
		Data: "item2",
	})
	time.Sleep(time.Second * 3)
}

func TestAggrSelector2(t *testing.T) {

	broker := brokerpk.NewStatefulStreamer()
	dataProcessor := NewDataPreprocessor()
	aggr := NewAggrSelector()
	actor.CreateActor("aggrselector", broker, []actor.Business{dataProcessor, aggr}, []string{"dataProcessor", "aggr"}, []*actor.Filter{}, 8)

	broker.Serve()

	broker.Send(msgData, &scommon.Message{Name: msgData, Data: "item1", Height: 0})
	broker.Send(msgList, &scommon.Message{Name: msgList, Data: []string{"item1", "item2"}, Height: 0})
	broker.Send(msgClear, &scommon.Message{Name: msgClear, Height: 0})
	broker.Send(msgClear, &scommon.Message{Name: msgClear, Height: 1})
	broker.Send(msgClear, &scommon.Message{Name: msgClear, Height: 2})
	time.Sleep(time.Second)
	broker.Send(msgList, &scommon.Message{Name: msgList, Data: []string{"item3", "item4"}, Height: 1})
	broker.Send(msgData, &scommon.Message{Name: msgData, Data: "item2", Height: 0})
	broker.Send(msgData, &scommon.Message{Name: msgData, Data: "item3", Height: 1})
	broker.Send(msgData, &scommon.Message{Name: msgData, Data: "item4", Height: 1})
	broker.Send(msgData, &scommon.Message{Name: msgData, Data: "garbage", Height: 1})
	broker.Send(msgData, &scommon.Message{Name: msgData, Data: "garbage", Height: 2})
	time.Sleep(time.Second)
	broker.Send(msgData, &scommon.Message{Name: msgData, Data: "item5", Height: 2})
	broker.Send(msgData, &scommon.Message{Name: msgData, Data: "item6", Height: 2})
	broker.Send(msgData, &scommon.Message{Name: msgList, Data: []string{"item5", "item6"}, Height: 2})
	time.Sleep(time.Second * 3)
}

func TestAggrSelectorWithMsgCleaner(t *testing.T) {
	broker := brokerpk.NewStatefulStreamer()
	dataProcessor := NewDataPreprocessor()
	aggr := NewAggrSelector()
	actor.CreateActor("aggrselector", broker, []actor.Business{dataProcessor, aggr}, []string{"dataProcessor", "aggr"}, []*actor.Filter{}, 8)

	filter := actor.NewOriginFilter(
		"origin-filter",
		func(msg *scommon.Message) bool {
			return msg.Name != msgData || msg.From != "sender"
		},
	)
	actor.CreateActor("aggrselector", broker, []actor.Business{aggr}, []string{"aggr"}, []*actor.Filter{filter}, 8)

	dataProcessor2 := NewDataPreprocessorV2()
	filter1 := actor.NewOriginFilter(
		"origin-filter1",
		actor.OnlyFrom("sender"),
	)
	actor.CreateActor("datapreprocessor", broker, []actor.Business{dataProcessor2}, []string{"dataProcessor2"}, []*actor.Filter{filter1}, 8)

	broker.Serve()

	broker.Send(msgData, &scommon.Message{From: "sender", Name: msgData, Data: "item1", Height: 0})
	broker.Send(msgList, &scommon.Message{From: "sender", Name: msgList, Data: []string{"item1", "item2"}, Height: 0})
	broker.Send(msgClear, &scommon.Message{From: "sender", Name: msgClear, Height: 0})
	broker.Send(msgClear, &scommon.Message{From: "sender", Name: msgClear, Height: 1})
	broker.Send(msgClear, &scommon.Message{From: "sender", Name: msgClear, Height: 2})
	time.Sleep(time.Second)
	broker.Send(msgList, &scommon.Message{From: "sender", Name: msgList, Data: []string{"item3", "item4"}, Height: 1})
	broker.Send(msgData, &scommon.Message{From: "sender", Name: msgData, Data: "item2", Height: 0})
	broker.Send(msgData, &scommon.Message{From: "sender", Name: msgData, Data: "item3", Height: 1})
	broker.Send(msgData, &scommon.Message{From: "sender", Name: msgData, Data: "item4", Height: 1})
	broker.Send(msgData, &scommon.Message{From: "sender", Name: msgData, Data: "garbage", Height: 1})
	broker.Send(msgData, &scommon.Message{From: "sender", Name: msgData, Data: "garbage", Height: 2})
	time.Sleep(time.Second)
	broker.Send(msgData, &scommon.Message{From: "sender", Name: msgData, Data: "item5", Height: 2})
	broker.Send(msgData, &scommon.Message{From: "sender", Name: msgData, Data: "item6", Height: 2})
	broker.Send(msgData, &scommon.Message{From: "sender", Name: msgList, Data: []string{"item5", "item6"}, Height: 2})
	time.Sleep(time.Second * 3)
}
