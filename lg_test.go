package lg_test

import (
	"errors"
	"testing"

	"github.com/neilotoole/lg"
)

func TestDiscard(t *testing.T) {
	log := lg.Discard()
	logItAll(log)
}

// logItAll executes all the methods of lg.Log.
func logItAll(log lg.Log) {
	log.Debug("Debug msg")
	log.Debugf("Debugf msg")
	log.Warn("Warn msg")
	log.Warnf("Warnf msg")
	log.Error("Error msg")
	log.Errorf("Errorf msg")

	log.WarnIfError(nil)
	log.WarnIfError(errors.New("error: WarnIfError msg"))

	log.WarnIfFuncError(nil)
	log.WarnIfFuncError(func() error { return nil })
	log.WarnIfFuncError(func() error { return errors.New("error: WarnIfFuncError msg") })

	log.WarnIfCloseError(nil)
	log.WarnIfCloseError(errCloser{})
}

type errCloser struct {
}

func (errCloser) Close() error {
	return errors.New("error: WarnIfCloseError msg")
}
