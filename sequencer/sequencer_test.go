package sequencer

import (
	"testing"
)

type MockMessage struct {
	name   string
	height uint64
}

func (this *MockMessage) Name() string {
	return this.name
}

func (this *MockMessage) Height() uint64 {
	return this.height
}

func TestHeightFilter(t *testing.T) {
	heightFilter := NewHeightFilter([]string{}, "msg_block_completed", 0)
	sequencer := NewSequencer(heightFilter)

	msg := &MockMessage{"eu_result", 0}
	if sequencer.Try(msg)[0].(*MockMessage).Height() != 0 || len(sequencer.buffer) != 0 {
		t.Error("Error: Wrong height", heightFilter.height)
	}

	passed := []interface{}{}
	for i := 0; i < 2; i++ {
		msg := &MockMessage{"eu_result", 2}
		passed = append(passed, sequencer.Try(msg)...)
	}

	for i := 0; i < 2; i++ {
		msg := &MockMessage{"eu_result", 1}
		passed = append(passed, sequencer.Try(msg)...)
	}

	for i := 0; i < 3; i++ {
		msg := &MockMessage{"whatever", 0}
		passed = append(passed, sequencer.Try(msg)...)
	}

	if len(passed) != 3 || passed[2].(MessageInterface).Name() != "whatever" || len(sequencer.buffer) != 4 {
		t.Error("Error: ", len(passed))
	}

	msg = &MockMessage{"msg_block_completed", 1}
	sequencer.Try(msg)

	if heightFilter.height != 1 {
		t.Error("Error: ", heightFilter.height)
	}

	if len(sequencer.buffer) != 2 || sequencer.buffer[0].(*MockMessage).Height() != 2 || sequencer.buffer[1].(*MockMessage).Height() != 2 {
		t.Error("Error: ", len(passed))
	}
}

func TestOrderFilter(t *testing.T) {
	orderFilter := NewOrderFilter([]string{"executed-tx", "inclusive-list", "roothash", "msg_block_completed"}, 0)
	sequencer := NewSequencer(orderFilter)

	msg := &MockMessage{"msg_block_completed", 1}
	if len(sequencer.Try(msg)) != 0 {
		t.Error("Error: msg_block_completed")
	}

	msg = &MockMessage{"roothash", 1}
	if len(sequencer.Try(msg)) != 0 {
		t.Error("Error: roothash")
	}

	msg = &MockMessage{"inclusive-list", 1}
	if len(sequencer.Try(msg)) != 0 {
		t.Error("Error: inclusivel-list")
	}

	msg = &MockMessage{"executed-tx", 1}
	seq := sequencer.Try(msg)
	if len(seq) != 4 {
		t.Error("Error: msg_block_completed")
	}

	if seq[0].(*MockMessage).Name() != "executed-tx" ||
		seq[1].(*MockMessage).Name() != "inclusive-list" ||
		seq[2].(*MockMessage).Name() != "roothash" ||
		seq[3].(*MockMessage).Name() != "msg_block_completed" {
		t.Error("Error: Wrong message sequence !!!")
	}
}

func TestHeightFilterMultipleMsgs(t *testing.T) {
	msgs := []interface{}{}

	msgs = append(msgs, &MockMessage{"roothash", 2})
	msgs = append(msgs, &MockMessage{"roothash", 3})
	msgs = append(msgs, &MockMessage{"roothash", 1})
	msgs = append(msgs, &MockMessage{"roothash", 0})

	msgs = append(msgs, &MockMessage{"executed-tx", 3})
	msgs = append(msgs, &MockMessage{"executed-tx", 2})
	msgs = append(msgs, &MockMessage{"executed-tx", 1})
	msgs = append(msgs, &MockMessage{"executed-tx", 0})

	msgs = append(msgs, &MockMessage{"inclusive-list", 1})
	msgs = append(msgs, &MockMessage{"inclusive-list", 3})
	msgs = append(msgs, &MockMessage{"inclusive-list", 2})
	msgs = append(msgs, &MockMessage{"inclusive-list", 0})

	HeightSeq := NewSequencer(NewHeightFilter([]string{}, "msg_block_completed", 0))

	output := []interface{}{}
	for i := 0; i < len(msgs); i++ {
		if outMsgs := HeightSeq.Try(msgs[i]); len(outMsgs) > 0 {
			output = append(output, outMsgs...)
		}
	}

	if len(output) != 3 {
		t.Error("Error: Wrong output messages !!!")
	}

	if outMsgs := HeightSeq.Try(&MockMessage{"msg_block_completed", 1}); len(outMsgs) > 0 {
		output = append(output, outMsgs...)
	}

	if len(output) != 7 || len(HeightSeq.buffer) != 6 {
		t.Error("Error: Wrong output messages !!!")
	}

	if outMsgs := HeightSeq.Try(&MockMessage{"msg_block_completed", 2}); len(outMsgs) > 0 {
		output = append(output, outMsgs...)
	}

	if len(output) != 11 || len(HeightSeq.buffer) != 3 {
		t.Error("Error: Wrong output messages !!!")
	}

	if outMsgs := HeightSeq.Try(&MockMessage{"msg_block_completed", 3}); len(outMsgs) > 0 {
		output = append(output, outMsgs...)
	}

	if len(output) != 15 || len(HeightSeq.buffer) != 0 {
		t.Error("Error: Wrong output messages !!!")
	}
}

func TestChainedFilters(t *testing.T) {
	HeightSeq := NewSequencer(NewHeightFilter([]string{}, "msg_block_completed", 0))
	orderFilter := NewSequencer(NewOrderFilter([]string{"executed-tx", "inclusive-list", "roothash", "msg_block_completed"}, 0))
	combinedSeq := Sequencers([]*Sequencer{HeightSeq, orderFilter})

	/* First batch */
	msgs := []interface{}{}
	msgs = append(msgs, &MockMessage{"roothash", 2})
	msgs = append(msgs, &MockMessage{"roothash", 3})
	msgs = append(msgs, &MockMessage{"roothash", 1})
	msgs = append(msgs, &MockMessage{"roothash", 0})

	out := combinedSeq.BatchTry(msgs)
	if len(out) != 0 {
		t.Error("Error: Wrong output messages, first batch !!!")
	}

	/* Second batch */
	msgs = msgs[:0]
	msgs = append(msgs, &MockMessage{"executed-tx", 3})
	msgs = append(msgs, &MockMessage{"executed-tx", 2})
	msgs = append(msgs, &MockMessage{"executed-tx", 1})
	msgs = append(msgs, &MockMessage{"executed-tx", 0})

	out = combinedSeq.BatchTry(msgs)
	if len(out) != 1 {
		t.Error("Error: Wrong output messages, second batch !!!")
	}

	/* Thrid batch */
	msgs = msgs[:0]
	msgs = append(msgs, &MockMessage{"inclusive-list", 1})
	msgs = append(msgs, &MockMessage{"inclusive-list", 3})
	msgs = append(msgs, &MockMessage{"inclusive-list", 2})
	msgs = append(msgs, &MockMessage{"inclusive-list", 0})

	out = combinedSeq.BatchTry(msgs)
	if len(out) != 2 || len(out) != 2 {
		t.Error("Error: Wrong output messages, third batch !!!")
	}
}
