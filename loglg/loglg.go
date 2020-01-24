// Package loglg adapts stdlib's log pkg for
// use with the lg interface.
package loglg

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// New returns a Log instance that writes to os.Stdout,
// reporting caller and log level.
func New() *Log {
	return NewWith(os.Stdout, true, true, true)
}

// NewWith returns a Log that writes to w. If caller is true,
// the call site is logged. If level is true, the log level
// (DEBUG, WARN, ERROR) is logged.
func NewWith(w io.Writer, timestamp, level, caller bool) *Log {
	var flag int
	if timestamp {
		flag = log.Ltime | log.Lmicroseconds
	}

	if caller {
		flag = flag | log.Lshortfile
	}

	logger := log.New(w, "", flag)
	return &Log{hasPrefix: timestamp || caller, level: level, impl: logger}
}

// Log implements lg.Log.
type Log struct {
	hasPrefix bool
	impl      *log.Logger
	level     bool
}

func (l *Log) Debugf(format string, a ...interface{}) {
	_ = l.impl.Output(2, l.sprintf("DEBUG", format, a...))
}

func (l *Log) Warnf(format string, a ...interface{}) {
	_ = l.impl.Output(2, l.sprintf("WARN", format, a...))
}

func (l *Log) Errorf(format string, a ...interface{}) {
	_ = l.impl.Output(2, l.sprintf("ERROR", format, a...))
}

func (l *Log) WarnIfError(err error) {
	if err == nil {
		return
	}

	_ = l.impl.Output(2, l.sprintf("WARN", err.Error()))
}

func (l *Log) WarnIfFnError(fn func() error) {
	if fn == nil {
		return
	}

	err := fn()
	if err == nil {
		return
	}

	_ = l.impl.Output(2, l.sprintf("WARN", err.Error()))
}

func (l *Log) sprintf(level, format string, a ...interface{}) string {
	sb := strings.Builder{}
	if l.hasPrefix {
		sb.WriteString("\t")
	}

	if l.level {
		level = fmt.Sprintf("%-5s\t", level)
		sb.WriteString(level)
	}

	sb.WriteString(fmt.Sprintf(format, a...))

	return sb.String()
}
