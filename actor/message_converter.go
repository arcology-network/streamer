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

package actor

import (
	"fmt"

	brokerpk "github.com/arcology-network/streamer/broker"
	"github.com/arcology-network/streamer/log"
)

type Actor struct {
	receiver chan interface{}

	Producer    brokerpk.StreamProducer
	name        string
	subscribeTo []string
	workerThd   IWorker
	broker      *brokerpk.StatefulStreamer
}

func NewActor(name string, broker *brokerpk.StatefulStreamer, subscribeTo []string, publishTo []string, bufferLen []int, workerThd IWorker) *Actor {
	actor := &Actor{
		receiver:    make(chan interface{}),
		name:        name,
		subscribeTo: subscribeTo,
		broker:      broker,
	}
	if publishTo != nil || len(publishTo) > 0 {
		actor.Producer = brokerpk.NewDefaultProducer(name+"-producer", publishTo, bufferLen)
	}

	actor.SetWorker(workerThd)
	workerThd.Init(name, broker)
	// workerThd.OnStart()

	for _, subscribe := range subscribeTo {
		log.Metas.Add(name, "sm_"+subscribe, false)
	}
	for _, publish := range publishTo {
		log.Metas.Add(name, "sm_"+publish, true)
	}

	go actor.Serve()
	return actor
}

func NewActorEx(name string, broker *brokerpk.StatefulStreamer, worker IWorkerEx) *Actor {
	var publishTo []string
	var bufferLen []int
	for output, bufferSize := range worker.Outputs() {
		publishTo = append(publishTo, output)
		bufferLen = append(bufferLen, bufferSize)
	}
	inputs, _ := worker.Inputs()
	return NewActor(name, broker, inputs, publishTo, bufferLen, worker)
}

func (actor *Actor) Connect(controller brokerpk.StreamController) {
	if actor.Producer != nil {
		actor.broker.RegisterProducer(actor.Producer)
	}

	actor.broker.RegisterConsumer(brokerpk.NewDefaultConsumer(actor.name+"-consumer", actor.subscribeTo, controller))
}

// SetWorkUint set one uint for work
func (actor *Actor) SetWorker(worker IWorker) {
	actor.workerThd = worker
}

func (actor *Actor) Consume(data interface{}) {
	actor.receiver <- data
}

func (actor *Actor) Serve() {
	for v := range actor.receiver {
		msgs := []*Message{}
		switch v.(type) {
		case []interface{}:
			params := v.([]interface{})
			for _, p := range params {
				actor.parseParams(&msgs, p)
			}
		case brokerpk.Aggregated:
			param := v.(brokerpk.Aggregated)
			actor.parseParams(&msgs, param.Data)
		}

		idx := actor.findMaxHeight(&msgs)
		refid := actor.log(&msgs)
		lstMessage := msgs[idx].CopyHeader()
		lstMessage.Msgid = refid
		actor.workerThd.ChangeEnvironment(lstMessage)
		actor.workerThd.OnMessageArrived(msgs)
	}

}

func (actor *Actor) log(msgs *[]*Message) uint64 {
	workthreadname := actor.name
	var latestMsg *Message
	source := ""
	for _, v := range *msgs {
		latestMsg = v
		log.Logger.AddLog(
			0,
			log.LogLevel_Info,
			v.Name,
			workthreadname,
			"received msg "+v.Name,
			"msg",
			v.Height,
			v.Round,
			v.Msgid,
			0,
		)
		source = source + "," + v.Name
	}

	interRefId := log.Logger.AddLog(
		0,
		log.LogLevel_Info,
		source[1:],
		workthreadname,
		fmt.Sprintf("%d messages enter workthread %v", len(*msgs), workthreadname),
		log.LogType_Act,
		latestMsg.Height,
		latestMsg.Round,
		latestMsg.Msgid,
		0,
	)
	return interRefId
}

func (actor *Actor) findMaxHeight(msgs *[]*Message) int {
	heightPlus := uint64(0)
	roundPlus := uint64(0)
	idx := 0
	for i, msg := range *msgs {
		if msg.Height > heightPlus || (msg.Height == heightPlus && msg.Round > roundPlus) {
			heightPlus = msg.Height
			roundPlus = msg.Round
			idx = i
		}
	}
	return idx
}

func (actor *Actor) parseParams(msgs *[]*Message, data interface{}) {
	switch data.(type) {
	case []interface{}:
		pparams := data.([]interface{})
		for _, pp := range pparams {
			*msgs = append(*msgs, pp.(*Message))
		}
	case brokerpk.Aggregated:
		param := data.(brokerpk.Aggregated)
		switch param.Data.(type) {
		case brokerpk.Aggregated:
			pparam := param.Data.(brokerpk.Aggregated)

			*msgs = append(*msgs, pparam.Data.(*Message))
		case []interface{}:
			pparams := param.Data.([]interface{})
			for _, pp := range pparams {
				*msgs = append(*msgs, pp.(*Message))
			}
		}
	case *Message:
		*msgs = append(*msgs, data.(*Message))
	}
}
