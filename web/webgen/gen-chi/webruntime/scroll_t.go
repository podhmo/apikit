package webruntime

import "reflect"

// scroll (pagination)
type ScrollT = int // generics?

func coerceScrollT(v reflect.Value) ScrollT {
	return ScrollT(v.Int())
}
