package lg_test

import (
	"os"

	"github.com/neilotoole/lg/loglg"
	"github.com/neilotoole/lg/zaplg"
)

// Demonstrate use with stdlib log.
func Example_stdlibLog() {
	// Default setup
	// log := loglg.New()

	// With options
	log := loglg.NewWith(os.Stdout, false, true, true)
	log.Debug("Hello", "World")
	// Output: example_test.go:17: 	DEBUG	HelloWorld
}

// Demonstrate use with uber/zap.
func Example_zap() {
	// Default setup
	// log := zaplg.New()

	// With options
	log := zaplg.NewWith(os.Stdout, "text", false, true, true, 0)
	log.Debug("Hello", "World")
	// Output: DEBUG	lg/example_test.go:28	HelloWorld
}
