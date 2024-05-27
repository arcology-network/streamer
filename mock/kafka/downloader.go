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
	"fmt"
	"testing"

	"github.com/arcology-network/streamer/actor"
	"github.com/arcology-network/streamer/broker"
)

// Download implements IWorker interface.
type Downloader struct {
	messageTypes []string
	broker       *actor.MessageWrapper
}

func NewDownloader(_ int, _ string, _, messageTypes []string, _ string) actor.IWorkerEx {
	t.Log("NewDownloader")
	return &Downloader{
		messageTypes: messageTypes,
	}
}

func (d *Downloader) Init(wtName string, broker *broker.StatefulStreamer) {
	t.Log("Downloader.Init")
	d.broker = &actor.MessageWrapper{
		MsgBroker:      broker,
		LatestMessage:  actor.NewMessage(),
		WorkThreadName: wtName,
	}
}

func (d *Downloader) ChangeEnvironment(_ *actor.Message) {

}

func (d *Downloader) Inputs() ([]string, bool) {
	return []string{}, false
}

func (d *Downloader) Outputs() map[string]int {
	outputs := make(map[string]int)
	for _, msg := range d.messageTypes {
		outputs[msg] = 100
	}
	return outputs
}

func (d *Downloader) OnStart() {
	t.Log("Downloader.OnStart")
}

func (d *Downloader) OnMessageArrived(_ []*actor.Message) error {
	return nil
}

// Receive is used for testing.
func (d *Downloader) Receive(msg *actor.Message) {
	for _, typ := range d.messageTypes {
		if typ == msg.Name {
			d.broker.LatestMessage = msg
			d.broker.Send(msg.Name, msg.Data)
			return
		}
	}
	panic(fmt.Sprintf("unknown message type got: %v", msg.Name))
}

var t testing.TB

func NewDownloaderCreator(logger testing.TB) func(int, string, []string, []string, string) actor.IWorkerEx {
	t = logger
	return NewDownloader
}
