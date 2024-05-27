package kafka

import (
	"strings"

	"github.com/arcology-network/streamer/actor"
	"github.com/arcology-network/streamer/kafka/lib"
)

type KafkaUploader struct {
	actor.WorkerThread
	outgoing  *lib.ComOutgoing
	relations map[string]string
	mqaddr    string
}

// return a Subscriber struct
func NewKafkaUploader(concurrency int, groupid string, relations map[string]string, mqaddr string) actor.IWorkerEx {
	uploader := KafkaUploader{}
	uploader.Set(concurrency, groupid)
	uploader.relations = relations
	uploader.mqaddr = mqaddr
	return &uploader
}

func (ku *KafkaUploader) Inputs() ([]string, bool) {
	var inputs []string
	for input := range ku.relations {
		inputs = append(inputs, input)
	}
	return inputs, false
}

func (ku *KafkaUploader) Outputs() map[string]int {
	return map[string]int{}
}

func (ku *KafkaUploader) OnStart() {
	ku.outgoing = new(lib.ComOutgoing)
	if err := ku.outgoing.Start(strings.Split(ku.mqaddr, ","), ku.relations, ku.Name); err != nil {
		panic(err)
	}
}

func (ku *KafkaUploader) OnMessageArrived(msgs []*actor.Message) error {
	ku.outgoing.Send(msgs[0])
	return nil
}
