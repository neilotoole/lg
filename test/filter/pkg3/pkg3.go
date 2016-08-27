package pkg3


import "github.com/neilotoole/go-lg/lg"

func LogDebug() {
	lg.Debugf("doing some logging for pkg3")
}

func LogError() {
	lg.Errorf("doing some error logging for pkg3")
}