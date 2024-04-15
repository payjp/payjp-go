package payjp

import (
	"fmt"
	"io"
	"os"
)

// LogLevel はログの出力レベルを表します。
type LogLevel int

const (
	LogLevelNull  LogLevel = 0
	LogLevelError LogLevel = 1
	LogLevelWarn  LogLevel = 2
	LogLevelInfo  LogLevel = 3
	LogLevelDebug LogLevel = 4
)

// LoggerInterface はログ出力を行うためのインターフェースです。
type LoggerInterface interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// NullLogger はログを出力しないロガーです。
var NullLogger LoggerInterface = NewPayjpLogger(LogLevelNull)

type PayjpLogger struct {
	logLevel LogLevel

	stdoutOverride io.Writer
	stderrOverride io.Writer
}

var _ LoggerInterface = &PayjpLogger{}

// NewPayjpLogger はPayjpLoggerを生成します。
// ログの出力レベルを指定することができます。
func NewPayjpLogger(logLevel LogLevel) *PayjpLogger {
	return &PayjpLogger{logLevel: logLevel}
}

func (l *PayjpLogger) stdout() io.Writer {
	if l.stdoutOverride != nil {
		return l.stdoutOverride
	}

	return os.Stdout
}

func (l *PayjpLogger) stderr() io.Writer {
	if l.stderrOverride != nil {
		return l.stderrOverride
	}

	return os.Stderr
}

func (l *PayjpLogger) Debugf(format string, args ...interface{}) {
	if l.logLevel >= LogLevelDebug {
		fmt.Fprintf(l.stdout(), "[DEBUG] "+format+"\n", args...)
	}
}

func (l *PayjpLogger) Infof(format string, args ...interface{}) {
	if l.logLevel >= LogLevelInfo {
		fmt.Fprintf(l.stdout(), "[INFO] "+format+"\n", args...)
	}
}

func (l *PayjpLogger) Warnf(format string, args ...interface{}) {
	if l.logLevel >= LogLevelWarn {
		fmt.Fprintf(l.stderr(), "[WARN] "+format+"\n", args...)
	}
}

func (l *PayjpLogger) Errorf(format string, args ...interface{}) {
	if l.logLevel >= LogLevelError {
		fmt.Fprintf(l.stderr(), "[ERROR] "+format+"\n", args...)
	}
}
