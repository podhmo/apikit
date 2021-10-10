package main

import (
	"fmt"
	"io"
	"path/filepath"
	"reflect"
	"runtime"

	"github.com/podhmo/apikit/pkg/emitfile"
)

func main() {
	rootdir := DefinedDir(main)
	config := emitfile.NewConfig(rootdir)
	emitter := emitfile.New(config)
	defer emitter.Emit()

	emitter.Register("docs/spring.md", &Content{Text: "# 春"})
	emitter.Register("docs/summer.md", &Content{Text: "# 夏"})
	emitter.Register("docs/fall.md", &Content{Text: "# 秋"})
	emitter.Register("docs/winter.md", &Content{Text: "# 冬"})
}

type Content struct {
	Text string
}

func (c *Content) Emit(w io.Writer) error {
	fmt.Fprintln(w, c.Text)
	return nil
}

var _ emitfile.Emitter = &Content{}

func DefinedDir(fn interface{}) string {
	return filepath.Dir(DefinedFile(fn))
}
func DefinedFile(fn interface{}) string {
	rfunc := runtime.FuncForPC(reflect.ValueOf(fn).Pointer())
	fpath, _ := rfunc.FileLine(rfunc.Entry())
	return fpath
}
