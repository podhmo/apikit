// Code generated by "github.com/podhmo/apikit"; DO NOT EDIT.

package handler

import (
	"m/14openapi-petstore/action"
	"main/handler/runtime"
	"net/http"
)

func AddPet(getProvider func(*http.Request) (*http.Request, Provider, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		req, provider, err := getProvider(req)
		if err != nil {
			runtime.HandleResult(w, req, nil, err)
			return
		}
		var p *action.PetStore
		{
			p = provider.PetStore()
		}
		var params action.NewPet
		if err := runtime.BindBody(&params, req.Body); err != nil {
			w.WriteHeader(400)
			runtime.HandleResult(w, req, nil, err)
			return
		}
		if err := runtime.ValidateStruct(&params); err != nil {
			w.WriteHeader(422)
			runtime.HandleResult(w, req, nil, err)
			return
		}
		result, err := action.AddPet(p, params)
		runtime.HandleResult(w, req, result, err)
	}
}
