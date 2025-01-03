package receipt

import (
	"net/http"

	"github.com/kevin07696/receipt-processor/domain/receipt"
)

type Handler struct {
	router     *http.ServeMux
	receiptAPI receipt.IReceiptProcessorService
}

func Handle(router *http.ServeMux, receiptAPI receipt.IReceiptProcessorService) *Handler {
	h := &Handler{
		router:     router,
		receiptAPI: receiptAPI,
	}

	h.initAppRoutes()

	return h
}
