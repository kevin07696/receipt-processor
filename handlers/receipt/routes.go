package receipt

func (h *Handler) initAppRoutes() {
	h.router.HandleFunc("POST /receipts/process", ProcessReceipt(h.receiptAPI))
	h.router.HandleFunc("GET /receipts/{id}/points", GetScore(h.receiptAPI))
}
