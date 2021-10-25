package webruntime

import "net/http"

func init() {
	HandleResult = CreateHandleResultFunction(func(err error) int {
		if err == nil {
			return http.StatusOK
		}
		return http.StatusInternalServerError
	})
}
