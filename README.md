# lg: golang loging library
`lg` is yet another golang logging package, primarily intended for debugging/tracing purposes. 
It outputs in Apache httpd `error_log` format. To make it easy to debug, it logs the the
invoking file name, line number, and function name.

To get the library: `go get github.com/neilotoole/go-lg/lg`

Here's how you use it:

```go
package example

import "github.com/neilotoole/go-lg/lg"

func ShowMe() {
        lg.Debugf("the answer is: %d", 42)
}
```
produces:

```
I [24/Aug/2016:20:26:41 -0600] [example.go:6:example.ShowMe] the answer is: 42
```

By default, lg outputs to `stdout`/`stderr`, but you can specify an alternative
destination with `lg.Use()`. Typically this is a log file, and your code might
look something like this:

```go
package main

import (
	"os"
	"github.com/neilotoole/go-lg/lg"
)

func init() {
	logFile, err := os.OpenFile("/path/to/file.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	lg.Use(logFile)
}

func main() {
	lg.Debugf("Hello world!")
	// do something useful
}

```

You can control which log levels produce output:

```go
lg.Levels(lg.LevelError) // only output error
lg.Levels() // output no levels, aka disable logging
```

If you want to output the full file path or full function name, set these variables:

```go
lg.LongFnName = true
lg.LongFilePath = true
````

You'll get crazy long output like this:

```
I [24/Aug/2016:20:34:02 -0600] [/Users/neilotoole/nd/go/src/github.com/neilotoole/go-lg/example/example.go:6:github.com/neilotoole/go-lg/example/example.ShowMe] the answer is: 42
```

Note that `lg` only actually outputs two Apache log levels: `INFO` and `ERROR`.
The `Debug` functions map to `INFO`, and the other functions map to `ERROR`. This
is somewhat in the spirit of Dave Cheney's [logging article](http://dave.cheney.net/2015/11/05/lets-talk-about-logging).
But the Apache `error_log` format doesn't have a `DEBUG` level, and `WARN` is a little useless,
so this is where we're at.