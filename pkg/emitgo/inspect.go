package emitgo

import (
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

func PackagePath(ob interface{}) string {
	rt := reflect.TypeOf(ob)
	if pkgPath := rt.PkgPath(); pkgPath != "" {
		return pkgPath
	}
	rfunc := runtime.FuncForPC(reflect.ValueOf(ob).Pointer())
	parts := strings.Split(rfunc.Name(), ".") // method is not supported
	return strings.Join(parts[:len(parts)-1], ".")
}

func DefinedDir(fn interface{}) string {
	return filepath.Dir(DefinedFile(fn))
}
func DefinedFile(fn interface{}) string {
	rfunc := runtime.FuncForPC(reflect.ValueOf(fn).Pointer())
	fpath, _ := rfunc.FileLine(rfunc.Entry())
	return fpath
}
