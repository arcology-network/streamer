package actor

import (
	"fmt"

	scommon "github.com/arcology-network/streamer/common"
)

type HeightSensitive interface {
	Height() uint64
}

type HeightController struct {
	client HeightSensitive
	buf    *MsgBuffer
}

func NewHeightController(client HeightSensitive) *HeightController {
	return &HeightController{
		client: client,
		buf:    NewMsgBuffer(),
	}
}

func (c *HeightController) OnMessage(msg *scommon.Message) ([]*scommon.Message, error) {
	if msg.Height <= c.client.Height() {
		return []*scommon.Message{msg}, nil
	}

	c.buf.Put(msg)
	fmt.Printf("HC: Push name=%s height=%d\n", msg.Name, msg.Height)
	return nil, nil
}

// Attempt to release messages in the cache that meet the height requirement
func (c *HeightController) OnAfterExecute() ([]*scommon.Message, error) {
	if msg := c.buf.PopByHeight(c.client.Height()); msg != nil {
		fmt.Printf("HC: Pop name=%s height=%d\n", msg.Name, msg.Height)
		return []*scommon.Message{msg}, nil
	}
	return nil, nil
}
