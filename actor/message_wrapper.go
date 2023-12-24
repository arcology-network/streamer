package actor

import (
	streamer "github.com/arcology-network/component-lib/broker"
	"github.com/arcology-network/component-lib/log"
)

type Copyable interface {
	Clone() interface{}
}

type MessageWrapper struct {
	MsgBroker      *streamer.StatefulBroker
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
