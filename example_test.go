package lg_test

import (
	"os"

	"github.com/neilotoole/lg/zaplg"
)

// Demonstrate use with uber/zap.
func Example_zap() {
	// Default setup
	// log := zaplg.New()

	// With options
	log := zaplg.NewWith(os.Stdout, "text", false, true, true, 0)
	log.Debug("Hello", "World")
	// Output: DEBUG	lg/example_test.go:16:Example_zap	HelloWorld
}
