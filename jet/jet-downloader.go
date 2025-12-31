package jet

import (
	"github.com/arcology-network/streamer/actor"
	jetlib "github.com/arcology-network/streamer/jet/lib"
)

type JetDownloader struct {
	stream *jetlib.JetKVStreamer
}

// return a Subscriber struct
func NewJetDownloader(concurrency int, groupid string, stream *jetlib.JetKVStreamer) *JetDownloader {
	downloader := JetDownloader{}
	downloader.stream = stream
	downloader.stream.AddListener(&downloader)
	return &downloader
}
func (kd *JetDownloader) RpcConfig() (string, int) {
	return "", 0
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

func (kd *JetDownloader) OnStart() {
}

func (kd *JetDownloader) OnMessageArrived(msgs []*actor.Message) error {
	return nil
}

func (kd *JetDownloader) Notify(name string, data interface{}) {
	// kd.MsgBroker.Send(msg.Name, msg.Data, msg.Height)
	//
}
