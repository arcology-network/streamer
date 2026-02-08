package jetlib

import (
	"time"

	server "github.com/nats-io/nats-server/v2/server"
	nats "github.com/nats-io/nats.go"
)

// ----------------- mockState for test -----------------
type MockState struct {
	V string
}

func RunJetTestServer() *server.Server {
	opts := &server.Options{
		Port:      -1, // auto-assign port
		JetStream: true,
	}
	s, _ := server.NewServer(opts)
	go s.Start()

	// wait for ready
	if !s.ReadyForConnections(10 * time.Second) {
		// if not ready panic to fail fast
		panic("nats test server failed to start")
	}
	// small sleep to ensure JS ready
	// time.Sleep(50 * time.Millisecond)
	// wait JetStream subsystem ready
	for {
		if s.JetStreamEnabled() {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	// 必须自行检查 JetStream 是否 ready（v2.10）
	waitJetStreamReady(s)
	return s
}

// v2.10.x 版本没有 JetStreamIsReady，只能自己检测 API 是否能正常调用
func waitJetStreamReady(s *server.Server) {
	url := s.ClientURL()

	// 最多等待 2 秒
	deadline := time.Now().Add(2 * time.Second)

	for time.Now().Before(deadline) {
		nc, err := nats.Connect(url)
		if err == nil {
			js, err := nc.JetStream()
			if err == nil {
				// 请求账户信息来确认JS是否可用
				_, err = js.AccountInfo()
				nc.Close()
				if err == nil {
					return // JetStream 已准备好
				}
			}
			nc.Close()
		}
		time.Sleep(20 * time.Millisecond)
	}

	panic("JetStream not ready after start")
}
