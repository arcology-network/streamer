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
	height uint64
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

func (dpp *DataPreprocessor) RegisterActions(reg actor.ActionRegistrar) {
	reg.Register(msgData, dpp.ReceivedData)
}

func (dpp *DataPreprocessor) ReceivedData(ctx *actor.ActionContext) error {
	msg := ctx.Messages[0]
	if msg.Name == msgData {
		dpp.height = msg.Height
		fmt.Printf("DPP: %v\n", msg)
	}
	return nil
}
func (dpp *DataPreprocessor) Height() uint64 {
	return dpp.height
}

type DataPreprocessorV2 struct {
}

func NewDataPreprocessorV2() *DataPreprocessorV2 {
	return &DataPreprocessorV2{}
}
func (dpp *DataPreprocessorV2) RpcConfig() (string, int) {
	return "", 0
}
func (dpp *DataPreprocessorV2) Inputs() ([]string, bool) {
	return []string{msgData}, false
}

func (dpp *DataPreprocessorV2) Outputs() map[string]int {
	return map[string]int{}
}

func (dpp *DataPreprocessorV2) RegisterActions(reg actor.ActionRegistrar) {
	reg.Register(msgData, dpp.OnMessageArrived)
}

func (dpp *DataPreprocessorV2) OnMessageArrived(ctx *actor.ActionContext) error {
	msg := ctx.Messages[0]
	if msg.Name == msgData {
		fmt.Printf("DPP: %v\n", msg)
		ctx.ExecCtx.Send(msg.Name, msg.Data)
	}
	return nil
}
