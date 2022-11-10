package lg_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/neilotoole/lg"
	"github.com/neilotoole/lg/testlg"
	"github.com/neilotoole/lg/zaplg"
)

func TestDiscard(t *testing.T) {
	log := lg.Discard()
	logItAll(log)
}

// TestLog is a smoke test of Log impls. Basically
// the test exists to verify that nothing explodes. The
// test does not verify that the output is correct.
func TestLog(t *testing.T) {
	t.Run("testlg", func(t *testing.T) {
		log := testlg.New(t)
		logItAll(log)
	})

	t.Run("zaplg", func(t *testing.T) {
		buf := &bytes.Buffer{}
		zlog := zaplg.NewWith(buf, "json", true, true, true, 0)
		logItAll(zlog)
		t.Log(buf.String())
	})
}

// TestImplsOutput verifies that the implementations of lg.Log
// output expected log entry text.
func TestImplsOutput(t *testing.T) { //nolint:gocognit
	const filename = "lg_test.go"

	var lineParts = [][]string{
		{"DEBUG", "Debug msg"},
		{"DEBUG", "Debugf msg"},
		{"WARN", "Warn msg"},
		{"WARN", "Warnf msg"},
		{"ERROR", "Error msg"},
		{"ERROR", "Errorf msg"},
		{"WARN", "WarnIfError msg"},
		{"WARN", "error: WarnIfFuncError msg"},
		{"WARN", "error: WarnIfCloseError msg"},
	}

	// testCases are the main configurable params (level and caller)
	// to the log impl constructs. Timestamp param is not tested.
	testCases := []struct {
		level  bool
		caller bool
	}{
		{level: true, caller: true},
		{level: true, caller: false},
		{level: false, caller: true},
		{level: false, caller: false},
	}

	logImpls := []struct {
		name  string
		newFn func(w io.Writer, level, caller bool) lg.Log
	}{
		{"zaplg", func(w io.Writer, level, caller bool) lg.Log {
			return zaplg.NewWith(w, "text", false, level, caller, 0)
		}},
	}

	for _, logImpl := range logImpls {
		logImpl := logImpl

		t.Run(logImpl.name, func(t *testing.T) {
			for _, tc := range testCases {
				tc := tc

				t.Run(fmt.Sprintf("level_%v__caller_%v", tc.level, tc.caller), func(t *testing.T) {
					buf := &bytes.Buffer{}

					log := logImpl.newFn(buf, tc.level, tc.caller)
					logItAll(log)

					sc := bufio.NewScanner(buf)
					var gotLines []string
					for sc.Scan() {
						gotLines = append(gotLines, sc.Text())
					}

					require.NoError(t, sc.Err())
					require.Equal(t, len(lineParts), len(gotLines))

					for i, gotLine := range gotLines {
						if tc.caller {
							require.Contains(t, gotLine, filename, "caller should be printed")
						} else {
							require.NotContains(t, gotLine, filename, "caller should not be printed")
						}

						if tc.level {
							require.Contains(t, gotLine, lineParts[i][0], "level should be printed")
						} else {
							require.NotContains(t, gotLine, lineParts[i][0], "level should not be printed")
						}

						require.Contains(t, gotLine, lineParts[i][1], "log msg should be printed")
					}
				})
			}
		})
	}
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
