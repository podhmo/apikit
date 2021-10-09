package webruntime

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/gorilla/schema"
	"github.com/morikuni/failure"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

var (
	// for parameters binding
	mu      sync.Mutex
	decoder = schema.NewDecoder()

	// for validation
	uni        *ut.UniversalTranslator
	translator ut.Translator
	validate   *validator.Validate
)

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

func ValidateStruct(ob interface{}) error {
	// TODO: wrap
	if err := validate.Struct(ob); err != nil {
		return err
	}
	if v, ok := ob.(interface{ Validate() error }); ok {
		return v.Validate() // TODO: 422
	}
	return nil
}

type HandleResultFunc func(w http.ResponseWriter, req *http.Request, v interface{}, err error)

var HandleResult HandleResultFunc

// CreateHandleResultFunction create HandleResult
func CreateHandleResultFunction(getHTTPStatus func(error) int) HandleResultFunc {
	return func(w http.ResponseWriter, req *http.Request, v interface{}, err error) {
		target := v

		if err != nil {
			var validationErr validator.ValidationErrors
			if errors.As(err, &validationErr) {
				r := make([]fieldError, len(err.(validator.ValidationErrors)))
				for i, fe := range err.(validator.ValidationErrors) {
					r[i] = fieldError{Field: fe.Field(), Path: fe.StructNamespace(), Message: fe.Translate(translator)}
				}
				target = &errorRender{
					HTTPStatusCode: http.StatusUnprocessableEntity,
					Error:          []fieldError{{Message: messageOf(err)}},
					DebugContext:   debugContextOf(err),
				}
			} else {
				// log if 5xx ? (or middleware?)
				target = &errorRender{
					HTTPStatusCode: getHTTPStatus(err),
					Error:          []fieldError{{Message: messageOf(err)}},
					DebugContext:   debugContextOf(err),
				}
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
}

// error

type errorRender struct {
	HTTPStatusCode int          `json:"-"`
	Error          []fieldError `json:"error"`
	DebugContext   string       `json:"debug-context,omitempty"`
}

func (e *errorRender) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

type fieldError struct {
	Field   string `json:"field"`
	Path    string `json:"path"`
	Message string `json:"message"`
}

func messageOf(err error) string {
	msg, ok := failure.MessageOf(err)
	if ok {
		return msg
	}
	return "error"
}

func debugContextOf(err error) string {
	if DEBUG {
		return fmt.Sprintf("%+v", err)
	}
	return ""
}

var DEBUG = false

func init() {
	if v, err := strconv.ParseBool(os.Getenv("DEBUG")); err == nil {
		DEBUG = v
	}

	// todo: fix
	en := en.New()
	uni := ut.New(en, en)

	// this is usually know or extracted from http 'Accept-Language' header
	// also see uni.FindTranslator(...)
	var found bool
	translator, found = uni.GetTranslator("en")
	if !found {
		panic("translator is not found")
	}

	validate = validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}
