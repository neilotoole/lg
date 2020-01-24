package testlg_test

import (
	"errors"
	"io"
	"testing"

	"github.com/neilotoole/lg"
	"github.com/neilotoole/lg/testlg"
	"github.com/neilotoole/lg/zaplg"
)

var _ lg.Log = (*testlg.Log)(nil)

func TestNew(t *testing.T) {
	log := testlg.New(t)
	logItAll(log)
}

func TestNewWith(t *testing.T) {
	log := testlg.NewWith(t, testlg.FactoryFn)
	logItAll(log)
}

func TestFactoryFn(t *testing.T) {
	log := testlg.New(t)
	logItAll(log)

	prevFn := testlg.FactoryFn
	defer func() { testlg.FactoryFn = prevFn }()

	testlg.FactoryFn = func(w io.Writer) lg.Log {
		return zaplg.NewWith(w, "text", true, 1)
	}

	t.Log("Switching to new testlg.FactoryFn")
	log = testlg.New(t)
	logItAll(log)
}

// logItAll executes all the methods of lg.Log.
func logItAll(log lg.Log) {
	log.Debugf("Debugf")
	log.Warnf("Warnf")
	log.Errorf("Errorf")

	log.WarnIfError(nil)
	log.WarnIfError(errors.New("WarnIfError"))

	log.WarnIfFnError(nil)
	fn := func() error {
		return nil
	}
	log.WarnIfFnError(fn)
	fn = func() error {
		return errors.New("WarnIfFnError")
	}
	log.WarnIfFnError(fn)
}
