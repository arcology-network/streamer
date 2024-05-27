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
	"github.com/arcology-network/streamer/log"
)

type Copyable interface {
	Clone() interface{}
}

type MessageWrapper struct {
	MsgBroker      *broker.StatefulStreamer
	LatestMessage  *Message
	WorkThreadName string
}

func (wrapper *MessageWrapper) Send(name string, data interface{}, args ...uint64) {
	msgId := log.Logger.GetLogId()
	sendHeight := wrapper.LatestMessage.Height
	if len(args) > 0 {
		sendHeight = args[0]
	}
	if len(args) > 1 {
		msgId = args[1]
	}

	msg := Message{
		From:   wrapper.WorkThreadName,
		Msgid:  msgId,
		Name:   name,
		Height: sendHeight,
		Round:  wrapper.LatestMessage.Round,
		Data:   data,
	}
	wrapper.MsgBroker.Send(name, &msg)

	log.Logger.AddLog(
		msgId,
		log.LogLevel_Info,
		name,
		wrapper.WorkThreadName,
		"send msg "+name,
		"msg",
		sendHeight,
		wrapper.LatestMessage.Round,
		wrapper.LatestMessage.Msgid,
		0)
}

func (wrapper *MessageWrapper) SendCopy(name string, data Copyable, height ...uint64) {
	msgId := log.Logger.GetLogId()
	sendHeight := wrapper.LatestMessage.Height
	if len(height) > 0 {
		sendHeight = height[0]
	}
	msg := Message{
		From:   wrapper.WorkThreadName,
		Msgid:  msgId,
		Name:   name,
		Height: sendHeight,
		Round:  wrapper.LatestMessage.Round,
		Data:   data.Clone(),
	}
	wrapper.MsgBroker.Send(name, &msg)

	log.Logger.AddLog(
		msgId,
		log.LogLevel_Info,
		name,
		wrapper.WorkThreadName,
		"send msg copy "+name,
		"msg",
		sendHeight,
		wrapper.LatestMessage.Round,
		wrapper.LatestMessage.Msgid,
		0)
}

// Shared message can be read, but cannot be taken over.
// The message is available during the block, but it is very likely to be overwritten or destroyed after the block.
func (wrapper *MessageWrapper) Share(name string, data interface{}, height ...uint64) {
	msgId := log.Logger.GetLogId()
	sendHeight := wrapper.LatestMessage.Height
	if len(height) > 0 {
		sendHeight = height[0]
	}
	msg := Message{
		From:       wrapper.WorkThreadName,
		Msgid:      msgId,
		Name:       name,
		Height:     sendHeight,
		Round:      wrapper.LatestMessage.Round,
		Data:       data,
		isReadOnly: true,
		owner:      wrapper.WorkThreadName,
	}
	wrapper.MsgBroker.Send(name, &msg)

	log.Logger.AddLog(
		msgId,
		log.LogLevel_Info,
		name,
		wrapper.WorkThreadName,
		"share msg "+name,
		"msg",
		sendHeight,
		wrapper.LatestMessage.Round,
		wrapper.LatestMessage.Msgid,
		0)
}

// Hand over message is no longer held by the sender after this function call.
// Consumers can either read or take the message over.
// Multiple readers can access the message simultaneously, but if any of the consumer take the message over,
// no more readers can access the message after that.
// Similarly, if the message has been read by someone, it cannot be taken over.
func (wrapper *MessageWrapper) HandOver(name string, data interface{}, height ...uint64) {
	msgId := log.Logger.GetLogId()
	sendHeight := wrapper.LatestMessage.Height
	if len(height) > 0 {
		sendHeight = height[0]
	}
	msg := Message{
		From:       wrapper.WorkThreadName,
		Msgid:      msgId,
		Name:       name,
		Height:     sendHeight,
		Round:      wrapper.LatestMessage.Round,
		Data:       data,
		isReadOnly: false,
	}
	wrapper.MsgBroker.Send(name, &msg)

	log.Logger.AddLog(
		msgId,
		log.LogLevel_Info,
		name,
		wrapper.WorkThreadName,
		"hand over msg "+name,
		"msg",
		sendHeight,
		wrapper.LatestMessage.Round,
		wrapper.LatestMessage.Msgid,
		0)
}
