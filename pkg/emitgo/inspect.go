package emitgo

import (
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

func PackagePath(ob interface{}) string {
	rt := reflect.TypeOf(ob)
	if pkg := pkgPath(rt); pkg != "" {
		return pkg
	}
	rfunc := runtime.FuncForPC(reflect.ValueOf(ob).Pointer())
	parts := strings.Split(rfunc.Name(), ".") // method is not supported
	return strings.Join(parts[:len(parts)-1], ".")
}

func pkgPath(rt reflect.Type) string {
	pkg := rt.PkgPath()
	if pkg != "" {
		return pkg
	}
	switch rt.Kind() {
	case reflect.Slice, reflect.Array, reflect.Ptr, reflect.Map:
		return pkgPath(rt.Elem())
	}
	return pkg
}

func DefinedDir(fn interface{}) string {
	return filepath.Dir(DefinedFile(fn))
}
func DefinedFile(fn interface{}) string {
	rfunc := runtime.FuncForPC(reflect.ValueOf(fn).Pointer())
	fpath, _ := rfunc.FileLine(rfunc.Entry())
	return fpath
}
