module m

go 1.16

replace github.com/podhmo/apikit => ../

require (
	github.com/go-chi/chi/v5 v5.0.4
	github.com/go-chi/render v1.0.1
	github.com/gorilla/schema v1.2.0
	github.com/morikuni/failure v0.14.0
	github.com/podhmo/apikit v0.0.0-20210923132720-b5e305e3689c
	golang.org/x/sys v0.0.0-20210806184541-e5e7981a1069 // indirect
)
