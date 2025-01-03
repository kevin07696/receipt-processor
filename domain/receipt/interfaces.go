package receipt

import (
	"context"

	"github.com/kevin07696/receipt-processor/domain"
)

type IReceiptProcessorService interface {
	ProcessReceipt(ctx context.Context, request ReceiptProcessorRequest) (ReceiptProcessorResponse, domain.StatusCode)
	GetReceiptScore(ctx context.Context, request ReceiptScoreRequest) (ReceiptScoreResponse, domain.StatusCode)
	GenerateID(ctx context.Context, input string) string
}

type IReceiptProcessorRepository interface {
	WriteReceiptScore(ctx context.Context, id string, value int64) domain.StatusCode
	ReadReceiptScore(ctx context.Context, id string) (int64, domain.StatusCode)
}

type IRepository interface {
	Set(ctx context.Context, id string, value interface{}) domain.StatusCode
	Get(ctx context.Context, id string) (interface{}, domain.StatusCode)
}
