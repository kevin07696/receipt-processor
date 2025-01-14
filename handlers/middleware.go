package handlers

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/kevin07696/receipt-processor/infrastructure/loggers"
)

type Middleware func(http.Handler) http.HandlerFunc

func ChainMiddlewaresToHandler(handler http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

func RequestLoggerMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.InfoContext(r.Context(), fmt.Sprintf("Method %s, Path: %s", r.Method, r.URL.Path))
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
