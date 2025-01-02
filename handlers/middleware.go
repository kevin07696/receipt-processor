package handlers

import (
	"log"
	"net/http"
)

type Middleware func(http.Handler) http.HandlerFunc

func MiddlewareChain(middleware ...Middleware) Middleware {
	return func(next http.Handler) http.HandlerFunc {
		for _, m := range middleware {
			next = m(next)
		}

		return next.ServeHTTP
	}
}

func RequestLoggerMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("method %s. path: %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	}
}
