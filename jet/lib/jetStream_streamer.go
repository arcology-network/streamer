package jetlib

import (
	"fmt"
	"log"
	"strings"
	"time"

	scommon "github.com/arcology-network/streamer/common"
	nats "github.com/nats-io/nats.go"
)

// ---------------- JetKVStreamer using JetStream KV -----------------
type JetKVStreamer struct {
	nc       *nats.Conn
	js       nats.JetStreamContext
	bucket   string
	cfg      *JetStreamConfig
	listener JetListener
}

func NewJetKVStreamerFromConfig(cfg *JetStreamConfig) (*JetKVStreamer, error) {
	opts := []nats.Option{
		nats.Name(cfg.Nats.Name),
		nats.ReconnectWait(cfg.Nats.ReconnectWait),
		nats.MaxReconnects(cfg.Nats.MaxReconnects),
		nats.UserInfo(cfg.Nats.User, cfg.Nats.Password),
	}

	nc, err := nats.Connect(
		strings.Join(cfg.Nats.Servers, ","),
		opts...)
	if err != nil {
		return nil, err
	}

	js, err := nc.JetStream()
	if err != nil {
		return nil, err
	}

	storage := nats.MemoryStorage
	if cfg.Stream.Storage == "file" {
		storage = nats.FileStorage
	}

	_, _ = js.AddStream(&nats.StreamConfig{
		Name:     cfg.Stream.Name,
		Subjects: cfg.Stream.Subjects,
		Storage:  storage,
		MaxBytes: cfg.Stream.MaxBytes,
		Replicas: cfg.Stream.Replicas,
	})

	return &JetKVStreamer{
		nc:     nc,
		js:     js,
		bucket: cfg.Stream.Name,
		cfg:    cfg,
	}, nil
}
func (jsm *JetKVStreamer) AddListener(listenr JetListener) {
	jsm.listener = listenr
}
func (jsm *JetKVStreamer) GetConfig() *JetStreamConfig {
	return jsm.cfg
}
func (jsm *JetKVStreamer) buildDurableName(consumerName string, topic string, nodeID string) string {
	if nodeID == "" {
		nodeID = "node01"
	}

	if jsm.cfg.TopicsD[topic].SubType == Broadcast {
		// ✅ One cursor per node
		return fmt.Sprintf("%s.%s.%s", consumerName, topic, nodeID)
	}

	// ✅ All nodes share a single cursor (load balancing)
	return fmt.Sprintf("%s.%s", consumerName, topic)
}

func (jsm *JetKVStreamer) Serve() {
	// for _, c := range jsm.consumers {
	// 	go c.GetStreamController().Serve()

	for input := range jsm.cfg.TopicsD {
		j_topic := TopicForJet(input)
		subject := jsm.bucket + "." + j_topic
		durable := jsm.buildDurableName("", j_topic, jsm.cfg.Nats.Name)
		durable = strings.ReplaceAll(durable, ".", "_")

		_, err := jsm.js.AddConsumer(jsm.bucket, &nats.ConsumerConfig{
			Durable:       durable,
			AckPolicy:     nats.AckExplicitPolicy,
			AckWait:       jsm.cfg.Consumer.AckWait,
			MaxDeliver:    jsm.cfg.Consumer.MaxDeliver,
			MaxAckPending: jsm.cfg.Consumer.MaxAckPending,
			FilterSubject: subject, // ← CRUCIAL! Only pull messages from this topic
		})

		if err != nil && err != nats.ErrConsumerNameAlreadyInUse {
			log.Fatalf("AddConsumer error: %v", err)
		}

		sub, err := jsm.js.PullSubscribe(
			subject,
			durable,
			nats.Bind(jsm.bucket, durable),
		)
		if err != nil {
			log.Fatalf("PullSubscribe error: %v", err)
		}

		go jsm.pullWorker(sub, jsm.listener)
	}
}

// ---------------- Worker -----------------
func (jsm *JetKVStreamer) pullWorker(sub *nats.Subscription, listener JetListener) {
	for {
		msgs, err := sub.Fetch(jsm.cfg.Consumer.PullBatch, nats.MaxWait(5*time.Second))
		if err != nil {
			continue
		}

		for _, m := range msgs {
			msg := scommon.Message{}

			// Decode failure → Permanent failure
			if err := scommon.GobDecode(m.Data, &msg); err != nil {
				_ = m.Term()
				continue
			}

			meta, err := m.Metadata()
			if err != nil {
				_ = m.Nak()
				continue
			}

			msg.Sequence = meta.Sequence.Stream
			msg.Redeliver = int(meta.NumDelivered) - 1

			msg.Name = TopicForUp(TopicForBase(msg.Name))

			listener.Notify(msg.Name, *msg.DeepCopy())

			_ = m.Ack() // Successful acknowledgment
		}
	}
}

func (jsm *JetKVStreamer) Send(topic string, msg *scommon.Message) {
	topic = TopicForJet(TopicForBase(topic))
	msg.Name = topic
	jsm.js.Publish(jsm.bucket+"."+topic, scommon.GobEncode(msg))
}
