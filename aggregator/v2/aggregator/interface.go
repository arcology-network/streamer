package aggregator

import (
	"github.com/arcology-network/streamer/actor"
)

type MsgBroker interface {
	Send(string, *actor.Message) error
}
