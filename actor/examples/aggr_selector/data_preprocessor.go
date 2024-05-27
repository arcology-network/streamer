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
	"fmt"

	"github.com/arcology-network/streamer/actor"
)

type DataPreprocessor struct {
	actor.WorkerThread
}

func NewDataPreprocessor() *DataPreprocessor {
	return &DataPreprocessor{}
}

func (dpp *DataPreprocessor) Inputs() ([]string, bool) {
	return []string{msgData}, false
}

func (dpp *DataPreprocessor) Outputs() map[string]int {
	return map[string]int{}
}

func (dpp *DataPreprocessor) OnStart() {}

func (dpp *DataPreprocessor) OnMessageArrived(msgs []*actor.Message) error {
	msg := msgs[0]
	if msg.Name == msgData {
		fmt.Printf("DPP: %v\n", msg)
	}
	return nil
}

type DataPreprocessorV2 struct {
	actor.WorkerThread
}

func NewDataPreprocessorV2() *DataPreprocessorV2 {
	return &DataPreprocessorV2{}
}

func (dpp *DataPreprocessorV2) Inputs() ([]string, bool) {
	return []string{msgData}, false
}

func (dpp *DataPreprocessorV2) Outputs() map[string]int {
	return map[string]int{}
}

func (dpp *DataPreprocessorV2) OnStart() {}

func (dpp *DataPreprocessorV2) OnMessageArrived(msgs []*actor.Message) error {
	msg := msgs[0]
	if msg.Name == msgData {
		fmt.Printf("DPP: %v\n", msg)
		dpp.MsgBroker.Send(msg.Name, msg.Data)
	}
	return nil
}
