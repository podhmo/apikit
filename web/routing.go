package web

import (
	"fmt"
	"log"
	"sync"

	"github.com/podhmo/apikit/web/pathpattern"
)

type T = interface{}
type Node = pathpattern.Node
type Router struct {
	Root         *Node
	ErrorHandler func(error)

	Parent *Router // []*Router?

	FullPrefix string
	prefix     string
}

func NewRouter() *Router {
	return &Router{
		Root: &Node{},
		ErrorHandler: func(err error) {
			log.Printf("ERROR: %+v", err)
		},
	}
}

type MetaData struct {
	Name              string // if not renamed, the value is  ""
	ExtraDependencies []interface{}
}

var (
	mu          sync.Mutex
	metadataMap = map[*Node]*MetaData{}
)

type RoutingOption func(*Node)

func WithExtraDependencies(deps ...interface{}) RoutingOption {
	return func(node *Node) {
		SetExtraDependencies(node, deps)
	}
}
func WithRename(name string) RoutingOption {
	return func(node *Node) {
		mu.Lock()
		defer mu.Unlock()
		metadataMap[node].Name = name
	}
}
func GetName(node *Node) string {
	mu.Lock()
	defer mu.Unlock()
	return metadataMap[node].Name
}

func SetExtraDependencies(node *Node, deps []interface{}) {
	mu.Lock()
	defer mu.Unlock()
	m := metadataMap[node]
	m.ExtraDependencies = append(m.ExtraDependencies, deps...)
}
func GetExtraDependencies(node *Node) []interface{} {
	mu.Lock()
	defer mu.Unlock()
	return metadataMap[node].ExtraDependencies
}

func (r *Router) Method(method, pattern string, fn T, options ...RoutingOption) *Node {
	path := fmt.Sprintf("%s %s%s", method, r.FullPrefix, pattern)
	node, err := r.Root.CreateNode(path, nil)
	if err != nil {
		r.ErrorHandler(err)
		return nil
	}
	node.Value = fn
	mu.Lock()
	metadataMap[node] = &MetaData{}
	mu.Unlock()

	for _, opt := range options {
		opt(node)
	}
	return node
}

func (r *Router) Group(pattern string, use func(*Router)) *Router {
	fullprefix := r.FullPrefix + pattern
	child := &Router{Root: r.Root, ErrorHandler: r.ErrorHandler, Parent: r, prefix: pattern, FullPrefix: fullprefix}
	use(child)
	return child
}

func (r *Router) Connect(pattern string, fn T, options ...RoutingOption) *Node {
	return r.Method("CONNECT", pattern, fn, options...)
}
func (r *Router) Delete(pattern string, fn T, options ...RoutingOption) *Node {
	return r.Method("DELETE", pattern, fn, options...)
}
func (r *Router) Get(pattern string, fn T, options ...RoutingOption) *Node {
	return r.Method("GET", pattern, fn, options...)
}
func (r *Router) Head(pattern string, fn T, options ...RoutingOption) *Node {
	return r.Method("HEAD", pattern, fn, options...)
}
func (r *Router) Options(pattern string, fn T, options ...RoutingOption) *Node {
	return r.Method("OPTIONS", pattern, fn, options...)
}
func (r *Router) Patch(pattern string, fn T, options ...RoutingOption) *Node {
	return r.Method("PATCH", pattern, fn, options...)
}
func (r *Router) Post(pattern string, fn T, options ...RoutingOption) *Node {
	return r.Method("POST", pattern, fn, options...)
}
func (r *Router) Put(pattern string, fn T, options ...RoutingOption) *Node {
	return r.Method("PUT", pattern, fn, options...)
}
func (r *Router) Trace(pattern string, fn T, options ...RoutingOption) *Node {
	return r.Method("TRACE", pattern, fn, options...)
}
