package logger

import (
	"fmt"
	"strings"
)

type Level int8

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

func ParseLevel(levelStr string) (Level, error) {
	var l Level
	err := l.UnmarshalText([]byte(levelStr))
	return l, err
}

// UnmarshalText implements the encoding.TextUnmarshaler interface, supporting parsing from a string
func (l *Level) UnmarshalText(text []byte) error {
	switch strings.ToLower(string(text)) {
	case "debug":
		*l = DebugLevel
	case "info":
		*l = InfoLevel
	case "warn", "warning":
		*l = WarnLevel
	case "error":
		*l = ErrorLevel
	case "fatal":
		*l = FatalLevel
	default:
		// Try to parse as a number
		var n int
		_, err := fmt.Sscanf(string(text), "%d", &n)
		if err != nil {
			return fmt.Errorf("invalid log level: %s", text)
		}
		if n < 0 || n > int(FatalLevel) {
			return fmt.Errorf("log level out of range: %d", n)
		}
		*l = Level(n)
	}
	return nil
}

// MarshalText implements the encoding.TextMarshaler interface, supporting serialization to a string
func (l Level) MarshalText() ([]byte, error) {
	return []byte(l.String()), nil
}
