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

// NewWith returns a Log that writes to w. The timestamp, level
// and caller params determine if those fields are reported.
func NewWith(w io.Writer, timestamp, level, caller bool) *Log {
	var flag int
	if timestamp {
		flag = log.Ltime | log.Lmicroseconds
	}

	if caller {
		flag |= log.Lshortfile
	}

	logger := log.New(w, "", flag)
	return &Log{hasPrefix: timestamp || caller, level: level, impl: logger}
}

const callDepth = 2

// Log implements lg.Log.
type Log struct {
	impl      *log.Logger
	hasPrefix bool
	level     bool
}

func (l *Log) Debug(a ...interface{}) {
	_ = l.impl.Output(callDepth, l.sprint("DEBUG", a...))
}
func (l *Log) Debugf(format string, a ...interface{}) {
	_ = l.impl.Output(callDepth, l.sprintf("DEBUG", format, a...))
}

func (l *Log) Warn(a ...interface{}) {
	_ = l.impl.Output(callDepth, l.sprint("WARN", a...))
}
func (l *Log) Warnf(format string, a ...interface{}) {
	_ = l.impl.Output(callDepth, l.sprintf("WARN", format, a...))
}

func (l *Log) WarnIfError(err error) {
	if err == nil {
		return
	}

	_ = l.impl.Output(callDepth, l.sprintf("WARN", err.Error()))
}

func (l *Log) WarnIfFuncError(fn func() error) {
	if fn == nil {
		return
	}

	err := fn()
	if err == nil {
		return
	}

	_ = l.impl.Output(callDepth, l.sprintf("WARN", err.Error()))
}

func (l *Log) WarnIfCloseError(c io.Closer) {
	if c == nil {
		return
	}

	err := c.Close()
	if err == nil {
		return
	}

	_ = l.impl.Output(callDepth, l.sprintf("WARN", err.Error()))
}

func (l *Log) Error(a ...interface{}) {
	_ = l.impl.Output(callDepth, l.sprint("ERROR", a...))
}
func (l *Log) Errorf(format string, a ...interface{}) {
	_ = l.impl.Output(callDepth, l.sprintf("ERROR", format, a...))
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
func (l *Log) sprint(level string, a ...interface{}) string {
	sb := strings.Builder{}
	if l.hasPrefix {
		sb.WriteString("\t")
	}

	if l.level {
		level = fmt.Sprintf("%-5s\t", level)
		sb.WriteString(level)
	}

	sb.WriteString(fmt.Sprint(a...))

	return sb.String()
}
