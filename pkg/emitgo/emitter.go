package emitgo

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/podhmo/apikit/pkg/emitfile"
	"github.com/podhmo/apikit/pkg/tinypkg"
)

type Emitter struct {
	FileEmitter *emitfile.Executor
	RootPkg     *tinypkg.Package
	RootDir     string
}

func NewFromRelativePath(fn interface{}, relative string) *Emitter {
	rootdir := filepath.Join(DefinedDir(fn), relative)
	rootpkg := tinypkg.NewPackage(path.Join(PackagePath(fn), relative), "")
	return New(rootdir, rootpkg)
}
func New(rootdir string, rootpkg *tinypkg.Package) *Emitter {
	emitter := &Emitter{
		FileEmitter: emitfile.New(rootdir),
		RootPkg:     rootpkg,
		RootDir:     rootdir,
	}
	emitter.FileEmitter.PathResolver.AddRoot("/"+rootpkg.Path, rootdir)
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
