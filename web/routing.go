package web

import (
	"fmt"
	"log"
	"strings"

	"github.com/podhmo/apikit/web/pathpattern"
)

type T = interface{}
type Node = pathpattern.Node
type Router struct {
	Root         *Node
	ErrorHandler func(error)

	Parent *Router // []*Router?
	Prefix string
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
	var prefix string
	if r.Parent != nil || r.Prefix != "" {
		this := r
		var s []string
		for {
			if this.Prefix != "" {
				s = append(s, this.Prefix)
			}
			if this.Parent == nil {
				break
			}
			this = this.Parent
		}
		for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
			s[i], s[j] = s[j], s[i]
		}
		prefix = strings.Join(s, "")
	}

	path := fmt.Sprintf("%s %s%s", method, prefix, pattern)
	node, err := r.Root.CreateNode(path, nil)
	if err != nil {
		r.ErrorHandler(err)
		return nil
	}
	node.Value = fn
	return node
}

func (r *Router) Group(pattern string, use func(*Router)) *Router {
	child := &Router{Root: r.Root, ErrorHandler: r.ErrorHandler, Parent: r, Prefix: pattern}
	use(child)
	return child
}
