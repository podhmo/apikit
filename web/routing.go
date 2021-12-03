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
	Method string
	Path   string

	DefaultStatusCode int // if 0 -> 200

	// additional
	Name              string // if not renamed, the value is  ""
	Description       string
	ExtraDependencies []interface{}
}

var (
	mu          sync.Mutex
	metadataMap = map[*Node]*MetaData{}
)

type RoutingOption func(*Node, *MetaData)

func WithExtraDependencies(deps ...interface{}) RoutingOption {
	return func(node *Node, metadata *MetaData) {
		metadata.ExtraDependencies = append(metadata.ExtraDependencies, deps...)
	}
}
func WithAnotherHandlerName(name string) RoutingOption {
	return func(node *Node, metadata *MetaData) {
		metadata.Name = name
	}
}
func WithDefaultStatusCode(code int) RoutingOption {
	return func(node *Node, metadata *MetaData) {
		metadata.DefaultStatusCode = code
	}
}
func GetMetaData(node *Node) MetaData {
	mu.Lock()
	defer mu.Unlock()
	return *metadataMap[node]
}

func (r *Router) Method(method, pattern string, fn T, options ...RoutingOption) *Node {
	k := fmt.Sprintf("%s %s%s", method, r.FullPrefix, pattern)
	node, err := r.Root.CreateNode(k, nil)
	if err != nil {
		r.ErrorHandler(err)
		return nil
	}
	node.Value = fn
	mu.Lock()
	metadata := &MetaData{
		Method: method,
		Path:   r.FullPrefix + pattern,
	}
	metadataMap[node] = metadata
	mu.Unlock()

	for _, opt := range options {
		opt(node, metadata)
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
