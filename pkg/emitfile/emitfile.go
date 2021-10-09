package emitfile

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

var DEBUG = false
var VERBOSE = false

func init() {
	if v, err := strconv.ParseBool(os.Getenv("DEBUG")); err == nil {
		DEBUG = v
	}
	if v, err := strconv.ParseBool(os.Getenv("VERBOSE")); err == nil {
		VERBOSE = v
	}
}

type Logger interface {
	Printf(fmt string, args ...interface{})
}

type Config struct {
	Verbose     bool
	Debug       bool
	AlwaysWrite bool

	CurDir string

	Log Logger
}

type Executor struct {
	*Config

	PathResolver *PathResolver
	saver        *fileSaver

	Actions []*EmitAction
	// TODO: permission
	// TODO: sort by priority
}
type EmitAction struct {
	Name     string
	Path     string
	Priority int

	Target     Emitter
	FormatFunc func([]byte) ([]byte, error)
}
type EmitFunc func(w io.Writer) error

func (f EmitFunc) Emit(w io.Writer) error {
	return f(w)
}

type Emitter interface {
	Emit(w io.Writer) error
}

func New(rootdir string) *Executor {
	c := &Config{
		Debug:       DEBUG,
		Verbose:     VERBOSE,
		Log:         log.New(os.Stderr, "", 0),
		AlwaysWrite: true,
	}
	if c.Debug {
		c.Verbose = true
		c.AlwaysWrite = true
	}
	if cwd, err := os.Getwd(); err == nil {
		c.CurDir = cwd
	}
	r := newPathResolver(rootdir, c)
	r.Config = c
	return &Executor{
		PathResolver: r,
		saver:        newfileSaver(c),
		Config:       c,
	}
}

func (e *Executor) Register(path string, emitter Emitter) *EmitAction {
	action := &EmitAction{
		Path:   path,
		Target: emitter,
	}
	if impl, ok := emitter.(interface{ Name() string }); ok {
		action.Name = impl.Name()
	}
	if impl, ok := emitter.(interface{ Priority() int }); ok {
		action.Priority = impl.Priority()
	}
	if impl, ok := emitter.(interface{ FormatBytes([]byte) ([]byte, error) }); ok {
		action.FormatFunc = impl.FormatBytes
	}
	e.Actions = append(e.Actions, action)
	return action
}

func (e *Executor) Emit() error {
	// TODO: strategy (failfast, runall)
	// TODO: run once
	e.Log.Printf("emit files ...")
	sort.SliceStable(e.Actions, func(i, j int) bool { return e.Actions[i].Priority < e.Actions[j].Priority })
	for _, action := range e.Actions {
		fpath, err := e.PathResolver.ResolvePath(action.Path)
		if err != nil {
			return fmt.Errorf("resolve-path is failed in action=%q: %w", action.Name, err)
		}

		buf := new(bytes.Buffer)
		if err := action.Target.Emit(buf); err != nil {
			return fmt.Errorf("emit-func is failed in action=%q: %w", action.Name, err)
		}
		b := buf.Bytes()
		if action.FormatFunc != nil {
			output, err := action.FormatFunc(b)
			if err != nil && !e.AlwaysWrite {
				return fmt.Errorf("format-func is failed in action=%q: %w", action.Name, err)
			}
			if err == nil {
				b = output
			}
		}

		if err := e.saver.SaveOrCreateFile(fpath, b); err != nil {
			return fmt.Errorf("write-file is failed in action=%q: %w", action.Name, err)
		}
	}
	return nil
}

type PathResolver struct {
	*Config
	RootDirs map[string]string
}

func newPathResolver(rootdir string, config *Config) *PathResolver {
	return &PathResolver{
		RootDirs: map[string]string{"/": rootdir},
		Config:   config,
	}
}

// AddRoot adds another rootdir.
func (r *PathResolver) AddRoot(pkgpath string, rootdir string) error {
	if !strings.HasPrefix(pkgpath, "/") {
		return ErrInvalidPath
	}
	r.RootDirs[pkgpath] = filepath.Join(rootdir, "")
	return nil
}

// ResolvePath resolves file path from package path.
func (r *PathResolver) ResolvePath(pkgpath string) (string, error) {
	if pkgpath == "/" || !strings.HasPrefix(pkgpath, "/") {
		fpath := filepath.Join(r.RootDirs["/"], pkgpath)
		if r.Debug {
			r.Log.Printf("\tresolve filepath %q -> %q", pkgpath, fpath)
		}
		return fpath, nil
	}

	if fpath, ok := r.RootDirs[pkgpath]; ok {
		if r.Debug {
			r.Log.Printf("\tresolve filepath %q -> %q (cached)", pkgpath, fpath)
		}
		return fpath, nil
	}

	parts := strings.Split(pkgpath, "/") // "/foo/bar" -> ["", "foo", "bar"]
	for i := len(parts) - 1; i > 0; i-- {
		prefix := strings.Join(parts[:i], "/")
		if prefix == "" {
			prefix = "/"
		}

		parent, ok := r.RootDirs[prefix]
		if r.Debug {
			r.Log.Printf("\t\tlookup pkgpath=%s, prefix=%s -> ok=%v", pkgpath, prefix, ok)
		}
		if ok {
			fpath := filepath.Join(parent, strings.Join(parts[i:], "/"))
			r.RootDirs[pkgpath] = fpath
			if r.Debug {
				r.Log.Printf("\tresolve filepath %q -> %q (registered)", pkgpath, fpath)
			}
			return fpath, nil
		}
	}

	if r.Debug {
		saved := make([]string, 0, len(r.RootDirs))
		for k := range r.RootDirs {
			saved = append(saved, k)
		}
		r.Log.Printf("\terror, not found, input pkgpath=%s ... saved=%v", pkgpath, saved)
	}
	return "", ErrNotFound
}

var (
	ErrNotFound    = fmt.Errorf("not found")
	ErrInvalidPath = fmt.Errorf("invalid path")
)
