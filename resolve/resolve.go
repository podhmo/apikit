package resolve

import (
	"reflect"
	"sync"

	"github.com/podhmo/apikit/pkg/tinypkg"
	reflectshape "github.com/podhmo/reflect-shape"
	"github.com/podhmo/reflect-shape/arglist"
)

type Resolver struct {
	extractor *reflectshape.Extractor
	universe  *tinypkg.Universe

	mu           sync.RWMutex
	symbolsCache map[reflectshape.Identity][]*symbolCacheItem
	defCache     map[uintptr]*Def
}

func NewResolver() *Resolver {
	e := &reflectshape.Extractor{
		Seen:           map[reflect.Type]reflectshape.Shape{},
		ArglistLookup:  arglist.NewLookup(),
		RevisitArglist: true,
	}
	return &Resolver{
		extractor:    e,
		universe:     tinypkg.NewUniverse(),
		symbolsCache: map[reflectshape.Identity][]*symbolCacheItem{},
		defCache:     map[uintptr]*Def{},
	}
}

func (r *Resolver) NewPackage(path, name string) *tinypkg.Package {
	return r.universe.NewPackage(path, name)
}

func (r *Resolver) Def(fn interface{}) *Def {
	r.mu.RLock()
	k := reflect.ValueOf(fn).Pointer()

	cached, ok := r.defCache[k]
	r.mu.RUnlock()
	if ok {
		return cached
	}

	def := ExtractDef(r.universe, r.extractor, fn)
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
	sym := ExtractSymbol(r, here, s)

	r.mu.Lock()
	r.symbolsCache[k] = append(cached, &symbolCacheItem{
		Here:  here,
		Shape: s,
		Node:  sym,
	})
	r.mu.Unlock()

	return sym
}
