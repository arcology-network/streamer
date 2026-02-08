package jet

import (
	"github.com/arcology-network/streamer/actor"
	jetlib "github.com/arcology-network/streamer/jet/lib"
)

type KafkaUploader struct {
	stream *jetlib.JetKVStreamer
}

// return a Subscriber struct
func NewKafkaUploader(stream *jetlib.JetKVStreamer) actor.Business {
	uploader := KafkaUploader{}
	return &uploader
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

func (ku *KafkaUploader) RegisterActions(reg actor.ActionRegistrar) {
	inputs, _ := ku.Inputs()
	for _, input := range inputs {
		reg.Register(input, ku.OnMessageArrived)
	}

}
func (ku *KafkaUploader) OnMessageArrived(ctx *actor.ActionContext) error {
	msg := ctx.Messages[0]
	ku.stream.Send(msg.Name, msg)
	return nil
}
