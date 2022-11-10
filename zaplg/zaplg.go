// Package zaplg adapts Uber's zap logging library for
// use with the lg interface.
package zaplg

import (
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/neilotoole/lg"
)

const (
	jsonFormat    = "json"
	textFormat    = "text"
	testingFormat = "testing"
)

// rfc3339Milli is an RFC3339 format with millisecond precision.
const rfc3339Milli = "2006-01-02T15:04:05.000Z07:00"

// New returns a Log that writes to os.Stdout
// in text format, reporting the timestamp, level, and caller.
func New() *Log {
	return NewWith(os.Stdout, textFormat, true, true, true, true, 0)
}

// timeEncoderOfLayout returns TimeEncoder which serializes a time.Time using
// given layout. If arg utc is true, the time is always converted to UTC.
func timeEncoderOfLayout(layout string, utc bool) zapcore.TimeEncoder {
	timeEncoderFn := zapcore.TimeEncoderOfLayout(layout)
	return func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		if utc {
			t = t.UTC()
		}
		timeEncoderFn(t, enc)
	}
}

// NewWith returns a Log that writes to w. Format should be one
// of "json", "text", or "testing"; defaults to "text". The timestamp, level
// and caller params determine if those fields are reported. If timestamp is
// true and utc is also true, the timestamp is displayed in UTC time.
// The addCallerSkip param is used to adjust the frame
// reported as the caller.
func NewWith(w io.Writer, format string, timestamp, utc, level, caller bool, addCallerSkip int) *Log {
	encoderCfg := zapcore.EncoderConfig{
		MessageKey:     "message",
		EncodeDuration: zapcore.StringDurationEncoder,
	}

	if caller {
		encoderCfg.CallerKey = "caller"
		if format == testingFormat {
			encoderCfg.EncodeCaller = testingCallerEncoder
		} else {
			encoderCfg.EncodeCaller = funcCallerEncoder
		}
	}

	if timestamp {
		encoderCfg.TimeKey = "timestamp"
		encoderCfg.EncodeTime = timeEncoderOfLayout(rfc3339Milli, utc)
	}

	if level {
		encoderCfg.LevelKey = "level"
	}

	switch {
	case format == textFormat, format == testingFormat:
		encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	default:
		encoderCfg.EncodeLevel = zapcore.LowercaseLevelEncoder
	}

	writeSyncer := zapcore.AddSync(w)
	zLevel := zap.NewAtomicLevelAt(zap.DebugLevel)
	var core zapcore.Core

	switch format {
	case jsonFormat:
		core = zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), writeSyncer, zLevel)
	default: // case text
		core = zapcore.NewCore(zapcore.NewConsoleEncoder(encoderCfg), writeSyncer, zLevel)
	}

	logger := zap.New(core)
	if caller {
		logger = logger.WithOptions(zap.AddCaller(), zap.AddCallerSkip(addCallerSkip))
	}

	sugarLogger := logger.Sugar()
	return &Log{SugaredLogger: sugarLogger, proto: logger}
}

// Log wraps zap's logger, adding the WarnIf_ functions.
type Log struct {
	*zap.SugaredLogger
	mu sync.Mutex

	// proto holds the unadulterated prototype logger instance.
	// This is used by method With to build a new logger with
	// kvs when the key already has been added by a previous
	// invocation of With (all of this in the service of avoiding
	// duplicate field output).
	proto *zap.Logger

	// kvs holds the set of keyVals added via method With.
	kvs []keyVal

	// callerSkip is additional caller callerSkip.
	callerSkip int
}

type keyVal struct {
	k string
	v any
}

func (l *Log) WarnIfError(err error) {
	if err == nil {
		return
	}

	logger := l.Desugar().WithOptions(zap.AddCallerSkip(1))
	logger.Warn(err.Error())
}

