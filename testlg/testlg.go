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
// This Log type does not itself generate log messages: this is
// delegated to a backing log impl (zaplg by default).
// An alternative impl can be set by passing a log factory func
// to NewWith, or by changing the testlg.FactoryFn package variable.
package testlg

import (
	"bytes"
	"io"
	"sync"
	"testing"

	"github.com/neilotoole/lg"
	"github.com/neilotoole/lg/zaplg"
)

// FactoryFn is used by New to create the backing Log impl.
// By default this func uses zaplg, but other impls
// could be used as follows:
//
//	// Use loglg as the log implementation.
//	testlg.FactoryFn = func(w io.Writer) lg.Log {
//	  return loglg.NewWith(w, true, true, false)
//	}
var FactoryFn = zaplg.TestingFactoryFn

// Log implements lg.Log, but directs its output to
// the logging functions of testing.T. This is implemented
// by having Log's underlying log impl writer to a buffer, and
// then the bytes of the buffer are passed to t.Log. The advantage
// of this approach is that Log maintains control over the
// calldepth when t.Log is invoked, thus t.Log outputs the
// correct caller information. Notably The uber/zap library's own
// testing.T wrapper results in t.Log outputting incorrect caller
// info (and this can't be fixed, because t.Helper only adjusts the
// calldepth by 1, which is insufficient given zap's structure).
type Log struct {
	t    testing.TB
	mu   sync.Mutex
	impl lg.Log
	buf  *bytes.Buffer

	factoryFn func(writer io.Writer) lg.Log
	kvs       []keyVal
}

// New returns a log that pipes output to t.
func New(t testing.TB) lg.Log {
	return NewWith(t, FactoryFn)
}

// NewWith returns a Log that pipes output to t, using
// the backing lg.Log instances returned by factoryFn
// to generate log messages.
func NewWith(t testing.TB, factoryFn func(io.Writer) lg.Log) *Log {
	tl := &Log{t: t, buf: &bytes.Buffer{}, factoryFn: factoryFn}
	tl.impl = factoryFn(tl.buf)
	return tl
}

// Debug logs at DEBUG level to t.Log.
func (l *Log) Debug(a ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.impl.Debug(a...)

	l.t.Helper()
	l.t.Log(string(stripNewLineEnding(l.buf.Bytes())))
	l.buf.Reset()
}

// Debugf logs at DEBUG level to t.Log.
func (l *Log) Debugf(format string, a ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.impl.Debugf(format, a...)

	l.t.Helper()
	l.t.Log(string(stripNewLineEnding(l.buf.Bytes())))
	l.buf.Reset()
}

// Warn implements Log.Warn.
func (l *Log) Warn(a ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.impl.Warn(a...)

	l.t.Helper()
	l.t.Log(string(stripNewLineEnding(l.buf.Bytes())))
	l.buf.Reset()
}

// Warnf implements Log.Warnf.
func (l *Log) Warnf(format string, a ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.impl.Warnf(format, a...)

	l.t.Helper()
	l.t.Log(string(stripNewLineEnding(l.buf.Bytes())))
	l.buf.Reset()
}

// WarnIfError implements Log.WarnIfError.
func (l *Log) WarnIfError(err error) {
	if err == nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.impl.Warn(err)

	l.t.Helper()
	l.t.Log(string(stripNewLineEnding(l.buf.Bytes())))
	l.buf.Reset()
}

// WarnIfFuncError implements Log.WarnIfFuncError.
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
	output, _ := io.ReadAll(l.buf)

	l.t.Helper()
	l.t.Log(string(stripNewLineEnding(output)))
}

// WarnIfCloseError implements Log.WarnIfCloseError.
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
	output, _ := io.ReadAll(l.buf)

	l.t.Helper()
	l.t.Log(string(stripNewLineEnding(output)))
}

// Error implements Log.Error.
func (l *Log) Error(a ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.impl.Error(a...)
	output, _ := io.ReadAll(l.buf)

	l.t.Helper()
	l.t.Log(string(stripNewLineEnding(output)))
}

// Errorf implements Log.Errorf.
func (l *Log) Errorf(format string, v ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.impl.Errorf(format, v...)
	output, _ := io.ReadAll(l.buf)

	l.t.Helper()
	l.t.Log(string(stripNewLineEnding(output)))
}

// With implements Log.With.
func (l *Log) With(key string, val any) lg.Log {
	// We want to prevent duplicate keys. The below code
	// results in a []keyVal without duplicate keys.

	keyIndex := -1
	for i, kv := range l.kvs {
		if kv.k == key {
			keyIndex = i
			break
		}
	}

	var kvs []keyVal
	if keyIndex == -1 {
		// Key does not exist.
		kvs = make([]keyVal, len(l.kvs)+1)
		copy(kvs, l.kvs)
		kvs[len(kvs)-1] = keyVal{k: key, v: val}
	} else {
		// Key does exists. We make a copy of l.kvs and set
		// the val for the existing key.
		kvs = make([]keyVal, len(l.kvs))
		copy(kvs, l.kvs)
		kvs[keyIndex].v = val
	}

	// Create a new log instance, and then add each
	// of kvs using impl.With.
	buf := &bytes.Buffer{}
	impl := l.factoryFn(buf)
	for _, kv := range kvs {
		impl = impl.With(kv.k, kv.v)
	}

	return &Log{
		t:         l.t,
		impl:      impl,
		buf:       buf,
		factoryFn: l.factoryFn,
		kvs:       kvs,
	}
}

type keyVal struct {
	k string
	v any
}

// stripNewLineEnding strips the trailing newline from
// the output generated by Log impls (which typically add
// a newline).
func stripNewLineEnding(b []byte) []byte {
	if len(b) == 0 {
		return b
	}

	if b[len(b)-1] == '\n' {
		b = b[:len(b)-1]
	}

	return b
}
