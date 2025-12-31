package actor

import (
	"sort"

	scommon "github.com/arcology-network/streamer/common"
)

// ===== msg cache =====
type MsgBuffer struct {
	msgs []*scommon.Message
}

func NewMsgBuffer() *MsgBuffer {
	return &MsgBuffer{
		msgs: make([]*scommon.Message, 0),
	}
}

/*
Put: Insert messages while maintaining order
Ordering rules:
Messages with smaller Height come first
For messages with the same Height, the one that arrives earlier comes first
/
*/
func (mb *MsgBuffer) Put(msg *scommon.Message) {
	mb.msgs = append(mb.msgs, msg)
	sort.SliceStable(mb.msgs, func(i, j int) bool {
		return mb.msgs[i].Height < mb.msgs[j].Height
	})
}

/*
PopByNames：
Find and remove the first message where name ∈ names, in order
*/
func (mb *MsgBuffer) PopByNames(names []string) *scommon.Message {
	nameSet := make(map[string]struct{}, len(names))
	for _, n := range names {
		nameSet[n] = struct{}{}
	}

	for i, msg := range mb.msgs {
		if _, ok := nameSet[msg.Name]; ok {
			return mb.popAt(i)
		}
	}
	return nil
}

/*
PopByHeight：
Find and remove the first message where msg.Height <= height
*/
func (mb *MsgBuffer) PopByHeight(height uint64) *scommon.Message {
	for i, msg := range mb.msgs {
		if msg.Height <= height {
			return mb.popAt(i)
		}
	}
	return nil
}

/*
PopIf：
General interface, pop the first message that meets the condition
*/
func (mb *MsgBuffer) PopIf(pred func(*scommon.Message) bool) *scommon.Message {
	for i, msg := range mb.msgs {
		if pred(msg) {
			return mb.popAt(i)
		}
	}
	return nil
}

/*
FlushByHeight：
Return and remove all messages where msg.Height <= height
*/
func (mb *MsgBuffer) FlushByHeight(height uint64) []*scommon.Message {
	var released []*scommon.Message
	remaining := mb.msgs[:0]
	for _, msg := range mb.msgs {
		if msg.Height <= height {
			released = append(released, msg)
		} else {
			remaining = append(remaining, msg)
		}
	}
	mb.msgs = remaining
	return released
}

/*
FlushAll：
Return all messages and clear the buffer
*/
func (mb *MsgBuffer) FlushAll() []*scommon.Message {
	out := mb.msgs
	mb.msgs = nil
	return out
}

// Internal helper function: Pop by index
func (mb *MsgBuffer) popAt(idx int) *scommon.Message {
	msg := mb.msgs[idx]
	mb.msgs = append(mb.msgs[:idx], mb.msgs[idx+1:]...)
	return msg
}

// Current cache length
func (mb *MsgBuffer) Len() int {
	return len(mb.msgs)
}
