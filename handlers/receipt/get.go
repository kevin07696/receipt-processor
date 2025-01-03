package receipt

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kevin07696/receipt-processor/domain"
	"github.com/kevin07696/receipt-processor/domain/receipt"
)

var idKey = "id"

func GetScore(receiptAPI receipt.IReceiptProcessorService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second)
		defer cancel()

		path := r.URL.Path
		path = strings.Trim(path, "/")
		segments := strings.Split(path, "/")

		// Assuming the route is always valid
		id := segments[1]
		if err := uuid.Validate(id); err != nil {
			slog.DebugContext(ctx, "StatusBadRequest: uuid is invalid", slog.String("id", id), slog.Any("error", err))
			http.Error(w, domain.ErrorToCodes[domain.ErrBadRequest].Message, domain.ErrorToCodes[domain.ErrBadRequest].Code)
			return
		}

		response, status := receiptAPI.GetReceiptScore(ctx, receipt.ReceiptScoreRequest{ID: id})
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
