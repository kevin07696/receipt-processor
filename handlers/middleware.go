package handlers

import (
	"log"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/kevin07696/receipt-processor/adapters/loggers"
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
		log.Printf("Method %s, Path: %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	}
}

var RequestID = "RequestID"

func RequestIDMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := loggers.AppendCtx(r.Context(), slog.String(RequestID, uuid.NewString()))
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
