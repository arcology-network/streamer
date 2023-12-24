package actor

import (
	"container/list"
)

type MsgBuffer struct {
	bufs    map[string]*list.List
	isMatch MatcherFunc
}

type MatcherFunc func(msg *Message, args ...interface{}) bool

func NewMsgBuffer(isMatch MatcherFunc) *MsgBuffer {
	return &MsgBuffer{
		bufs:    make(map[string]*list.List),
		isMatch: isMatch,
	}
}

func (mb *MsgBuffer) Put(msg *Message) {
	if buf, ok := mb.bufs[msg.Name]; ok {
		buf.PushBack(msg)
	} else {
		mb.bufs[msg.Name] = list.New()
		mb.bufs[msg.Name].PushBack(msg)
	}
}

func (mb *MsgBuffer) Get(args ...interface{}) *Message {
	for _, buf := range mb.bufs {
		if buf.Len() == 0 {
			continue
		}

		front := buf.Front()
		msg := front.Value.(*Message)
		if mb.isMatch(msg, args...) {
			buf.Remove(front)
			return msg
		}
	}
	return nil
}
