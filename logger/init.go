package logger

var Log Logger

const (
	MSG_SEND = "message_sent"
	MSG_RECV = "message_received"
	RPC_SEND = "rpc_sent"
	RPC_RECV = "rpc_received"
)

func InitLog(logcfgfile, logfile string) {
	cfg := LoadConfigWithFallback(logcfgfile)
	if len(logfile) > 0 {
		cfg.LogFile = logfile
	}
	Log = New(*cfg)
}
