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

const (
	stateInit = iota
	stateCollecting
	stateDone
)

const (
	msgData  = "data"
	msgList  = "list"
	msgClear = "clear"
)

type AggrSelector struct {
	state    int
	height   uint64
	buf      []string
	waitings map[string]struct{}
}

func NewAggrSelector() *AggrSelector {
	return &AggrSelector{
		state: stateInit,
	}
}

func (aggr *AggrSelector) Inputs() ([]string, bool) {
	return []string{msgData, msgList, msgClear}, false
}

func (aggr *AggrSelector) Outputs() map[string]int {
	return map[string]int{}
}

func (aggr *AggrSelector) RegisterActions(reg actor.ActionRegistrar) {
	reg.Register(msgData, aggr.ReceivedData)
	reg.Register(msgList, aggr.ReceivedList)
	reg.Register(msgClear, aggr.Clear)
}

func (aggr *AggrSelector) ReceivedData(ctx *actor.ActionContext) error {
	msg := ctx.Messages[0]
	switch aggr.state {
	case stateInit:
		aggr.buf = append(aggr.buf, msg.Data.(string))
	case stateCollecting:
		delete(aggr.waitings, msg.Data.(string))
		if len(aggr.waitings) == 0 {
			aggr.state = stateDone
			fmt.Println("done.")
		}
	case stateDone:
		//
	}
	return nil
}
func (aggr *AggrSelector) ReceivedList(ctx *actor.ActionContext) error {
	msg := ctx.Messages[0]
	list := msg.Data.([]string)
	aggr.waitings = make(map[string]struct{})
	for _, elem := range list {
		aggr.waitings[elem] = struct{}{}
	}

	for _, elem := range aggr.buf {
		delete(aggr.waitings, elem)
	}

	if len(aggr.waitings) == 0 {
		aggr.state = stateDone
		fmt.Println("done.")
	} else {
		aggr.state = stateCollecting
		fmt.Println("collecting.")
	}
	return nil
}
func (aggr *AggrSelector) Clear(ctx *actor.ActionContext) error {
	aggr.state = stateInit
	aggr.height = ctx.Messages[0].Height + 1
	fmt.Println("cleared.")
	return nil
}

func (aggr *AggrSelector) GetFSMRules() map[int]actor.FSMRule {
	return map[int]actor.FSMRule{
		stateInit:       {Accept: []string{msgData, msgList}},
		stateCollecting: {Accept: []string{msgData}},
		stateDone:       {Accept: []string{msgData, msgClear}},
	}
}

func (aggr *AggrSelector) GetCurrentState() int {
	return aggr.state
}

func (aggr *AggrSelector) Height() uint64 {
	return aggr.height
}
