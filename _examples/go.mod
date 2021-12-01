module m

go 1.16

replace github.com/podhmo/apikit => ../

require (
	github.com/getkin/kin-openapi v0.83.0
	github.com/go-chi/chi/v5 v5.0.4
	github.com/go-chi/render v1.0.1
	github.com/go-playground/locales v0.14.0
	github.com/go-playground/universal-translator v0.18.0
	github.com/go-playground/validator/v10 v10.9.0
	github.com/gorilla/schema v1.2.0
	github.com/morikuni/failure v0.14.0
	github.com/podhmo/apikit v0.0.0-20210923132720-b5e305e3689c
	github.com/podhmo/reflect-openapi v0.0.13
	github.com/podhmo/reflect-shape v0.3.5
)
