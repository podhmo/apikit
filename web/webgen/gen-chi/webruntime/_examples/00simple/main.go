package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"webruntime"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

func Hello(w http.ResponseWriter, r *http.Request) {
	v := map[string]string{"message": "hello"}
	webruntime.HandleResult(w, r, v, nil)
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("!! %+v", err)
	}
}

func run() error {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.Get("/", Hello)

	port := 3000
	if v, err := strconv.Atoi(os.Getenv("PORT")); err == nil {
		port = v
	}
	addr := fmt.Sprintf(":%d", port)

	log.Println("listen ...", addr)
	return http.ListenAndServe(addr, r)
}
