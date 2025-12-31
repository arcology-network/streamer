package jet

import (
	"github.com/arcology-network/streamer/actor"
	jetlib "github.com/arcology-network/streamer/jet/lib"
)

type KafkaUploader struct {
	stream *jetlib.JetKVStreamer
}

// return a Subscriber struct
func NewKafkaUploader(stream *jetlib.JetKVStreamer) *KafkaUploader {
	uploader := KafkaUploader{}
	return &uploader
}
func (ku *KafkaUploader) RpcConfig() (string, int) {
	return "", 0
}
func (ku *KafkaUploader) Inputs() ([]string, bool) {
	baseTopics := ku.stream.GetConfig().TopicsU
	inputs := make([]string, len(baseTopics))
	for i := range baseTopics {
		d_topic := jetlib.TopicForDown(baseTopics[i])
		inputs[i] = d_topic
	}
	return inputs, false
}

func (ku *KafkaUploader) Outputs() map[string]int {
	return map[string]int{}
}

func (ku *KafkaUploader) OnStart() {
}

func (ku *KafkaUploader) OnMessageArrived(msgs []*actor.Message) error {
	// ku.outgoing.Send(msgs[0])
	// ku.stream.Send("", msgs[0])
	return nil
}
