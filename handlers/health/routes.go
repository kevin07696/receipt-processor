package health

func (h *Handler) initAppRoutes() {
	h.router.HandleFunc("GET /health", HealthCheck())
	h.router.HandleFunc("GET /exit/{code}", Exit())
}
