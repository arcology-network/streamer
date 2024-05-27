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
	"github.com/arcology-network/streamer/broker"
)

type sequencer struct {
	worker IWorkerEx

	defs  map[int][]string
	state int
}

func AsSequencer(worker IWorkerEx, sequence []string) *sequencer {
	defs := make(map[int][]string)
	for index, item := range sequence {
		defs[index] = []string{item}
	}

	return &sequencer{
		worker: worker,
		defs:   defs,
	}
}

func (s *sequencer) Init(wtName string, broker *broker.StatefulStreamer) {
	s.worker.Init(wtName, broker)
}

func (s *sequencer) ChangeEnvironment(message *Message) {
	s.worker.ChangeEnvironment(message)
}

func (s *sequencer) Inputs() ([]string, bool) {
	return s.worker.Inputs()
}

func (s *sequencer) Outputs() map[string]int {
	return s.worker.Outputs()
}

func (s *sequencer) Config(params map[string]interface{}) {
	if c, ok := s.worker.(Configurable); ok {
		c.Config(params)
	} else {
		panic("not configurable")
	}
}

func (s *sequencer) OnStart() {
	s.worker.OnStart()
}

func (s *sequencer) OnMessageArrived(msgs []*Message) error {
	if s.state == len(s.defs)-1 {
		s.state = 0
	} else {
		s.state++
	}
	return s.worker.OnMessageArrived(msgs)
}

func (s *sequencer) GetStateDefinitions() map[int][]string {
	return s.defs
}

func (s *sequencer) GetCurrentState() int {
	return s.state
}
