package zaplg_test

import (
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/neilotoole/lg"
	"github.com/neilotoole/lg/testlg"
	"github.com/neilotoole/lg/zaplg"
)

var _ lg.Log = (*zaplg.Log)(nil)

func TestNew(t *testing.T) {
	log := zaplg.New()
	logItAll(log)
}
func TestNewWith_Caller(t *testing.T) {
	log := zaplg.NewWith(os.Stdout, "text", true, true, 0)
	logItAll(log)
}

func TestNewWith_NoCaller(*testing.T) {
	log := zaplg.NewWith(os.Stdout, "text", true, false, 0)
	logItAll(log)
}

func TestNewWithZap(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	log := zaplg.NewWithZap(logger)
	logItAll(log)
}

func TestTestingFactoryFn(t *testing.T) {
	log := testlg.NewWith(t, zaplg.TestingFactoryFn)
	logItAll(log)
}

// TestZapTestVsTestLg demonstrates the incorrect
// caller info reported by the testing framework when
// using zaptest as opposed to testlg.
func TestZapTestVsTestLg(t *testing.T) {
	t.Log("zaptest -- Observe the clashing caller info reported by the testing framework (misleading) vs zap itself (desired)")
	zlog := zaptest.NewLogger(t, zaptest.WrapOptions(zap.AddCaller()))
	zlog.Debug("misleading caller info")
	zlog.Warn("misleading caller info")
	zlog.Error("misleading caller info")

	t.Log("testlg -- Observe the concurring caller info reported by the testing framework and zap itself")
	factoryFn := func(w io.Writer) lg.Log {
		return zaplg.NewWith(w, "text", true, true, 1)
	}
	log := testlg.NewWith(t, factoryFn)
	log.Debugf("accurate caller info")
	log.Warnf("accurate caller info")
	log.Errorf("accurate caller info")
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

	log.WarnIfFnError(nil)
	log.WarnIfFnError(func() error { return nil })
	log.WarnIfFnError(func() error { return errors.New("error: WarnIfFnError msg") })

	log.WarnIfCloseError(nil)
	log.WarnIfCloseError(errCloser{})
}

type errCloser struct {
}

func (errCloser) Close() error {
	return errors.New("error: WarnIfCloseError msg")
}
