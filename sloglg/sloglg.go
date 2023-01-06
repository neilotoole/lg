// Package sloglg adapts the sloglg library for
// use with the lg interface.
package sloglg

import (
	"github.com/neilotoole/lg/v2"
	"golang.org/x/exp/slog"
	"io"
	"os"
)

const (
	jsonFormat    = "json"
	textFormat    = "text"
	testingFormat = "testing"
)

// rfc3339Milli is an RFC3339 format with millisecond precision.
const rfc3339Milli = "2006-01-02T15:04:05.000Z07:00"

var _ lg.Log = (*Log)(nil)

// New returns a Log that writes to os.Stdout
// in text format, reporting the timestamp, level, and caller.
func New() *Log {
	opts := slog.HandlerOptions{
		AddSource:   true,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	}

	handler := opts.NewTextHandler(os.Stdout)
	logger := slog.New(handler)

	return &Log{Logger: logger}
}

// NewWith returns a Log that writes to w. Format should be one
// of "json", "text", or "testing"; defaults to "text". The timestamp, level
// and caller params determine if those fields are reported. If timestamp is
// true and utc is also true, the timestamp is displayed in UTC time.
// The addCallerSkip param is used to adjust the frame
// reported as the caller.
func NewWith(w io.Writer, format string, timestamp, utc, level, caller bool, addCallerSkip int) *Log {
	opts := slog.HandlerOptions{
		AddSource:   caller,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	}

	var handler slog.Handler

	switch format {
	case textFormat:
	case jsonFormat:
	default:
		panic("invalid log format: " + format)
	}

	if format == textFormat {
		handler = opts.NewTextHandler(w)
	} else {
		handler = opts.NewJSONHandler(w)
	}
	logger := slog.New(handler)

	return &Log{Logger: logger, addCallerSkip: addCallerSkip}

	//encoderCfg := zapcore.EncoderConfig{
	//	MessageKey:     "message",
	//	EncodeDuration: zapcore.StringDurationEncoder,
	//}
	//
	//if caller {
	//	encoderCfg.CallerKey = "caller"
	//	if format == testingFormat {
	//		encoderCfg.EncodeCaller = testingCallerEncoder
	//	} else {
	//		encoderCfg.EncodeCaller = funcCallerEncoder
	//	}
	//}
	//
	//if timestamp {
	//	encoderCfg.TimeKey = "timestamp"
	//	encoderCfg.EncodeTime = timeEncoderOfLayout(rfc3339Milli, utc)
	//}
	//
	//if level {
	//	encoderCfg.LevelKey = "level"
	//}
	//
	//switch {
	//case format == textFormat, format == testingFormat:
	//	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	//default:
	//	encoderCfg.EncodeLevel = zapcore.LowercaseLevelEncoder
	//}
	//
	//writeSyncer := zapcore.AddSync(w)
	//zLevel := zap.NewAtomicLevelAt(zap.DebugLevel)
	//var core zapcore.Core
	//
	//switch format {
	//case jsonFormat:
	//	core = zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), writeSyncer, zLevel)
	//default: // case text
	//	core = zapcore.NewCore(zapcore.NewConsoleEncoder(encoderCfg), writeSyncer, zLevel)
	//}
	//
	//logger := zap.New(core)
	//if caller {
	//	logger = logger.WithOptions(zap.AddCaller(), zap.AddCallerSkip(addCallerSkip))
	//}
	//
	//sugarLogger := logger.Sugar()
	//return &Log{SugaredLogger: sugarLogger, proto: logger}
}

// Log wraps zap's logger, adding the WarnIf_ functions.
type Log struct {
	*slog.Logger

	// addCallerSkip is additional caller addCallerSkip.
	addCallerSkip int
}

type keyVal struct {
	k string
	v any
}

func (l *Log) WarnIfError(err error) {
	if err == nil {
		return
	}

	l.Warn(err.Error())
}

// AddCallerSkip adds additional caller skip.
func (l *Log) AddCallerSkip(skip int) lg.Log {
	return &Log{
		Logger:        l.Logger,
		addCallerSkip: l.addCallerSkip + skip,
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

	//logger := l.Desugar().WithOptions(zap.AddCallerSkip(1))
	l.Warn(err.Error())
}

func (l *Log) WarnIfCloseError(c io.Closer) {
	if c == nil {
		return
	}

	err := c.Close()
	if err == nil {
		return
	}

	//logger := l.Desugar().WithOptions(zap.AddCallerSkip(1))
	l.Warn(err.Error())
}

func (l *Log) Error(msg string, args ...any) {
	l.Logger.LogDepth(2+l.addCallerSkip, slog.LevelError, msg, args...)
}

func (l *Log) Err(err error) {
	if err == nil {
		return
	}
	l.Logger.Error(err.Error(), err)
}

func (l *Log) With(key string, val any) lg.Log {
	sl := l.Logger

	sl = sl.With(key, val)
	return &Log{Logger: sl}
}

// TestingFactoryFn can be passed to testlg.NewWith to
// use zap as the backing impl.
var TestingFactoryFn = func(w io.Writer) lg.Log {
	// caller arg is false because testing.T will
	// report the caller anyway.
	return NewWith(w, testingFormat, true, true, true, true, 1)
}
