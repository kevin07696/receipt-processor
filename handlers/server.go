package handlers

import (
	"fmt"
	"net/http"
)

type App struct {
	router *http.ServeMux
	port   int
}

func NewApp(port int, router *http.ServeMux) *App {
	a := &App{
		router: router,
		port:   port,
	}

	return a
}

func (a *App) Run(middlewares ...Middleware) error {
	middlewareChain := MiddlewareChain(middlewares...)
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", a.port),
		Handler: middlewareChain(a.router),
	}

	return server.ListenAndServe()
}
