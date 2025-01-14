package admin

import "net/http"

func InitializeRoutes(router *http.ServeMux) {
	router.HandleFunc("GET /health", HealthCheck())
	router.HandleFunc("GET /exit/{code}", Exit())
}
