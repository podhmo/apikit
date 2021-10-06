package webruntime

import (
	"context"
	"io"
	"net/http"
	"reflect"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/schema"
)

var mu sync.Mutex
var decoder = schema.NewDecoder()
var validate = validator.New()

func Validate(ob interface{}) error {
	// TODO: merge error
	if err := validate.Struct(ob); err != nil {
		return err
	}
	if v, ok := ob.(interface{ Validate() error }); ok {
		return v.Validate()
	}
	return nil
}

// TODO: performance
func BindPath(dst interface{}, req *http.Request, keys []string) error {
	params := make(map[string][]string, len(keys))
	rctx := chi.RouteContext(req.Context())
	if rctx == nil {
		return nil
	}

	for _, k := range keys {
		params[k] = []string{rctx.URLParam(k)}
	}
	mu.Lock()
	defer mu.Unlock()
	decoder.SetAliasTag("path")
	return decoder.Decode(dst, params)
}

func BindQuery(dst interface{}, req *http.Request) error {
	mu.Lock()
	defer mu.Unlock()
	decoder.SetAliasTag("query")
	return decoder.Decode(dst, req.URL.Query())
}

func BindHeader(dst interface{}, req *http.Request, keys []string) error {
	mu.Lock()
	defer mu.Unlock()
	decoder.SetAliasTag("header")
	return decoder.Decode(dst, req.Header)
}

func BindBody(dst interface{}, r io.ReadCloser) error {
	if err := render.DecodeJSON(r, dst); err != nil {
		return err
	}
	return nil
}

func HandleResult(w http.ResponseWriter, req *http.Request, v interface{}, err error) {
	// Force to return empty JSON array [] instead of null in case of zero slice.
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Slice && val.IsNil() {
		v = reflect.MakeSlice(val.Type(), 0, 0).Interface()
	}

	var code int
	if err != nil {
		code = 500
		if impl, ok := err.(interface{ Code() int }); ok {
			code = impl.Code()
		}
		v = map[string]interface{}{"message": err.Error()}
	}
	if code != 0 {
		req = req.WithContext(context.WithValue(req.Context(), render.StatusCtxKey, code))
	}
	render.JSON(w, req, v)
}
