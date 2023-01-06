// Package lg is an exploration of a small, leveled,
// unstructured logging interface for enterprise developers.
// It is not suggested for production use.
//
// This pkg does not itself perform logging. A concrete impl
// must be used, e.g. sloglg.New. Use testlg.New
// to adapt lg to output to a testing.T.
package lg

import "io"

// Log is a logging interface that adds WarnIf methods
// to the basic Debug, Warn and Error methods. The methods
// of Log are safe for concurrent use.
//
// Style note: Being that Log is an interface, idiomatically
// the type name should be Logger. But the sense is that
// Orwell's sixth rule prevails here. It's not that lg.Logger
// is barbaric in and of itself, but a thousand appearances
// of lg.Logger vs lg.Log constitutes a horde.
type Log interface {
	// Debugf logs at DEBUG level.
	Debugf(msg string, args ...any)

	// Warnf logs at WARN level.
	Warnf(msg string, args ...any)

	// WarnIfError is no-op if err is nil; if non-nil, err
	// is logged at WARN level.
	WarnIfError(err error)

	// WarnIfFuncError is no-op if fn is nil; if fn is non-nil,
	// fn is executed and if fn's error is non-nil, that error
	// is logged at WARN level.
	WarnIfFuncError(fn func() error)

	// WarnIfCloseError is no-op if c is nil; if c is non-nil,
	// c.Close is executed and if Close's error is non-nil,
	// that error is logged at WARN level.
	//
	// WarnIfCloseError is preferred to WarnIfFuncError
	// when c may be nil.
	//
	//  var c io.Closer = nil
	//  log.WarnIfCloseError(c)      // ok
	//  log.WarnIfFuncError(c.Close) // panic
	WarnIfCloseError(c io.Closer)

	// Errorf logs at ERROR level.
	Errorf(msg string, args ...any)

	// With returns a child Log instance that has a structured
	// field key with val.
	With(key string, val any) Log
}

// addCallerSkipper is an optional interface that Log impls
// can implement to support additional caller skip.
type addCallerSkipper interface {
	AddCallerSkip(skip int) Log
}

// AddCallerSkip adds caller skip to log. If the log impl does
// not support additional caller skip, log is returned unchanged.
func AddCallerSkip(log Log, skip int) Log {
	if log == nil {
		return nil
	}

	if skipper, ok := log.(addCallerSkipper); ok {
		log = skipper.AddCallerSkip(skip)
	}

	return log
}

// Discard returns a Log whose methods are no-op.
func Discard() Log {
	return discardLog{}
}

type discardLog struct {
}

func (discardLog) Debug(a ...any) {
}

func (discardLog) Debugf(format string, a ...any) {
}

func (discardLog) Warn(a ...any) {
}

func (discardLog) Warnf(format string, a ...any) {
}

func (discardLog) WarnIfError(err error) {
}

func (discardLog) WarnIfFuncError(fn func() error) {
	if fn != nil {
		_ = fn()
	}
}

func (discardLog) WarnIfCloseError(c io.Closer) {
	if c != nil {
		_ = c.Close()
	}
}

func (discardLog) Error(a ...any) {
}

func (discardLog) Errorf(format string, a ...any) {
}

func (discardLog) With(key string, val any) Log {
	return discardLog{}
}
