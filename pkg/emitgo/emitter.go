package emitgo

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/podhmo/apikit/pkg/emitfile"
	"github.com/podhmo/apikit/pkg/tinypkg"
)

type Config struct {
	RootPkg        *tinypkg.Package
	FilenamePrefix string

	*emitfile.Config
}

func NewConfigFromRelativePath(fn interface{}, relative string) *Config {
	rootdir := filepath.Join(DefinedDir(fn), relative)
	rootpkg := tinypkg.NewPackage(path.Join(PackagePath(fn), relative), "")
	return NewConfig(rootdir, rootpkg)
}

func NewConfig(rootdir string, rootpkg *tinypkg.Package) *Config {
	return &Config{
		RootPkg: rootpkg,
		Config:  emitfile.NewConfig(rootdir),
	}
}

func (c *Config) NewEmitter() *Emitter {
	emitter := &Emitter{
		FileEmitter: c.Config.NewEmitter(),
		Config:      c,
	}
	emitter.FileEmitter.PathResolver.AddRoot("/"+c.RootPkg.Path, c.RootDir)
	return emitter
}

func (c *Config) EmitWith(fn func(*Emitter) error) (err error) {
	emitter := c.NewEmitter()
	defer emitter.EmitWith(&err)
	return fn(emitter)
}
func (c *Config) MustEmitWith(fn func(*Emitter) error) {
	if err := c.EmitWith(fn); err != nil {
		panic(fmt.Sprintf("!! %+v", err))
	}
}

type Emitter struct {
	*Config
	FileEmitter *emitfile.Executor
}

func (e *Emitter) EmitWith(errptr *error) {
	if err := e.FileEmitter.Emit(); err != nil {
		*errptr = err
	}
}
func (e *Emitter) Emit() error {
	return e.FileEmitter.Emit()
}

func (e *Emitter) Register(pkg *tinypkg.Package, name string, target emitfile.Emitter) *emitfile.EmitAction {
	if !strings.HasSuffix(name, ".go") {
		name = name + ".go"
	}
	if e.Config.FilenamePrefix != "" {
		name = e.Config.FilenamePrefix + name
	}
	return e.FileEmitter.Register("/"+path.Join(pkg.Path, name), target)
}
