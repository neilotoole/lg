package main

import (
	"fmt"

	"github.com/neilotoole/go-lg/lg"
)

func main() {

	lg.Errorf("SHOULD print")

	//lg.Levels(lg.LevelDebug)

	fmt.Printf("\n\n\n========\n\n\n")

	lg.Disable()
	lg.Errorf("NONONO print")
	//lg.Debugf("should print")
}
