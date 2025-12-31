package logger

var Log Logger

const (
	MSG_SEND = "message_sent"
	MSG_RECV = "message_received"
	RPC_SEND = "rpc_sent"
	RPC_RECV = "rpc_received"
)

func init() {
	cfg := LoadConfigWithFallback("./log.toml")
	Log = New(*cfg)
}
