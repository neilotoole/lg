package pkg1


import "github.com/neilotoole/go-lg/lg"

func LogDebug() {
	lg.Debugf("doing some debug logging for pkg1")
}
func LogError() {
	lg.Errorf("doing some error logging for pkg1")
}
