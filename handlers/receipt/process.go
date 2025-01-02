package receipt

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/kevin07696/receipt-processor/domain"
	"github.com/kevin07696/receipt-processor/domain/receipt"
)

func ProcessReceipt(receiptAPI receipt.IReceiptProcessorService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second)
		defer cancel()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Fatalf("Failed to read request body: %v", err)
		}

		var input receipt.Receipt
		if err := json.Unmarshal(body, &input); err != nil {
			http.Error(w, domain.ErrorToCodes[domain.ErrBadRequest].Message, domain.ErrorToCodes[domain.ErrBadRequest].Code)
			return
		}

		isValid := input.Validate()
		if !isValid {
			log.Printf("Validation Error: %v", err)
			http.Error(w, domain.ErrorToCodes[domain.ErrBadRequest].Message, domain.ErrorToCodes[domain.ErrBadRequest].Code)
			return

		}

		id := receiptAPI.GenerateID(ctx, "")

		response, status := receiptAPI.ProcessReceipt(ctx, receipt.ReceiptProcessorRequest{ID: id, Receipt: input})
		if status > 0 {
			http.Error(w, domain.ErrorToCodes[status].Message, domain.ErrorToCodes[status].Code)
			return
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			log.Fatalf("Failed to marshal response: %v", err)
		}

		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	}
}
