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
// in text format, reporting the timestamp, level, and caller.
func New() *Log {
	return NewWith(os.Stdout, textFormat, true, true, true, 0)
}

// NewWith returns a Log that writes to w. Format should be one
// of "json" or "text"; defaults to "text". The timestamp, level
// and caller params determine if those fields are reported.
// The addCallerSkip param is used to to adjust the frame
// reported as the caller.
//
// Use NewWithZap if more control over output options is desired.
func NewWith(w io.Writer, format string, timestamp, level, caller bool, addCallerSkip int) *Log {
	encoderCfg := zapcore.EncoderConfig{
		MessageKey: "msg",

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

	if level {
		encoderCfg.LevelKey = "level"
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
	zLevel := zap.NewAtomicLevelAt(zap.DebugLevel)
	var core zapcore.Core

	switch format {
	case "json":
		core = zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), writeSyncer, zLevel)
	default: // case text
		core = zapcore.NewCore(zapcore.NewConsoleEncoder(encoderCfg), writeSyncer, zLevel)
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

func (l *Log) WarnIfFuncError(fn func() error) {
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

// isTerminal returns true if w is a terminal.
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
	return NewWith(w, "text", true, true, false, 0)
}
