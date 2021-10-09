package webruntime

import (
	"io"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/gorilla/schema"
	"github.com/morikuni/failure"
)

var mu sync.Mutex
var decoder = schema.NewDecoder()

// TODO: performance
func BindPathParams(dst interface{}, req *http.Request, keys ...string) error {
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
	target := v

	if err != nil {
		// log if 5xx ? (or middleware?)
		target = &errorRender{
			HTTPStatusCode: statusOf(err),
			Message:        messageOf(err),
			DebugContext:   debugContextOf(err),
		}
	} else {
		// Force to return empty JSON array [] instead of null in case of zero slice.
		val := reflect.ValueOf(v)
		if val.Kind() == reflect.Slice && val.IsNil() {
			target = reflect.MakeSlice(val.Type(), 0, 0).Interface()
		}
	}
	render.JSON(w, req, target)
}

// error

type errorRender struct {
	HTTPStatusCode int    `json:"-"`
	Message        string `json:"message"`
	DebugContext   string `json:"debug-context,omitempty"`
}

func (e *errorRender) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

// API Error
// error codes for your application.
const (
	NotFound  failure.StringCode = "NotFound"
	Forbidden failure.StringCode = "Forbidden"
)

// TODO: inject from external

func statusOf(err error) int {
	c, ok := failure.CodeOf(err)
	if !ok {
		return http.StatusInternalServerError
	}
	switch c {
	case NotFound:
		return http.StatusNotFound
	case Forbidden:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

func messageOf(err error) string {
	msg, ok := failure.MessageOf(err)
	if ok {
		return msg
	}
	return "Error"
}

func debugContextOf(err error) string {
	msg, ok := failure.MessageOf(err)
	if ok {
		return msg
	}
	return "Error"
}

var DEBUG = false

func init() {
	if v, err := strconv.ParseBool(os.Getenv("DEBUG")); err == nil {
		DEBUG = v
	}
}
