package kafka

import (
	"strings"

	"github.com/arcology-network/streamer/actor"
	"github.com/arcology-network/streamer/kafka/v2/lib"
	"github.com/arcology-network/streamer/log"

	"github.com/spf13/viper"
)

type KafkaDownloader struct {
	actor.WorkerThread

	inComing     *lib.ComIncoming
	topics       []string
	messageTypes []string
}

// return a Subscriber struct
func NewKafkaDownloader(concurrency int, groupid string, topics, messageTypes []string) *KafkaDownloader {
	downloader := KafkaDownloader{}
	downloader.Set(concurrency, groupid)
	downloader.topics = topics
	downloader.messageTypes = messageTypes
	return &downloader
}

func (kd *KafkaDownloader) OnStart() {
	kd.AddLog(log.LogLevel_Info, "start exec subscriber ")

	mqaddr := viper.GetString("mqaddr")
	kd.inComing = new(lib.ComIncoming)
	kd.inComing.Start(strings.Split(mqaddr, ","), kd.topics, kd.messageTypes, kd.Groupid, kd.Name, kd.onKafkaMessageArrived)

}

func (kd *KafkaDownloader) OnMessageArrived(msgs []*actor.Message) error {
	return nil
}

func (kd *KafkaDownloader) onKafkaMessageArrived(msg *actor.Message) error {
	kd.MsgBroker.Send(msg.Name, msg.Data)
	return nil
}
