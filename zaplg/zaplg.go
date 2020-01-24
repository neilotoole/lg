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

// New returns a Log that writes to os.Stdout
// in text format, reporting the caller.
func New() *Log {
	return NewWith(os.Stdout, "text", true, 0)
}

// NewWith returns a Log that writes to w. Format should be one
// of "json" or "text"; defaults to "json". If caller is true
// the call site is logged. The addCallerSkip param is used to
// to adjust the frame reported as the caller.
//
// Use NewWithZap if more control over output options is desired.
func NewWith(w io.Writer, format string, caller bool, addCallerSkip int) *Log {
	encoderCfg := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		CallerKey:      "caller",
		EncodeCaller:   zapcore.ShortCallerEncoder,
		TimeKey:        "time",
		EncodeTime:     timeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}

	term := isTerminal(w)

	switch {
	case term && format == "text":
		encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	case term:
		encoderCfg.EncodeLevel = zapcore.LowercaseColorLevelEncoder
	case format == "text":
		encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	default:
		encoderCfg.EncodeLevel = zapcore.LowercaseLevelEncoder
	}

	writeSyncer := zapcore.AddSync(w)
	level := zap.NewAtomicLevelAt(zap.DebugLevel)
	var core zapcore.Core

	switch format {
	case "text":
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
	return &Log{*logger.Sugar()}
}

// Log implements lg.Log.
type Log struct {
	zap.SugaredLogger
}

func (l *Log) WarnIfError(err error) {
	if err == nil {
		return
	}

	logger := l.Desugar().WithOptions(zap.AddCallerSkip(1))
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

	logger := l.Desugar().WithOptions(zap.AddCallerSkip(1))
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
	return NewWith(w, "text", false, 0)
}
