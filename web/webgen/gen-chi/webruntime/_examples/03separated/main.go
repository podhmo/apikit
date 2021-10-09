package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"m/config/dbconfig"
	"m/handler"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Config struct {
	dbconfig.DBConfig
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("!! %+v", err)
	}
}

func run() error {
	r := chi.NewRouter()

	// TODO: use httplog
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(middleware.Heartbeat("/_ping"))

	config := &Config{}

	getProvider := func(req *http.Request) (*http.Request, handler.Provider, error) {
		return req, config, nil
	}
	handler.Mount(r, getProvider)

	port := 8888
	if v, err := strconv.Atoi(os.Getenv("PORT")); err == nil {
		port = v
	}
	addr := fmt.Sprintf(":%d", port)
	log.Println("listen ...", addr)
	return http.ListenAndServe(addr, r)
}
