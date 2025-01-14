package receipt

import (
	"net/http"

	"github.com/kevin07696/receipt-processor/domain/receipt"
)

func InitializeRoutes(router *http.ServeMux, receiptAPI receipt.IReceiptProcessorService) {
	router.HandleFunc("POST /receipts/process", ProcessReceipt(receiptAPI))
	router.HandleFunc("GET /receipts/{id}/points", GetScore(receiptAPI))
}
