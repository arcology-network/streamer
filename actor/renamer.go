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
	brokerpk "github.com/arcology-network/streamer/broker"
)

const (
	RenamerPrefix = "__rename__"
)

func RenamerName(from, to string) string {
	return RenamerPrefix + from + "-" + to
}

type Renamer struct {
	WorkerThread
	from string
	to   string
}

func Rename(from string) *Renamer {
	return &Renamer{from: from}
}

func (r *Renamer) To(to string) *Renamer {
	r.to = to
	return r
}

func (r *Renamer) On(broker *brokerpk.StatefulStreamer) *Renamer {
	renamer := NewActorEx(
		RenamerName(r.from, r.to),
		broker,
		r,
	)
	renamer.Connect(brokerpk.NewDisjunctions(renamer, 1))
	return r
}

func (r *Renamer) Inputs() ([]string, bool) {
	return []string{r.from}, false
}

func (r *Renamer) Outputs() map[string]int {
	return map[string]int{
		r.to: 1,
	}
}

func (r *Renamer) OnStart() {

}

func (r *Renamer) OnMessageArrived(msgs []*Message) error {
	for _, v := range msgs {
		switch v.Name {
		case r.from:
			r.MsgBroker.Send(r.to, v.Data)
		}
	}

	return nil
}
