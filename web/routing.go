package web

import (
	"fmt"
	"log"

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

func (r *Router) Method(method, pattern string, fn T) *Node {
	path := fmt.Sprintf("%s %s%s", method, r.FullPrefix, pattern)
	node, err := r.Root.CreateNode(path, nil)
	if err != nil {
		r.ErrorHandler(err)
		return nil
	}
	node.Value = fn
	return node
}

func (r *Router) Group(pattern string, use func(*Router)) *Router {
	fullprefix := r.FullPrefix + pattern
	child := &Router{Root: r.Root, ErrorHandler: r.ErrorHandler, Parent: r, prefix: pattern, FullPrefix: fullprefix}
	use(child)
	return child
}

func (r *Router) Connect(pattern string, fn T) *Node {
	return r.Method("CONNECT", pattern, fn)
}
func (r *Router) Delete(pattern string, fn T) *Node {
	return r.Method("DELETE", pattern, fn)
}
func (r *Router) Get(pattern string, fn T) *Node {
	return r.Method("GET", pattern, fn)
}
func (r *Router) Head(pattern string, fn T) *Node {
	return r.Method("HEAD", pattern, fn)
}
func (r *Router) Options(pattern string, fn T) *Node {
	return r.Method("OPTIONS", pattern, fn)
}
func (r *Router) Patch(pattern string, fn T) *Node {
	return r.Method("PATCH", pattern, fn)
}
func (r *Router) Post(pattern string, fn T) *Node {
	return r.Method("POST", pattern, fn)
}
func (r *Router) Put(pattern string, fn T) *Node {
	return r.Method("PUT", pattern, fn)
}
func (r *Router) Trace(pattern string, fn T) *Node {
	return r.Method("TRACE", pattern, fn)
}
