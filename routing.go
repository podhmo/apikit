package apikit

// type Router interface {
// 	Method(method, pattern string, fn T)
// 	Mount(pattern string, r Router)
// 	Group(fn func(r Router)) Router
// }

func NewRouter() *Router {
	return &Router{}
}

type Router struct {
	Paths    []*PathItem
	Children []*RouterItem
	Parents  []*Router
}

type RouterItem struct {
	Name string
	*Router
}

func (r *Router) Group(pattern string, use func(*Router)) *RouterItem {
	sub := &Router{}
	item := sub.Mount(pattern, sub)
	use(sub)
	return item
}

func (r *Router) Mount(pattern string, child *Router) *RouterItem {
	item := &RouterItem{
		Name:   pattern,
		Router: child,
	}
	r.Children = append(r.Children, item)
	return item
}

func (r *Router) Method(method, pattern string, fn T) *PathItem {
	path := &Path{
		Method: method,
		Path:   pattern,
		Raw:    fn,
	}
	item := &PathItem{
		Path: path,
	}
	r.Paths = append(r.Paths, item)
	return item
}

type Path struct {
	Method string
	Path   string
	Raw    T
}

func (r *Router) Connect(pattern string, fn T) *PathItem {
	return r.Method("CONNECT", pattern, fn)
}
func (r *Router) Delete(pattern string, fn T) *PathItem {
	return r.Method("DELETE", pattern, fn)
}
func (r *Router) Get(pattern string, fn T) *PathItem {
	return r.Method("GET", pattern, fn)
}
func (r *Router) Head(pattern string, fn T) *PathItem {
	return r.Method("HEAD", pattern, fn)
}
func (r *Router) Options(pattern string, fn T) *PathItem {
	return r.Method("OPTIONS", pattern, fn)
}
func (r *Router) Patch(pattern string, fn T) *PathItem {
	return r.Method("PATCH", pattern, fn)
}
func (r *Router) Post(pattern string, fn T) *PathItem {
	return r.Method("POST", pattern, fn)
}
func (r *Router) Put(pattern string, fn T) *PathItem {
	return r.Method("PUT", pattern, fn)
}
func (r *Router) Trace(pattern string, fn T) *PathItem {
	return r.Method("TRACE", pattern, fn)
}

type PathItem struct {
	Name string
	*Path
}

type PathParam struct {
	Name    string
	Type    string
	Pattern string
}
