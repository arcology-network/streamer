package logger

import (
	"io"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

func buildOutput(cfg Config) io.Writer {
	var writers []io.Writer

	if cfg.EnableStdout {
		writers = append(writers, os.Stdout)
	}

	if cfg.LogFile != "" {
		writers = append(writers, &lumberjack.Logger{
			Filename:   cfg.LogFile,
			MaxSize:    cfg.MaxSize,
			MaxAge:     cfg.MaxAge,
			MaxBackups: cfg.MaxBackups,
			Compress:   true,
		})
	}

	if len(writers) == 1 {
		return writers[0]
	}

	return io.MultiWriter(writers...)
}
