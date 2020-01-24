// Package lg is an exploration of a minimal, leveled,
// unstructured logging interface for enterprise developers.
// It is not suggested for production use.
//
// This pkg does not itself perform logging; a concrete impl
// must be used, e.g. loglg.New or zaplg.New.
package lg

// Log is a logging interface that adds methods WarnIfError
// and WarnIfFnError to the basic Debugf, Warnf and Errorf
// methods.
//
// For brevity, Log omits Debug, Warn and Error methods as
// the Debugf, Warnf and Errorf cousins are sufficient for
// this exploration.
//
// Style note: Being that Log is an interface, idiomatically
// the type name should be Logger. But the sense is that
// Orwell's sixth rule prevails here. It's not that lg.Logger
// is barbaric in and of itself, but a thousand appearances
// of lg.Logger vs lg.Log sells the brevity.
type Log interface {
	// Debugf logs at DEBUG level.
	Debugf(format string, a ...interface{})

	// Warnf logs at WARN level.
	Warnf(format string, a ...interface{})

	// Errorf logs at ERROR level.
	Errorf(format string, a ...interface{})

	// WarnIfError is no-op if err is nil; if non-nil, err
	// is logged at WARN level.
	WarnIfError(err error)

	// WarnIfFnError is no-op if fn is nil; if fn is non-nil,
	// fn is executed and if fn's error is non-nil, that error
	// is logged at WARN level.
	WarnIfFnError(fn func() error)
}

// Discard returns a Log whose methods are no-op.
func Discard() Log {
	return discardLog{}
}

type discardLog struct {
}

func (discardLog) Debugf(format string, a ...interface{}) {
}

func (discardLog) Warnf(format string, a ...interface{}) {
}

func (discardLog) Errorf(format string, a ...interface{}) {
}

func (discardLog) WarnIfError(err error) {
}

func (discardLog) WarnIfFnError(fn func() error) {
}
