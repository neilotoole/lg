// Package zaplg adapts Uber's zap logging library for
// use with the lg interface.
package zaplg

import (
	"io"
	"os"
	"time"

	"golang.org/x/crypto/ssh/terminal"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/neilotoole/lg"
)

const textFormat = "text"

// New returns a Log that writes to os.Stdout
// in text format, reporting the caller.
func New() *Log {
	return NewWith(os.Stdout, textFormat, true, true, 0)
}

// NewWith returns a Log that writes to w. Format should be one
// of "json" or "text"; defaults to "json". If caller is true
// the call site is logged. The addCallerSkip param is used to
// to adjust the frame reported as the caller.
//
// Use NewWithZap if more control over output options is desired.
func NewWith(w io.Writer, format string, timestamp, caller bool, addCallerSkip int) *Log {
	encoderCfg := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		EncodeDuration: zapcore.StringDurationEncoder,
	}

	if caller {
		encoderCfg.CallerKey = "caller"
		encoderCfg.EncodeCaller = zapcore.ShortCallerEncoder
	}

	if timestamp {
		encoderCfg.TimeKey = "time"
		encoderCfg.EncodeTime = timeEncoder
	}

	term := isTerminal(w)

	switch {
	case term && format == textFormat:
		encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	case term:
		encoderCfg.EncodeLevel = zapcore.LowercaseColorLevelEncoder
	case format == textFormat:
		encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	default:
		encoderCfg.EncodeLevel = zapcore.LowercaseLevelEncoder
	}

	writeSyncer := zapcore.AddSync(w)
	level := zap.NewAtomicLevelAt(zap.DebugLevel)
	var core zapcore.Core

	switch format {
	case textFormat:
		core = zapcore.NewCore(zapcore.NewConsoleEncoder(encoderCfg), writeSyncer, level)
	default: // case json
		core = zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), writeSyncer, level)
	}

	logger := zap.New(core)
	if caller {
		logger = logger.WithOptions(zap.AddCaller(), zap.AddCallerSkip(addCallerSkip))
	}

	return NewWithZap(logger)
}

// NewWithZap returns a Log using the supplied zap.Logger, thus
// permitting customization of logging behavior.
func NewWithZap(logger *zap.Logger) *Log {
	return &Log{logger.Sugar()}
}

// Log implements lg.Log.
type Log struct {
	*zap.SugaredLogger
}

const callerSkip = 1

func (l *Log) WarnIfError(err error) {
	if err == nil {
		return
	}

	logger := l.Desugar().WithOptions(zap.AddCallerSkip(callerSkip))
	logger.Warn(err.Error())
}

func (l *Log) WarnIfFnError(fn func() error) {
	if fn == nil {
		return
	}

	err := fn()
	if err == nil {
		return
	}

	logger := l.Desugar().WithOptions(zap.AddCallerSkip(callerSkip))
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

	logger := l.Desugar().WithOptions(zap.AddCallerSkip(callerSkip))
	logger.Warn(err.Error())
}

// isTerminal returns true if w is a terminal. Currently
// always returns false for windows.
func isTerminal(w io.Writer) bool {
	switch v := w.(type) {
	case *os.File:
		return terminal.IsTerminal(int(v.Fd()))
	default:
		return false
	}
}

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("15:04:05.000000"))
}

// TestingFactoryFn can be passed to testlg.NewWith to
// use zap as the backing impl.
var TestingFactoryFn = func(w io.Writer) lg.Log {
	return NewWith(w, "text", true, false, 0)
}