// AddCallerSkip adds additional caller skip.
func (l *Log) AddCallerSkip(skip int) lg.Log {
	return &Log{
		SugaredLogger: l.Desugar().WithOptions(zap.AddCallerSkip(skip)).Sugar(),
		proto:         l.proto,
		kvs:           l.kvs,
		callerSkip:    l.callerSkip + skip,
	}
}
func (l *Log) WarnIfFuncError(fn func() error) {
	if fn == nil {
		return
	}

	err := fn()
	if err == nil {
		return
	}

	logger := l.Desugar().WithOptions(zap.AddCallerSkip(1))
	logger.Warn(err.Error())
}

func (l *Log) WarnIfCloseError(c io.Closer) {
	if c == nil {
		return
	}

	err := c.Close()
	if err == nil {
		return
	}

	logger := l.Desugar().WithOptions(zap.AddCallerSkip(1))
	logger.Warn(err.Error())
}

func (l *Log) With(key string, val any) lg.Log {
	l.mu.Lock()
	defer l.mu.Unlock()

	// zap allows there to be multiple fields with the same key.
	// Thus l.With("k1", 1).With("k1", 2) will print {"k1":1, "k1:2}
	// which is dodgy output (especially for JSON). The code
	// below works around that.

	var kvs []keyVal
	var impl *zap.SugaredLogger
	keyIndex := -1

	for i, kv := range l.kvs {
		if kv.k == key {
			keyIndex = i
			break
		}
	}

	if keyIndex == -1 {
		// Key does not exist.
		impl = l.SugaredLogger.With(key, val)

		kvs = make([]keyVal, len(l.kvs)+1)
		copy(kvs, l.kvs)
		kvs[len(kvs)-1] = keyVal{k: key, v: val}

		return &Log{proto: l.proto, kvs: kvs, SugaredLogger: impl, callerSkip: l.callerSkip}
	}

	// Key does exists. We make a copy of l.kvs and set
	// the val for the existing key.
	kvs = make([]keyVal, len(l.kvs))
	copy(kvs, l.kvs)
	kvs[keyIndex].v = val

	// Builds args slice for kvs
	args := make([]any, len(kvs)*2)
	for i := 0; i < len(kvs); i++ {
		args[i*2] = kvs[i].k
		args[(i*2)+1] = kvs[i].v
	}

	// Use the proto to build the new logger.
	impl = l.proto.WithOptions(zap.AddCallerSkip(l.callerSkip)).Sugar().With(args...)

	return &Log{proto: l.proto, kvs: kvs, SugaredLogger: impl, callerSkip: l.callerSkip}
}

// TestingFactoryFn can be passed to testlg.NewWith to
// use zap as the backing impl.
var TestingFactoryFn = func(w io.Writer) lg.Log {
	// caller arg is false because testing.T will
	// report the caller anyway.
	return NewWith(w, testingFormat, true, true, true, true, 1)
}

// funcCallerEncoder extends the behavior of zapcore.ShortCallerEncoder
// to also include the calling function name. That is, it
// serializes the caller in package/file:line:func format,
// trimming all but the final directory from the full path.
// This implementation is probably not very efficient, so
// use with caution.
func funcCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	if !caller.Defined {
		return
	}

	frame, _ := runtime.CallersFrames([]uintptr{caller.PC}).Next()
	// ditch the path
	s := frame.Function[strings.LastIndex(frame.Function, "/")+1:]
	// and ditch the package
	s = s[strings.IndexRune(s, '.')+1:]
	enc.AppendString(caller.TrimmedPath() + ":" + s)
}

// funcCallerEncoder serializes the caller in package.func format.
// This is especially useful when working with the testing
// framework, t.Log etc already report file:line.
// This implementation is probably not very efficient, so
// use with caution.
func testingCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	if !caller.Defined {
		return
	}

	frame, _ := runtime.CallersFrames([]uintptr{caller.PC}).Next()
	// ditch the path
	s := "[" + frame.Function[strings.LastIndex(frame.Function, "/")+1:] + "]"
	enc.AppendString(s)
}
