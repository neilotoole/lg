package lg_test

import (
	"errors"
	"testing"

	"github.com/neilotoole/lg"
)

func TestDiscard(t *testing.T) {
	log := lg.Discard()
	logItAll(log)
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
