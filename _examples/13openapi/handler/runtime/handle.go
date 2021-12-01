// Code generated by "github.com/podhmo/apikit"; DO NOT EDIT.

package runtime

import (
	"m/13openapi/design"
)

import (
	"errors"
	"net/http"
	"reflect"

	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

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
func init() {
	HandleResult = CreateHandleResultFunction(design.HTTPStatusOf)
}
