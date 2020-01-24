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

func TestNew_Caller(t *testing.T) {
	log := zaplg.NewWith(os.Stdout, "text", true, 0)
	logItAll(log)
}
func TestNew_NoCaller(*testing.T) {
	log := zaplg.NewWith(os.Stdout, "text", false, 0)
	logItAll(log)
}

func TestNewWith(t *testing.T) {
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
		return zaplg.NewWith(w, "text", true, 1)
	}
	log := testlg.NewWith(t, factoryFn)
	log.Debugf("accurate caller info")
	log.Warnf("accurate caller info")
	log.Errorf("accurate caller info")
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
