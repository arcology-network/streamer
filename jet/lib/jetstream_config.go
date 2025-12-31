package jetlib

import (
	"fmt"
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v3"
)

type ConsumeMode int

const (
	Broadcast   ConsumeMode = iota // Each node independently Durables the broadcast message
	LoadBalance                    // Multi-node sharing Durable + Queue Group
)

func (m *ConsumeMode) UnmarshalYAML(value *yaml.Node) error {
	switch value.Value {
	case "broadcast":
		*m = Broadcast
	case "loadBalance":
		*m = LoadBalance
	default:
		return fmt.Errorf(
			"invalid topic mode %q (must be broadcast|loadBalance)",
			value.Value,
		)
	}
	return nil
}

type TopicConsumeConfig struct {
	SubType ConsumeMode `yaml:"sub_type"`
	BufSize int         `yaml:"buf_size"`
}

type JetStreamConfig struct {
	Nats struct {
		Servers       []string      `yaml:"servers"`
		Name          string        `yaml:"name"`
		ReconnectWait time.Duration `yaml:"reconnect_wait"`
		MaxReconnects int           `yaml:"max_reconnects"`
	} `yaml:"nats"`

	Stream struct {
		Name     string   `yaml:"name"`
		Subjects []string `yaml:"subjects"`
		Storage  string   `yaml:"storage"`
		MaxBytes int64    `yaml:"max_bytes"`
		Replicas int      `yaml:"replicas"`
	} `yaml:"stream"`

	Consumer struct {
		AckWait       time.Duration `yaml:"ack_wait"`
		MaxDeliver    int           `yaml:"max_deliver"`
		MaxAckPending int           `yaml:"max_ack_pending"`
		PullBatch     int           `yaml:"pull_batch"`
		Workers       int           `yaml:"workers"`
	} `yaml:"consumer"`

	// topics_d: Topic with consumption semantics
	TopicsD map[string]TopicConsumeConfig `yaml:"topics_d"`

	// topics_u: Topic for publishing only
	TopicsU []string `yaml:"topics_u"`
}

// ---------------- load config -----------------
func LoadConfig(path string) (*JetStreamConfig, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg := &JetStreamConfig{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
