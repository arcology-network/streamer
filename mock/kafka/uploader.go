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

package kafka

import (
	"testing"

	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/streamer/actor"
	"github.com/arcology-network/streamer/broker"
	scommon "github.com/arcology-network/streamer/common"
)

type OnMessageArrivedCallback func(msgs []*scommon.Message) error

type Uploader struct {
	messages        map[string]string
	broker          *actor.MessageWrapper
	counter         map[string]int
	encChan         chan *scommon.Message
	dataSizeCounter map[string]int
	cb              OnMessageArrivedCallback
}

func NewUploader(_ int, _ string, messages map[string]string, _ string) actor.IWorkerEx {
	tu.Log("NewUploader")
	return &Uploader{
		messages:        messages,
		counter:         make(map[string]int),
		encChan:         make(chan *scommon.Message, 1000),
		dataSizeCounter: make(map[string]int),
	}
}

func (u *Uploader) SetCallback(cb OnMessageArrivedCallback) {
	u.cb = cb
}

func (u *Uploader) Init(owner actor.IWorker, wtName string, broker *broker.StatefulStreamer) *actor.ActionRouter {
	tu.Log("Uploader.Init")
	u.broker = &actor.MessageWrapper{
		MsgBroker:      broker,
		WorkThreadName: wtName,
	}
	return nil
}

func (u *Uploader) ChangeEnvironment(_ *scommon.Message) {

}

func (u *Uploader) Inputs() ([]string, bool) {
	var inputs []string
	for input := range u.messages {
		inputs = append(inputs, input)
	}
	return inputs, false
}

func (u *Uploader) Outputs() map[string]int {
	return map[string]int{}
}

func (u *Uploader) OnStart() {
	tu.Log("Uploader.OnStart")
	go func() {
		for msg := range u.encChan {
			bs, err := common.GobEncode(msg)
			if err != nil {
				panic(err)
			}
			u.dataSizeCounter[msg.Name] += len(bs)
		}
	}()
}

func (u *Uploader) OnMessageArrived(msgs []*scommon.Message) error {
	if u.cb != nil {
		u.cb(msgs)
	}

	for _, m := range msgs {
		u.counter[m.Name]++
		u.encChan <- m
	}
	return nil
}
func (u *Uploader) ActionRegister() {

}
func (u *Uploader) RpcConfig() (string, int) {
	return "", 0
}

// For test.
func (u *Uploader) GetCounter() map[string]int {
	return u.counter
}

func (u *Uploader) GetDataSizeCounter() map[string]int {
	return u.dataSizeCounter
}

var tu testing.TB

func NewUploaderCreator(logger testing.TB) func(int, string, map[string]string, string) actor.IWorkerEx {
	tu = logger
	return NewUploader
}
