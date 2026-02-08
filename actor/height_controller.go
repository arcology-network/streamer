package actor

import (
	"context"

	scommon "github.com/arcology-network/streamer/common"
	"github.com/arcology-network/streamer/logger"
)

type HeightSensitive interface {
	Height() uint64
}

type HeightController struct {
	client       HeightSensitive
	buf          *MsgBuffer
	businessName string
}

func NewHeightController(client HeightSensitive, businessName string) *HeightController {
	return &HeightController{
		client:       client,
		buf:          NewMsgBuffer(),
		businessName: businessName,
	}
}

func (c *HeightController) OnMessage(msg *scommon.Message) ([]*scommon.Message, error) {
	if msg.Height <= c.client.Height() {
		return []*scommon.Message{msg}, nil
	}

	c.buf.Put(msg)

	logger.Log.Debug(context.Background(), c.businessName, "HC: Push", logger.F("msg", msg.Name), logger.F("msgheight", msg.Height))
	return nil, nil
}

// Attempt to release messages in the cache that meet the height requirement
func (c *HeightController) OnAfterExecute() ([]*scommon.Message, error) {
	height := c.client.Height()
	logger.Log.Debug(context.Background(), c.businessName, "HC Status", logger.F("height", height))
	if msg := c.buf.PopByHeight(height); msg != nil {

		logger.Log.Debug(context.Background(), c.businessName, "HC: Pop", logger.F("msg", msg.Name), logger.F("msgheight", msg.Height))
		return []*scommon.Message{msg}, nil
	}
	return nil, nil
}
