package health

import "net/http"

type Handler struct {
	router *http.ServeMux
}

func Handle(router *http.ServeMux) *Handler {
	h := &Handler{
		router: router,
	}

	h.initAppRoutes()

	return h
}
