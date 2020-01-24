package loglg_test

import (
	"errors"
	"os"
	"testing"

	"github.com/neilotoole/lg"
	"github.com/neilotoole/lg/loglg"
)

var _ lg.Log = (*loglg.Log)(nil)

func TestNew(t *testing.T) {
	log := loglg.New()
	logItAll(log)
}

func TestNewWith(t *testing.T) {
	log := loglg.NewWith(os.Stdout, true, true, true)
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
