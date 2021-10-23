package resolve

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync"

	"github.com/podhmo/apikit/pkg/tinypkg"
	reflectshape "github.com/podhmo/reflect-shape"
	"github.com/podhmo/reflect-shape/arglist"
)

type Resolver struct {
	Config *Config

	extractor *reflectshape.Extractor
	universe  *tinypkg.Universe

	mu             sync.RWMutex
	symbolsCache   map[reflectshape.Identity][]*symbolCacheItem
	defCache       map[uintptr]*Def
	preModuleCache map[reflectshape.Identity]*PreModule
}

func NewResolver() *Resolver {
	e := &reflectshape.Extractor{
		Seen:           map[reflect.Type]reflectshape.Shape{},
		ArglistLookup:  arglist.NewLookup(),
		RevisitArglist: true,
	}
	return &Resolver{
		Config:         DefaultConfig(),
		extractor:      e,
		universe:       tinypkg.NewUniverse(),
		symbolsCache:   map[reflectshape.Identity][]*symbolCacheItem{},
		defCache:       map[uintptr]*Def{},
		preModuleCache: map[reflectshape.Identity]*PreModule{},
	}
}

func (r *Resolver) NewPackage(path, name string) *tinypkg.Package {
	return r.universe.NewPackage(path, name)
}

func (r *Resolver) DetectKind(s reflectshape.Shape) Kind {
	return DetectKind(s, r.Config.IgnoreMap)
}

func (r *Resolver) NewPackageFromInterface(ob interface{}, name string) *tinypkg.Package {
	rv := reflect.TypeOf(ob)
	for {
		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}
		break
	}
	path := rv.PkgPath()
	if path != "" {
		return r.universe.NewPackage(path, name)
	}

	// maybe function?
	rfunc := runtime.FuncForPC(reflect.ValueOf(ob).Pointer())
	parts := strings.Split(rfunc.Name(), ".") // method is not supported
	path = strings.Join(parts[:len(parts)-1], ".")
	return r.universe.NewPackage(path, name)
}

func (r *Resolver) PreModule(ob interface{}) (*PreModule, error) {
	shape := r.extractor.Extract(ob)

	r.mu.RLock()
	k := shape.GetIdentity()

	cached, ok := r.preModuleCache[k]
	r.mu.RUnlock()
	if ok {
		return cached, nil
	}

	s, ok := shape.(reflectshape.Struct)
	if !ok {
		return nil, fmt.Errorf("must be struct (%s): %w", shape, ErrInvalidType)
	}
	m, err := NewPreModule(r, s)
	if err != nil {
		return nil, err
	}
	r.mu.Lock()
	r.preModuleCache[k] = m
	r.mu.Unlock()

	return m, nil
}

func (r *Resolver) Shape(ob interface{}) reflectshape.Shape {
	return r.extractor.Extract(ob)
}

func (r *Resolver) Def(fn interface{}) *Def {
	r.mu.RLock()
	k := reflect.ValueOf(fn).Pointer()

	cached, ok := r.defCache[k]
	r.mu.RUnlock()
	if ok {
		return cached
	}

	def := ExtractDef(r.universe, r.extractor, fn, r.Config.IgnoreMap)
	r.mu.Lock()
	r.defCache[k] = def
	r.mu.Unlock()

	return def
}

type symbolCacheItem struct {
	Here  *tinypkg.Package
	Shape reflectshape.Shape
	Node  tinypkg.Node
}

func (r *Resolver) Symbol(here *tinypkg.Package, s reflectshape.Shape) tinypkg.Node {
	r.mu.RLock()
	k := s.GetIdentity()
	cached, ok := r.symbolsCache[k]
	r.mu.RUnlock()

	if ok {
		for _, item := range cached {
			if item.Here == here {
				return item.Node
			}
		}
	}
	sym := ExtractSymbol(r.universe, here, s)

	r.mu.Lock()
	r.symbolsCache[k] = append(cached, &symbolCacheItem{
		Here:  here,
		Shape: s,
		Node:  sym,
	})
	r.mu.Unlock()

	return sym
}
