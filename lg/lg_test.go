package lg_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/neilotoole/go-lg/lg"
	"github.com/neilotoole/go-lg/test/filter/pkg1"
	"github.com/neilotoole/go-lg/test/filter/pkg2"
	"github.com/neilotoole/go-lg/test/filter/pkg3"
	"github.com/stretchr/testify/assert"
)

func TestFilters(t *testing.T) {

	buf := useNewLgBuf()
	logPackages()
	assert.Equal(t, 6, countLines(buf))

	buf = useNewLgBuf()
	lg.Levels(lg.LevelDebug)
	logPackages()
	assert.Equal(t, 3, countLines(buf))

	buf = useNewLgBuf()
	lg.Levels()
	logPackages()
	assert.Equal(t, 0, countLines(buf))

	buf = useNewLgBuf()
	lg.Levels(lg.LevelAll)
	logPackages()
	assert.Equal(t, 6, countLines(buf), "levels should be reset to all")

	buf = useNewLgBuf()
	lg.Exclude("github.com/neilotoole/go-lg/test/filter/pkg1")
	logPackages()
	assert.Equal(t, 4, countLines(buf))

	buf = useNewLgBuf()
	lg.Exclude("github.com/neilotoole/go-lg/test/filter/pkg1", "github.com/neilotoole/go-lg/test/filter/pkg2")
	logPackages()
	assert.Equal(t, 2, countLines(buf))

	buf = useNewLgBuf()
	lg.Exclude("github.com/neilotoole/go-lg/test/filter/pkg1", "github.com/neilotoole/go-lg/test/filter/pkg2", "github.com/neilotoole/go-lg/test/filter/pkg3")
	logPackages()
	assert.Equal(t, 0, countLines(buf))

	buf = useNewLgBuf()
	lg.Exclude("github.com/neilotoole/go-lg/test/filter")
	logPackages()
	assert.Equal(t, 0, countLines(buf), "all sub-packages should have been excluded")

	buf = useNewLgBuf()
	lg.Excluded = nil
	logPackages()
	assert.Equal(t, 6, countLines(buf), "should have reset all pkg filters")

	buf = useNewLgBuf()
	lg.Disable()
	logPackages()
	assert.Equal(t, 0, countLines(buf), "logging should be entirely disabled")

	buf = useNewLgBuf()
	lg.Enable()
	logPackages()
	assert.Equal(t, 6, countLines(buf), "logging should be re-enabled")
}

func countLines(buf *bytes.Buffer) int {
	return strings.Count(buf.String(), "\n")
}

func useNewLgBuf() *bytes.Buffer {
	buf := &bytes.Buffer{}
	lg.Use(buf)
	return buf
}

func logPackages() {

	pkg1.LogDebug()
	pkg1.LogError()
	pkg2.LogDebug()
	pkg2.LogError()
	pkg3.LogDebug()
	pkg3.LogError()
}
