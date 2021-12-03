// Code generated by "github.com/podhmo/apikit"; DO NOT EDIT.


package handler

import (
	"m/14openapi-petstore/action"
	"net/http"
	"main/handler/runtime"
)

func FindPets(getProvider func(*http.Request) (*http.Request, Provider, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		req, provider, err := getProvider(req)
		if err != nil {
			runtime.HandleResult(w, req, nil, err); return
		}
		var p *action.PetStore
		{
			p = provider.PetStore()
		}
		var tags *[]string
		{
			tags = provider.()
		}
		var queryParams struct {
			limit *int32 `query:"limit"`
		}
		if err := runtime.BindQuery(&queryParams, req); err != nil {
			_ = err // ignored
		}
		result, err := action.FindPets(p, tags, queryParams.limit)
		runtime.HandleResult(w, req, result, err)
	}
}
