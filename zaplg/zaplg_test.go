package zaplg_test

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/neilotoole/lg/v2"
	"github.com/neilotoole/lg/v2/testlg"
	"github.com/neilotoole/lg/v2/zaplg"
)

var _ lg.Log = (*zaplg.Log)(nil)

func TestNew(t *testing.T) {
	log := zaplg.New()
	logItAll(log)
}

func TestNewWith(t *testing.T) {
	// TestNewWith doesn't actually test the log output, only
	// verifies that the various input arg combinations don't
	// blow it up.
	testCases := []struct {
		format    string
		timestamp bool
		level     bool
		caller    bool
	}{
		{format: "text", timestamp: true, level: true, caller: true},
		{format: "text", timestamp: true, level: true, caller: false},
		{format: "text", timestamp: true, level: false, caller: true},
		{format: "text", timestamp: true, level: false, caller: false},
		{format: "text", timestamp: false, level: true, caller: true},
		{format: "text", timestamp: false, level: true, caller: false},
		{format: "text", timestamp: false, level: false, caller: true},
		{format: "text", timestamp: false, level: false, caller: false},

		{format: "json", timestamp: true, level: true, caller: true},
		{format: "json", timestamp: true, level: true, caller: false},
		{format: "json", timestamp: true, level: false, caller: true},
		{format: "json", timestamp: true, level: false, caller: false},
		{format: "json", timestamp: false, level: true, caller: true},
		{format: "json", timestamp: false, level: true, caller: false},
		{format: "json", timestamp: false, level: false, caller: true},
		{format: "json", timestamp: false, level: false, caller: false},
	}

	for _, tc := range testCases {
		tc := tc

		name := fmt.Sprintf("%s__timestamp_%v__level_%v__caller_%v", tc.format, tc.timestamp, tc.level, tc.caller)
		t.Run(name, func(t *testing.T) {
			log := testlg.NewWith(t, func(w io.Writer) lg.Log {
				return zaplg.NewWith(w, tc.format, tc.timestamp, true, tc.level, tc.caller, 1)
			})

			logItAll(log)
		})
	}
}

func TestTestingFactoryFn(t *testing.T) {
	log := testlg.NewWith(t, zaplg.TestingFactoryFn)
	logItAll(log)
}

// TestZapTestVsTestLg demonstrates the incorrect
// caller info reported by the testing framework when
// using zaptest as opposed to testlg.
func TestZapTestVsTestLg(t *testing.T) {
	t.Log(`zaptest -- Observe the clashing caller info reported by
the testing framework (misleading) vs zap itself (desired)`)
	zlog := zaptest.NewLogger(t, zaptest.WrapOptions(zap.AddCaller()))
	zlog.Debug("misleading caller info")
	zlog.Warn("misleading caller info")
	zlog.Error("misleading caller info")

	t.Log("testlg -- Observe the concurring caller info reported by the testing framework and zap itself")
	factoryFn := func(w io.Writer) lg.Log {
		return zaplg.NewWith(w, "text", true, true, true, true, 1)
	}
	log := testlg.NewWith(t, factoryFn)
	log.Debug("accurate caller info")
	log.Warn("accurate caller info")
	log.Error("accurate caller info")
}

// logItAll executes all the methods of lg.Log.
func logItAll(log lg.Log) {
	log.Debug("Debug msg")
	log.Warn("Warn msg")
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
