package emitfile

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/podhmo/apikit/pkg/emitfile/classify"
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

	RootDir  string // root directory for file-generation
	CurDir   string // use for relative-path calculation in logging
	HistDir  string // the directory for saving history
	HistFile string

	Log Logger
}

func NewConfig(rootdir string) *Config {
	c := &Config{
		Debug:       DEBUG,
		Verbose:     VERBOSE,
		Log:         log.New(os.Stderr, "", 0),
		AlwaysWrite: true,
		RootDir:     rootdir,
	}
	if c.Debug {
		c.Verbose = true
		c.AlwaysWrite = true
	}

	if c.HistDir == "" {
		c.HistDir = c.RootDir
	}
	if c.HistFile == "" {
		c.HistFile = ".emitfile.json"
	}

	if cwd, err := os.Getwd(); err == nil {
		c.CurDir = cwd
	}
	return c
}

func (c *Config) NewEmitter() *Executor {
	r := newPathResolver(c)
	r.Config = c
	return &Executor{
		PathResolver: r,
		saver:        newfileSaver(c),
		Config:       c,
	}
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
	// TODO: dry-run option
	// TODO: keep-going option

	e.Log.Printf("emit files ...")
	sort.SliceStable(e.Actions, func(i, j int) bool { return e.Actions[i].Priority < e.Actions[j].Priority })

	hash := sha1.New()
	store := classify.JSONFileStore{Mtime: time.Now()}

	histfilePath := filepath.Join(e.HistDir, e.HistFile)
	prevEntries, err := store.ReadFile(histfilePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading history %q is failed: %w", histfilePath, err)
	}
	entries := make([]classify.Entry, 0, len(e.Actions))
	saveFuncMap := map[string]func() error{}

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

		entries = append(entries, classify.NewEntry(fpath, func() ([]byte, error) {
			hash.Reset()
			if _, err := hash.Write(b); err != nil {
				return nil, fmt.Errorf("write %s: %w", fpath, err)
			}
			return hash.Sum(nil), nil
		}))

		if _, alreadyExisted := saveFuncMap[fpath]; alreadyExisted {
			e.Log.Printf("WARNING: %s is conflicted", fpath)
		}
		saveFuncMap[fpath] = func() error {
			if err := e.saver.SaveOrCreateFile(fpath, b); err != nil {
				return fmt.Errorf("write-file is failed in action=%q: %w", action.Name, err)
			}
			return nil
		}
	}

	classified, err := classify.Classify(prevEntries, entries)
	if err != nil {
		return fmt.Errorf("classify, something wrong: %w", err)
	}

	for _, r := range classified {
		if e.Verbose {
			e.Log.Printf("\t%s %s", r.Type, r.Name())
		}
		switch r.Type {
		case classify.ResultTypeCreate, classify.ResultTypeUpdate:
			if err := saveFuncMap[r.Name()](); err != nil {
				return err
			}
		case classify.ResultTypeDelete:
			if err := os.Remove(r.Name()); err != nil {
				e.Log.Printf("WARNING: remove %q is failed", r.Name())
			}
		case classify.ResultTypeNotChanged:
			// noop
		default:
			if !e.AlwaysWrite {
				return fmt.Errorf("unexpected result type %v, file=%q", r.Type, r.Name())
			}
			e.Log.Printf("unexpected result type %v, file=%q", r.Type, r.Name())
		}
	}

	if err := os.MkdirAll(filepath.Dir(histfilePath), 0744); err != nil {
		return fmt.Errorf("prepare history directory: %w", err)
	}
	if err := store.WriteFile(histfilePath, classified); err != nil {
		return fmt.Errorf("writing history %q is failed: %w", histfilePath, err)
	}
	return nil
}

type PathResolver struct {
	*Config
	RootDirs map[string]string
}

func newPathResolver(config *Config) *PathResolver {
	return &PathResolver{
		RootDirs: map[string]string{"/": config.RootDir},
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
