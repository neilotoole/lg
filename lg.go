// Package lg is an exploration of a minimal, leveled,
// unstructured logging interface for enterprise developers.
// It is not suggested for production use.
//
// This pkg does not itself perform logging. A concrete impl
// must be used, e.g. loglg.New or zaplg.New. Use testlg.New
// to adapt lg to output to a testing.T.
package lg

import "io"

// Log is a logging interface that adds WarnIf methods
// to the basic Debug, Warn and Error methods.
//
// Style note: Being that Log is an interface, idiomatically
// the type name should be Logger. But the sense is that
// Orwell's sixth rule prevails here. It's not that lg.Logger
// is barbaric in and of itself, but a thousand appearances
// of lg.Logger vs lg.Log sells constitutes a horde.
type Log interface {
	// Debug logs at DEBUG level.
	Debug(a ...interface{})

	// Debugf logs at DEBUG level.
	Debugf(format string, a ...interface{})

	// Warn logs at WARN level.
	Warn(a ...interface{})

	// Warnf logs at WARN level.
	Warnf(format string, a ...interface{})

	// Error logs at ERROR level.
	Error(a ...interface{})

	// Errorf logs at ERROR level.
	Errorf(format string, a ...interface{})

	// WarnIfError is no-op if err is nil; if non-nil, err
	// is logged at WARN level.
	WarnIfError(err error)

	// WarnIfFnError is no-op if fn is nil; if fn is non-nil,
	// fn is executed and if fn's error is non-nil, that error
	// is logged at WARN level.
	WarnIfFnError(fn func() error)

	// WarnIfCloseError is no-op if c is nil; if c is non-nil,
	// c.Close is executed and if Close's error is non-nil,
	// that error is logged at WARN level.
	//
	// WarnIfCloseError is preferred to WarnIfFnError
	// when c may be nil.
	//
	//  var c io.Closer = nil
	//  log.WarnIfCloseError(c)    // ok
	//  log.WarnIfFnError(c.Close) // panic
	WarnIfCloseError(c io.Closer)
}

// Discard returns a Log whose methods are no-op.
func Discard() Log {
	return discardLog{}
}

type discardLog struct {
}

func (discardLog) Debug(a ...interface{}) {
}

func (discardLog) Debugf(format string, a ...interface{}) {
}

func (discardLog) Warn(a ...interface{}) {
}

func (discardLog) Warnf(format string, a ...interface{}) {
}

func (discardLog) Error(a ...interface{}) {
}

func (discardLog) Errorf(format string, a ...interface{}) {
}

func (discardLog) WarnIfError(err error) {
}

func (discardLog) WarnIfFnError(fn func() error) {
	if fn != nil {
		_ = fn()
	}
}

func (discardLog) WarnIfCloseError(c io.Closer) {
	if c != nil {
		_ = c.Close()
	}
}
