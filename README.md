![Actions Status](https://github.com/neilotoole/lg/workflows/Go/badge.svg)
[![release](https://img.shields.io/badge/release-v2.0.0-green.svg)](https://github.com/neilotoole/errgroup/releases/tag/v2.0.0)
[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://pkg.go.dev/github.com/neilotoole/lg)
[![license](https://img.shields.io/github/license/neilotoole/lg)](./LICENSE)

# neilotoole/lg

`lg` is an exploration of a small, leveled,
unstructured logging interface for enterprise applications.
It is not suggested for production use.

## TLDR

Log levels `ERROR`, `WARN` and `DEBUG` are appropriate for many enterprise applications.

Use this idiom with `io.Closer`:

```go
func DoSomething(log lg.Log) error {
  f, err := os.Open("filename")
  if err != nil {
    return err
  }
  defer log.WarnIfFuncError(f.Close)
```

## Installation

`go get -u github.com/neilotoole/lg/v2`

## Quick Start

```go
// Use uber/zap adapter with options
log := zaplg.NewWith(os.Stdout, "json", true, true, true, true, 0)

// Add a field to the log
log = log.With("request_id", 12345)

log.Debug("hello world")
log.Warnf("not %s at all", "good")
log.Error(err)

log.WarnIfError(f.Close())

// WarnIfFuncError typically used with defer
defer log.WarnIfFuncError(f.Close)

// Alternatively
defer log.WarnIfCloseError(f)
```

When testing:

```go
func TestMe(t *testing.T) {
	log := testlg.New(t)
	log.Debug("Hello world") // directs output to t.Log
}
```


## Overview

Contra [Cheney](https://dave.cheney.net/2015/11/05/lets-talk-about-logging), `lg`'s
primary thesis is that three log levels are appropriate for many 
enterprise applications, and that these levels
should be `ERROR`, `WARN`, and `DEBUG`.

- `ERROR`: a business operation failed, and the user experienced it.
We've lost actual dollars because of this thing. Ops needs to look at it immediately.
- `WARN`: the business operation didn't fail, but something fishy happened
that should be diagnosed before it does start costing us money. Ops needs
to look at it eventually, possibly passing it on to dev.
- `DEBUG`: for developers, for after-the-fact diagnosis of problems. Sometimes `DEBUG` logs are the only practical way for devs to sherlock the disaster that occurred at our primary customer's installation, which of course cannot be recreated in tests because Mephisto himself couldn't conjure this production environment.

This exploration examines a specific issue in detail: how to handle an `io.Closer`
error after the main business operation has succeeded. The conclusion
is that this error is best logged at `WARN` level. The `lg.Log` interface
provides convenience methods for this scenario.

Additionally, `lg` demonstrates the separation of a logging interface
from concrete implementations. Note that `lg` itself doesn't perform rendering
of log entries: this is left to a backing log library. Implementations can be
found in `lg/zaplg` and `lg/testlg`. The `testlg` impl is used in
conjunction with Go's testing framework. If using `zap`, `testlg` has 
[benefits](#zaptest) over `zaptest`.

`lg` does not address structured logging, the virtues of which are outside scope.

## `WarnIf` methods
In addition to typical logging methods such as `Debugf`, the `Log` interface
defines methods `WarnIfError`, `WarnIfFuncError`, and `WarnIfCloseError`.

> **TLDR**
>
> Do this:
> 
> ```go
>   defer log.WarnIfFuncError(dataSource.Close)
> ```
> 
> Not this:
> 
> ```go
>  defer func() {
>    err := dataSource.Close()
>    if err != nil {
>      log.Warn(err)
>    }
>  }()
> ```

We'll use examples to work our way through the origin story of the `WarnIf` methods. 

### `BusinessOperationV1`

Let's start with this function:

```go
// BusinessOperationV1 performs a business operation against
// an external API. If the business operation fails, a non-nil
// error is returned. If the business operation succeeds,
// a non-empty transaction receipt is returned.
//
// BusinessOperationV1 closes dataSource via defer, but ignores
// any error from Close.
func BusinessOperationV1(log lg.Log) (receipt string, err error) {
  dataSource, err := OpenBizData()
  if err != nil {
    return "", err
  }
  defer dataSource.Close() // Ignores any error from Close

  data, err := ioutil.ReadAll(dataSource)
  if err != nil {
    return "", err
  }

  return ExternalAPICall(data) // e.g. book a flight
}
```

`BusinessOperationV1` calls `OpenBizData` which returns an `io.ReadCloser` (could be a file for example), reads data, and then invokes an external API, e.g. to book a flight. The `ExternalAPICall` function returns a transaction receipt or error, which `BusinessOperationV1` returns.

We're concerned with the line `defer dataSource.Close()`.

Although the `Close` method returns an error, it is not checked. Nothing is done with it. Now, being that all the data has already been read from `dataSource`, it is unlikely that the `Close` method will fail, but it could. What to do with a `Close` error? Assume that `ExternalAPICall` has succeeded, and `BusinessOperationV1` can return the transaction receipt to the caller. Should `BusinessOperationV1` return an error and no receipt?

A common judgment is that if the business operation succeeded, the error on `Close` is not worth failing the entire operation for. Plus, then we'd be in the position of potentially having to roll back the effect of `ExternalAPICall`, which simply may not be possible.

We could return the successful receipt and the `Close` error, but that's counter to the Go idiom that if an error is returned, the other return items should be the zero value.

In many cases, the usual handling is simply not to check the `Close` error (`defer dataSource.Close()` reads nicely as well). But this error should not be ignored. It is symptomatic of some underlying issue, and it should be investigated, even it's not causing a business operation to fail (yet).

A quick reminder of how `lg` chooses to define log levels:

- `ERROR`: a business operation failed.
- `WARN`: the business operation succeeded, but something worrying happened.
- `DEBUG`: for devs, for after-the-fact diagnosis of problems.

This `Close` error seems an ideal candidate to log at `WARN` level.

### `BusinessOperationV2`

This is the next iteration of our function:

```go
// BusinessOperationV2 closes dataSource in a defer,
// and logs at WARN level if an error results from Close.
func BusinessOperationV2(log lg.Log) (receipt string, err error) {
  dataSource, err := OpenBizData()
  if err != nil {
    return "", err
  }
  defer func() {
    closeErr := dataSource.Close()
    if closeErr != nil {
      log.Warnf(closeErr.Error())
    }
  }()
    
  // rest of function omitted 
```

This achieves our goals. The `Close` error is logged at `WARN` level. We could call it a day here and go home. However, there's no question that:

```go
defer func() {
  closeErr := dataSource.Close()
  if closeErr != nil {
    log.Warn(closeErr)
  }
}()
```

is less pleasant to read than:

```go
defer dataSource.Close()
```

We can do better.

### `BusinessOperationV3`

In `BusinessOperationV3`, we make the `defer` tidier by using `Log.WarnIfError`.

```go
// WarnIfError is no-op if err is nil; if non-nil, err
// is logged at WARN level.
WarnIfError(err error)
```

Here's how it looks:

```go
// BusinessOperationV3 uses WarnIfError to make the defer statement
// more succinct.
func BusinessOperationV3(log lg.Log) (receipt string, err error) {
  dataSource, err := OpenBizData()
  if err != nil {
    return "", err
  }
  defer func() {
    log.WarnIfError(dataSource.Close())
  }()

  // rest of function omitted   
```

That `defer` looks significantly cleaner now. We could even
write it on one line:

```go
defer func() { log.WarnIfError(dataSource.Close()) }()
```


### `BusinessOperationV4`

But we can get cleaner yet. Here's `Log.WarnIfFuncError`:

```go
// WarnIfFuncError is no-op if fn is nil; if fn is non-nil,
// fn is executed and if fn's error is non-nil, that error
// is logged at WARN level.
WarnIfFuncError(fn func() error)
```

In practice, this reads nicely:

```go
// BusinessOperationV4 uses WarnIfFuncError to make the defer statement
// yet more succinct.
func BusinessOperationV4(log lg.Log) (receipt string, err error) {
  dataSource, err := OpenBizData()
  if err != nil {
    return "", err
  }
  defer log.WarnIfFuncError(dataSource.Close)
  
  // rest of function omitted 
```	

As a variation when `dataSource` can be `nil`, we could use `WarnIfCloseError`:

```go
// WarnIfCloseError is no-op if c is nil; if c is non-nil,
// c.Close is executed and if Close's error is non-nil,
// that error is logged at WARN level.
//
// WarnIfCloseError is preferred to WarnIfFuncError when c may be nil.
//
//  var c io.Closer = nil
//  log.WarnIfCloseError(c)    // ok
//  log.WarnIfFuncError(c.Close) // panic
WarnIfCloseError(c io.Closer)
```	

And invoke like so:

```go
defer log.WarnIfCloseError(dataSource)
```

## `testlg` adapter for `testing`

Package `testlg` provides a `lg.Log` implementation that can output its
log entries to `testing.T`. This test:

```go
func TestNew(t *testing.T) {
  log := testlg.New(t)
  log.Debugf("hello")
  log.Warnf("hola")
  log.Errorf("jambo")
}
```

outputs:

```
=== RUN   TestNew
--- PASS: TestNew (0.00s)
    testlg_test.go:22: 22:12:15.668336 DEBUG  hello
    testlg_test.go:23: 22:12:15.668469 WARN   hola
    testlg_test.go:24: 22:12:15.668474 ERROR  jambo
```

### <a name="zaptest"></a> Prefer `testlg` to `zaptest`

If you're using Uber's [`zap`](https://github.com/uber-go/zap) as your logging impl, you'll
have noticed that pkg `zaptest` provides an adapter for use with `testing`.
Alas, `zaptest` has one ugly drawback: it causes `testing` to
output incorrect caller information. The test output from  [`TestZapTestVsTestLg`](zaplg/zaplg_test.go#L45) demonstrates the issue (edited for brevity):


```
=== RUN   TestZapTestVsTestLg
--- PASS: TestZapTestVsTestLg (0.00s)
    zaplg_test.go:68: zaptest -- Observe the clashing caller info reported by the testing framework (misleading) vs zap itself (desired)
    logger.go:130:  DEBUG  zaplg/zaplg_test.go:70  misleading caller info
    zaplg_test.go:74: testlg -- Observe the concurring caller info reported by the testing framework and zap itself
    zaplg_test.go:79:  DEBUG  zaplg/zaplg_test.go:79  accurate caller info
```

Note `line 4`, where `testing` reports `logger.go:130` as the caller, while `zap` itself
reports the correct caller (`zaplg/zaplg_test.go:70`). In contrast, `testlg` causes `testing`
to accurately report the caller.

## Acknowledgements

- [golang-ci](https://golangci-lint.run) team, for the linter.
- The Uber team, for [zap](https://github.com/uber-go/zap).
- [marwan-at-work](https://github.com/marwan-at-work), for the [mod](https://github.com/marwan-at-work/mod) tool to easily upgrade to `v2+`.
