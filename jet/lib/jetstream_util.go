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
	// You must manually check whether JetStream is ready (v2.10).
	waitJetStreamReady(s)
	return s
}

// In version v2.10.x, there is no JetStreamIsReady, so you can only manually check if the API can be called normally.
func waitJetStreamReady(s *server.Server) {
	url := s.ClientURL()

	// Wait for up to 2 seconds
	deadline := time.Now().Add(2 * time.Second)

	for time.Now().Before(deadline) {
		nc, err := nats.Connect(url)
		if err == nil {
			js, err := nc.JetStream()
			if err == nil {
				// Request account information to confirm if JS is available.
				_, err = js.AccountInfo()
				nc.Close()
				if err == nil {
					return // JetStream Ready
				}
			}
			nc.Close()
		}
		time.Sleep(20 * time.Millisecond)
	}

	panic("JetStream not ready after start")
}
