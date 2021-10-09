module m

go 1.16

replace webruntime => ../..

require (
	github.com/go-chi/chi/v5 v5.0.4
	github.com/go-playground/validator/v10 v10.9.0
	webruntime v0.0.0-00010101000000-000000000000
)
