package loglg_test

import (
	"bufio"
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/neilotoole/lg"
	"github.com/neilotoole/lg/loglg"
)

var _ lg.Log = (*loglg.Log)(nil)

func TestNew(t *testing.T) {
	var lineParts = [][]string{
		{"loglg_test.go:", "DEBUG", "Debugf"},
		{"loglg_test.go:", "WARN", "Warnf"},
		{"loglg_test.go:", "ERROR", "Errorf"},
		{"loglg_test.go:", "WARN", "WarnIfError"},
		{"loglg_test.go:", "WARN", "WarnIfFnError"},
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
