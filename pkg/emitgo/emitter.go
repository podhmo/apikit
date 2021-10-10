package emitgo

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/podhmo/apikit/pkg/emitfile"
	"github.com/podhmo/apikit/pkg/tinypkg"
)

type Config struct {
	RootPkg *tinypkg.Package
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

type Emitter struct {
	*Config
	FileEmitter *emitfile.Executor
}

func New(config *Config) *Emitter {
	emitter := &Emitter{
		FileEmitter: emitfile.New(config.Config),
		Config:      config,
	}
	emitter.FileEmitter.PathResolver.AddRoot("/"+config.RootPkg.Path, config.RootDir)
	return emitter
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
	return e.FileEmitter.Register("/"+path.Join(pkg.Path, name), target)
}
