package emitfile

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var DEBUG = false

func init() {
	if v, err := strconv.ParseBool(os.Getenv("DEBUG")); err == nil {
		DEBUG = v
	}
}

type Emitter struct {
	PathResolver *PathResolver
}

func NewEmitter(rootdir string) *Emitter {
	return &Emitter{
		PathResolver: newPathResolver(rootdir),
	}
}

type PathResolver struct {
	RootDirs map[string]string
}

func newPathResolver(rootdir string) *PathResolver {
	return &PathResolver{
		RootDirs: map[string]string{"/": rootdir},
	}
}

func (r *PathResolver) Add(pkgpath string, rootdir string) error {
	if !strings.HasPrefix(pkgpath, "/") {
		return ErrInvalidPath
	}
	r.RootDirs[pkgpath] = filepath.Join(rootdir, "")
	return nil
}

func (r *PathResolver) Resolve(pkgpath string) (string, error) {
	if pkgpath == "/" || !strings.HasPrefix(pkgpath, "/") {
		fpath := filepath.Join(r.RootDirs["/"], pkgpath)
		if DEBUG {
			log.Printf("input pkgpath=%s -> resolve filepath=%q", pkgpath, fpath)
		}
		return fpath, nil
	}

	if fpath, ok := r.RootDirs[pkgpath]; ok {
		if DEBUG {
			log.Printf("input pkgpath=%s -> resolve filepath=%q (cached)", pkgpath, fpath)
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
		if DEBUG {
			log.Printf("\tlookup pkgpath=%s, prefix=%s -> ok=%v", pkgpath, prefix, ok)
		}
		if ok {
			fpath := filepath.Join(parent, strings.Join(parts[i:], "/"))
			r.RootDirs[pkgpath] = fpath
			if DEBUG {
				log.Printf("input pkgpath=%s -> resolve filepath=%q (saved)", pkgpath, fpath)
			}
			return fpath, nil
		}
	}
	if DEBUG {
		saved := make([]string, 0, len(r.RootDirs))
		for k := range r.RootDirs {
			saved = append(saved, k)
		}
		log.Printf("error, not found, input pkgpath=%s ... saved=%v", pkgpath, saved)
	}
	return "", ErrNotFound
}

var (
	ErrNotFound    = fmt.Errorf("not found")
	ErrInvalidPath = fmt.Errorf("invalid path")
)
