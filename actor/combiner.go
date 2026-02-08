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
	"sort"
	"strings"

	brokerpk "github.com/arcology-network/streamer/broker"
	scommon "github.com/arcology-network/streamer/common"
)

const (
	CombinerPrefix = "__combine__"
)

func CombinedName(inputs ...string) string {
	sort.Strings(inputs)
	return CombinerPrefix + strings.Join(inputs, "-")
}

type CombinerElements struct {
	Msgs map[string]*scommon.Message
}

func (cl *CombinerElements) Get(msgname string) *scommon.Message {
	if msg, ok := cl.Msgs[msgname]; ok {
		return msg
	} else {
		return nil
	}
}

type Combiner struct {
	inputs []string
	outMsg string
	els    *CombinerElements
}

func Combine(inputs ...string) *Combiner {
	return &Combiner{
		inputs: inputs,
		outMsg: CombinedName(inputs...),
		els: &CombinerElements{
			Msgs: map[string]*scommon.Message{},
		},
	}
}

func (c *Combiner) On(broker *brokerpk.StatefulStreamer) (Business, *Actor) {
	act := CreateActor(
		c.outMsg,
		broker,
		[]Business{c},
		[]string{"combiner"},
		// []*Filter{},
		2,
		[]string{""},
	)
	return c, act
}

func (c *Combiner) Inputs() ([]string, bool) {
	return c.inputs, true
}

func (c *Combiner) Outputs() map[string]int {
	return map[string]int{
		c.outMsg: 1,
	}
}

func (c *Combiner) PrimaryMsg() string {
	return c.inputs[0]
}

func (c *Combiner) Config(params map[string]interface{}) {

}

func (c *Combiner) RegisterActions(reg ActionRegistrar) {
	for i := range c.inputs {
		reg.Register(c.inputs[i], c.combine)
	}
}

func (c *Combiner) combine(ctx *ActionContext) error {
	for _, v := range ctx.Messages {
		c.els.Msgs[v.Name] = v
	}
	ctx.ExecCtx.Send(c.outMsg, c.els)
	c.els = &CombinerElements{
		Msgs: map[string]*scommon.Message{},
	}
	return nil
}
