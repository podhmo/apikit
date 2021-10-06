package webruntime

import (
	"bytes"
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/go-chi/chi/v5"
)

var PathParam = chi.URLParam

func Wrap(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		next.ServeHTTP(w, req)
	}
}

func HandleResult(w http.ResponseWriter, req *http.Request, v interface{}, err error) {
	// Force to return empty JSON array [] instead of null in case of zero slice.
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Slice && val.IsNil() {
		v = reflect.MakeSlice(val.Type(), 0, 0).Interface()
	}

	JSON(w, req, v, err)
}

func JSON(w http.ResponseWriter, req *http.Request, v interface{}, err error) {
	var code int
	target := v

	if err != nil {
		code = 500
		if impl, ok := err.(interface{ Code() int }); ok {
			code = impl.Code()
		}
		target = map[string]interface{}{"message": err.Error()}
	}

	buf := &bytes.Buffer{}
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(true)

	if err := encoder.Encode(target); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if code > 0 {
		w.WriteHeader(code)
	}
	w.Write(buf.Bytes())
}
