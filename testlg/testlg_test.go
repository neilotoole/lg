package testlg_test

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/neilotoole/lg/v2"
	"github.com/neilotoole/lg/v2/testlg"
	"github.com/neilotoole/lg/v2/zaplg"
)

var _ lg.Log = (*testlg.Log)(nil)

func TestNew(t *testing.T) {
	log := testlg.New(t)
	logItAll(log)
}

func TestNewWith(t *testing.T) {
	log := testlg.NewWith(t, testlg.FactoryFn)
	logItAll(log)

	t.Log("Switching to new testlg.FactoryFn")
}

func TestNewWith_Zap(t *testing.T) {
	log := testlg.NewWith(t, zaplg.TestingFactoryFn)
	logItAll(log)
}

func TestFactoryFn(t *testing.T) {
	prevFn := testlg.FactoryFn
	defer func() { testlg.FactoryFn = prevFn }()

	testlg.FactoryFn = func(w io.Writer) lg.Log {
		return zaplg.NewWith(w, time.RFC3339, true, true, true, false, 0)
	}

	log := testlg.New(t)
	logItAll(log)

	testlg.FactoryFn = func(w io.Writer) lg.Log {
		return zaplg.NewWith(w, "test", true, true, true, true, 1)
	}

	t.Log("Switching to new testlg.FactoryFn")
	log = testlg.New(t) // should pick up the zap impl from testlg.FactoryFn
	logItAll(log)
}

// logItAll executes all the methods of lg.Log.
func logItAll(log lg.Log) {
	log.Debug("Debug msg")
	log.Debug("Debug msg")
	log.Warn("Warn msg")
	log.Warn("Warn msg")
	log.Error("Error msg")
	log.Error("Error msg")

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
