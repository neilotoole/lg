package loglg_test

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

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

func TestOutput(t *testing.T) {
	var lineParts = [][]string{
		{"loglg_test.go:", "DEBUG", "Debug msg"},
		{"loglg_test.go:", "DEBUG", "Debugf msg"},
		{"loglg_test.go:", "WARN", "Warn msg"},
		{"loglg_test.go:", "WARN", "Warnf msg"},
		{"loglg_test.go:", "ERROR", "Error msg"},
		{"loglg_test.go:", "ERROR", "Errorf msg"},
		{"loglg_test.go:", "WARN", "WarnIfError msg"},
		{"loglg_test.go:", "WARN", "WarnIfFuncError msg"},
		{"loglg_test.go:", "WARN", "WarnIfCloseError msg"},
	}

	testCases := []struct {
		name   string
		level  bool
		caller bool
	}{
		{name: "level_caller", level: true, caller: true},
		{name: "level_no_caller", level: true, caller: false},
		{name: "no_level_caller", level: false, caller: true},
		{name: "no_level_no_caller", level: false, caller: false},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			log := loglg.NewWith(buf, false, tc.level, tc.caller)
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
					require.Contains(t, gotLine, lineParts[i][0], "caller should be printed")
				} else {
					require.NotContains(t, gotLine, lineParts[i][0], "caller should not be printed")
				}

				if tc.level {
					require.Contains(t, gotLine, lineParts[i][1], "level should be printed")
				} else {
					require.NotContains(t, gotLine, lineParts[i][1], "level should not be printed")
				}

				require.Contains(t, gotLine, lineParts[i][2], "log msg should be printed")
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
