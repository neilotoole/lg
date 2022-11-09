// Package testlg implements a lg.Log that
// directs output to the testing framework.
//
// This is useful if your code under test writes to a log,
// and you want to capture that log output under testing.T.
// For example:
//
//	func TestMe(t *testing.T) {
//	  log := testlg.New(t)
//	  log.Debugf("Hello %s", "World")
//	  log.Warn("Hello Mars")
//	  log.Error("Hello Venus")
//	}
//
// produces the following:
//
//	=== RUN   TestMe
//	--- PASS: TestMe (0.00s)
//	    testlg_test.go:64: 09:48:38.849066 	DEBUG	Hello World
//	    testlg_test.go:65: 09:48:38.849215 	WARN 	Hello Mars
//	    testlg_test.go:66: 09:48:38.849304 	ERROR	Hello Venus
//
// Log has a "strict" mode which pipes Error and Errorf output
// to t.Error instead of t.Log, resulting in test failure. This:
//
//	func TestMe(t *testing.T) {
//	  log := testlg.New(t).Strict(true)
//	  log.Debug("Hello World")
//	  log.Warn("Hello Mars")
//	  log.Error("Hello Venus") // pipes to t.Error, resulting in test failure
//	}
//
// produces:
//
//	=== RUN   TestMe
//	--- FAIL: TestMe (0.00s)
//	    testlg_test.go:64: 09:52:28.706482 	DEBUG	Hello World
//	    testlg_test.go:65: 09:52:28.706591 	WARN 	Hello Mars
//	    testlg_test.go:66: 09:52:28.706599 	ERROR	Hello Venus
//
// This Log type does not itself generate log messages: this is
// delegated to a backing log impl (zaplg by default).
// An alternative impl can be set by passing a log factory func
// to NewWith, or by changing the testlg.FactoryFn package variable.
package testlg

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/neilotoole/lg"
	"github.com/neilotoole/lg/zaplg"
)

// FactoryFn is used by New to create the backing Log impl.
// By default this func uses zaplg, but other impls
// can be used as follows:
//
//	// Use loglg as the log implementation.
//	testlg.FactoryFn = func(w io.Writer) lg.Log {
//	  return loglg.NewWith(w, true, true, false)
//	}
var FactoryFn = zaplg.TestingFactoryFn

// Log implements lg.Log, but directs its output to
// the logging functions of testing.T.
type Log struct {
	t      testing.TB
	strict bool
	impl   lg.Log
	buf    bytes.Buffer
	mu     sync.Mutex
}

// New returns a log that pipes output to t.
func New(t testing.TB) *Log {
	return NewWith(t, FactoryFn)
}

// NewWith returns a Log that pipes output to t, using
// the backing lg.Log instances returned by factoryFn
// to generate log messages.
func NewWith(t testing.TB, factoryFn func(io.Writer) lg.Log) *Log {
	tl := &Log{t: t}
	tl.impl = factoryFn(&tl.buf)
	return tl
}

// Strict sets strict mode. When in strict mode, Errorf logs
// via t.Error instead of t.Log, thus resulting in test failure.
func (l *Log) Strict(strict bool) *Log {
	l.strict = strict
	return l
}

// Debug logs at DEBUG level to t.Log.
func (l *Log) Debug(a ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.impl.Debug(a...)
	output, _ := io.ReadAll(&l.buf)

	l.t.Helper()
	l.t.Log(stripNewLineEnding(string(output)))
}

// Debugf logs at DEBUG level to t.Log.
func (l *Log) Debugf(format string, a ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.impl.Debugf(format, a...)
	output, _ := io.ReadAll(&l.buf)

	l.t.Helper()
	l.t.Log(stripNewLineEnding(string(output)))
}

// Warn logs at WARN level to t.Log.
func (l *Log) Warn(a ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.impl.Warn(a...)
	output, _ := io.ReadAll(&l.buf)

	l.t.Helper()
	l.t.Log(stripNewLineEnding(string(output)))
}

// Warnf logs at WARN level to t.Log.
func (l *Log) Warnf(format string, a ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.impl.Warnf(format, a...)
	output, _ := io.ReadAll(&l.buf)

	l.t.Helper()
	l.t.Log(stripNewLineEnding(string(output)))
}

func (l *Log) WarnIfError(err error) {
	if err == nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.impl.Warn(err)
	output, _ := io.ReadAll(&l.buf)

	l.t.Helper()
	l.t.Log(stripNewLineEnding(string(output)))
}

func (l *Log) WarnIfFuncError(fn func() error) {
	if fn == nil {
		return
	}

	err := fn()
	if err == nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.impl.Warn(err)
	output, _ := io.ReadAll(&l.buf)

	l.t.Helper()
	l.t.Log(stripNewLineEnding(string(output)))
}

func (l *Log) WarnIfCloseError(c io.Closer) {
	if c == nil {
		return
	}

	err := c.Close()
	if err == nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.impl.Warn(err)
	output, _ := io.ReadAll(&l.buf)

	l.t.Helper()
	l.t.Log(stripNewLineEnding(string(output)))
}

// Error logs at ERROR level to t.Log, or if in strict mode,
// the message is logged via t.Error, resulting in test failure.
func (l *Log) Error(a ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.impl.Error(a...)
	output, _ := io.ReadAll(&l.buf)

	l.t.Helper()

	if l.strict {
		l.t.Error(stripNewLineEnding(string(output)))
	} else {
		l.t.Log(stripNewLineEnding(string(output)))
	}
}

// Errorf logs at ERROR level to t.Log, or if in strict mode,
// the message is logged via t.Error, resulting in test failure.
func (l *Log) Errorf(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.impl.Errorf(format, v...)
	output, _ := io.ReadAll(&l.buf)

	l.t.Helper()

	if l.strict {
		l.t.Error(stripNewLineEnding(string(output)))
	} else {
		l.t.Log(stripNewLineEnding(string(output)))
	}
}

// stripNewLineEnding strips the trailing newline from
// the output generated by Log impls (which typically add
// a newline).
func stripNewLineEnding(s string) string {
	return strings.TrimSuffix(s, "\n")
}
