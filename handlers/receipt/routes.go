package receipt

import "fmt"

func (h *Handler) initAppRoutes() {
	h.router.HandleFunc("GET /health", HealthCheck())
	h.router.HandleFunc("POST /receipts/process", ProcessReceipt(h.receiptAPI))
	h.router.HandleFunc(fmt.Sprintf("GET /receipts/{%s}/points", idKey), GetScore(h.receiptAPI))
}
