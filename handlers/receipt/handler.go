package receipt

import (
	"net/http"

	"github.com/kevin07696/receipt-processor/domain/receipt"
	"github.com/kevin07696/receipt-processor/handlers"
)

type Handler struct {
	router     *http.ServeMux
	receiptAPI receipt.IReceiptProcessorService
}

func Handle(router *http.ServeMux, receiptAPI receipt.IReceiptProcessorService, middlewares ...handlers.Middleware) *Handler {
	h := &Handler{
		router:     router,
		receiptAPI: receiptAPI,
	}

	h.initAppRoutes()

	middlewareChain := handlers.MiddlewareChain(middlewares...)
	middlewareChain(h.router)

	return h
}
