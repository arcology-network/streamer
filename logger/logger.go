package logger

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger interface {
	Debug(ctx context.Context, module string, msg string, fields ...Field)
	Info(ctx context.Context, module string, msg string, fields ...Field)
	Warn(ctx context.Context, module string, msg string, fields ...Field)
	Error(ctx context.Context, module string, msg string, fields ...Field)
	Fatal(ctx context.Context, module string, msg string, fields ...Field)

	Sync() // ✅ Force flush to disk, no log loss

}

type jsonLogger struct {
	cfg    Config
	fields []Field

	out    chan map[string]any
	writer io.Writer

	wg sync.WaitGroup
	mu sync.Mutex

	// ✅✅✅ Current log level (supports dynamic modification)
	level Level
}

// ✅✅✅ Automatically create directories
func ensureDir(logPath string) error {
	dir := filepath.Dir(logPath)
	return os.MkdirAll(dir, 0755)
}

func New(cfg Config) Logger {
	var writers []io.Writer

	if cfg.EnableStdout {
		writers = append(writers, os.Stdout)
	}

	if cfg.LogFile != "" {
		_ = ensureDir(cfg.LogFile)

		lj := &lumberjack.Logger{
			Filename:   cfg.LogFile,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   true,
		}
		writers = append(writers, lj)
	}

	mw := io.MultiWriter(writers...)

	l := &jsonLogger{
		cfg:    cfg,
		out:    make(chan map[string]any, 8192),
		writer: mw,
		level:  cfg.Level, // ✅
	}

	l.wg.Add(1)
	go l.loop()

	return l
}

func (l *jsonLogger) loop() {
	defer l.wg.Done()

	enc := json.NewEncoder(l.writer)

	for rec := range l.out {
		_ = enc.Encode(rec)
	}
}

func (l *jsonLogger) log(
	ctx context.Context,
	level Level,
	module string,
	msg string,
	fields ...Field,
) {
	// ✅✅✅ Key: Logs below the current level are directly discarded
	if level < l.level {
		return
	}

	record := map[string]any{
		"ts":    time.Now().UnixMilli(),
		"level": level.String(),
		// "service": l.cfg.Service,
		"module":  module,
		"event":   msg,
		"version": l.cfg.Version,
	}

	if t, ok := GetTrace(ctx); ok {
		record["trace_id"] = t.TraceID
		// record["span_id"] = t.SpanID
		record["span_name"] = t.SpanName
		// record["parent_id"] = t.ParentID
		// record["request_id"] = t.ReqID
	}

	if m, ok := GetMsg(ctx); ok {
		record["msg_type"] = m.MsgType
		record["f_topic"] = m.Name
		// record["key"] = m.Key
		record["sequence"] = m.Sequence
		// record["redeliver"] = m.Redeliver
		record["from"] = m.From
		record["height"] = m.Height
		record["method"] = m.Method
		record["request_id"] = m.ReqID
	}

	for _, f := range l.fields {
		record[f.Key] = f.Value
	}

	for _, f := range fields {
		record[f.Key] = f.Value
	}

	select {
	case l.out <- record:
	default:
		// ✅ Drop logs when full, never block business threads
	}

}

func (l *jsonLogger) Debug(ctx context.Context, module string, msg string, fs ...Field) {
	l.log(ctx, DebugLevel, module, msg, fs...)
}

func (l *jsonLogger) Info(ctx context.Context, module string, msg string, fs ...Field) {
	l.log(ctx, InfoLevel, module, msg, fs...)
}

func (l *jsonLogger) Warn(ctx context.Context, module string, msg string, fs ...Field) {
	l.log(ctx, WarnLevel, module, msg, fs...)
}

func (l *jsonLogger) Error(ctx context.Context, module string, msg string, fs ...Field) {
	l.log(ctx, ErrorLevel, module, msg, fs...)
}

func (l *jsonLogger) Fatal(ctx context.Context, module string, msg string, fs ...Field) {
	l.log(ctx, FatalLevel, module, msg, fs...)
}

func (l *jsonLogger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// ✅✅✅ Core capability: Safe disk flushing (mandatory for go test / program exit)
func (l *jsonLogger) Sync() {
	l.mu.Lock()
	defer l.mu.Unlock()

	close(l.out)
	l.wg.Wait()
}
