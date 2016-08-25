/*
Package lg is a yet another simple logging package, intended primarily for code
debugging/tracing purposes. It outputs in Apache httpd error log format.

        lg.Debugf("the answer is: %d", 42)
        // results in
        I [24/Aug/2016:20:26:41 -0600] [example.go:13:example.MyFunction] the answer is: 42

By default, lg outputs to stdout/stderr, but you can specify an alternative
destination using lg.Use(). You can use lg.Levels() to specify which log levels
to produce output for.

See https://github.com/neilotoole/go-lg for more information.
*/
package lg

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Level represents log levels.
type Level string

const (
	LevelError Level = "E"
	LevelDebug Level = "I"
)

var lvls = []Level{LevelDebug, LevelError}

// Levels specifies the set of levels to output. The default is all levels.
// To disable logging, invoke  this function with no parameters. Note that you
// must explicitly specify each level that you desire output for.
func Levels(levels ...Level) {
	lvls = levels
}

// apacheFormat is the standard apache timestamp format.
const apacheFormat = `02/Jan/2006:15:04:05 -0700`

var wOut io.Writer = os.Stdout
var wErr io.Writer = os.Stderr
var mu sync.Mutex

// LongFnName determines whether the full path/to/pkg.func is used. Default is pkg.func.
var LongFnName = false

// LongFilePath determines whether the full /path/to/file.go is used. Default is file.go.
var LongFilePath = false

// Use specifies the log output destination. The default is os.Stdout/os.Stderr.
func Use(dest io.Writer) {
	mu.Lock()
	defer mu.Unlock()
	wOut = dest
	wErr = dest
}

// Debugf logs an information message.
func Debugf(format string, v ...interface{}) {
	log(false, 1, LevelDebug, format, v...)
}

// DebugfN is similar to Debugf, but it allows the caller to specify additional
// call depth. This is useful, for example, in situations where a utility function
// is logging on behalf of its parent function.
func DebugN(calldepth int, format string, v ...interface{}) {
	log(false, 1+calldepth, LevelDebug, format, v...)
}

// Errorf logs an error message.
func Errorf(format string, v ...interface{}) {
	log(false, 1, LevelError, format, v...)
}

// ErrorfN is similar to Errof, but it allows the caller to specify additional
// call depth. This is useful, for example, in situations where a utility function
// is logging on behalf of its parent function.
func ErrorfN(calldepth int, format string, v ...interface{}) {
	log(false, 1+calldepth, LevelError, format, v...)
}

// Fatalf is similar to Errorf, but calls os.Exit(1) after logging the message.
// Additionally, if the log destination is not os.Stdout or os.Stderr, then
// the message is also printed to os.Stderr.
func Fatalf(format string, v ...interface{}) {

	mu.Lock()
	defer mu.Unlock()
	msg := fmt.Sprintf(format, v...)
	log(true, 1, LevelError, msg)

	if wOut != os.Stdout && wOut != os.Stderr {
		fmt.Fprintln(os.Stderr, msg)
	}
	os.Exit(1)
}

func log(locked bool, calldepth int, level Level, format string, v ...interface{}) {

	isLevelEnabled := false
	for _, lvl := range lvls {
		if level == lvl {
			isLevelEnabled = true
			break
		}
	}

	if !isLevelEnabled {
		return
	}

	t := time.Now()
	pc := make([]uintptr, 10) // at least 1 entry needed
	runtime.Callers(2+calldepth, pc)
	fnName := runtime.FuncForPC(pc[0]).Name()
	if !LongFnName {
		parts := strings.Split(fnName, "/")
		fnName = parts[len(parts)-1]
	}
	_, file, line, ok := runtime.Caller(calldepth + 1)
	if !ok {
		file = "???"
		line = 0
	} else if !LongFilePath {
		// We just want the file name, not the whole path
		parts := strings.Split(file, "/")
		file = parts[len(parts)-1]
	}

	stamp := t.Format(apacheFormat)
	// E [08/Jun/2013:11:28:58 -0700] [ql.go:60] ql.ToSQL: my message text
	tpl := `%s [%s] [%s:%d:%s] %s`
	str := fmt.Sprintf(tpl, level, stamp, file, line, fnName, fmt.Sprintf(format, v...))
	if !locked {
		mu.Lock()
		defer mu.Unlock()
	}

	if level == LevelError {
		fmt.Fprintln(wErr, str)
		return
	}
	fmt.Fprintln(wOut, str)
}
