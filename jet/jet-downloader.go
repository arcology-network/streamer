package jet

import (
	"github.com/arcology-network/streamer/actor"
	scommon "github.com/arcology-network/streamer/common"
	jetlib "github.com/arcology-network/streamer/jet/lib"
)

type JetDownloader struct {
	stream *jetlib.JetKVStreamer
	sender actor.OutboundSender
}

// return a Subscriber struct
func NewJetDownloader(stream *jetlib.JetKVStreamer) actor.Business {
	downloader := JetDownloader{}
	downloader.stream = stream
	downloader.stream.AddListener(&downloader)
	return &downloader
}

func (kd *JetDownloader) Inputs() ([]string, bool) {
	return []string{}, false
}

func (kd *JetDownloader) Outputs() map[string]int {
	outputs := make(map[string]int)
	cfg := kd.stream.GetConfig()
	for topic := range cfg.TopicsD {
		u_topic := jetlib.TopicForUp(topic)
		outputs[u_topic] = cfg.TopicsD[topic].BufSize
	}
	return outputs
}

func (kd *JetDownloader) RegisterActions(reg actor.ActionRegistrar) {
}
func (kd *JetDownloader) SetSender(sender actor.OutboundSender) {
	kd.sender = sender
}
func (kd *JetDownloader) Notify(name string, data interface{}) {
	msg := data.(*scommon.Message)
	kd.sender.Send(msg.Name, msg.Data, msg.Height)
}
