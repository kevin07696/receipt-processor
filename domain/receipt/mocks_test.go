package receipt_test

import (
	"context"

	"github.com/kevin07696/receipt-processor/domain"
)

type MockReceiptRepository struct {
	WriteReceiptScoreMock func(ctx context.Context, id string, points uint16, scores map[string]uint16) domain.StatusCode
	ReadReceiptScoreMock  func(ctx context.Context, id string, scores map[string]uint16) (uint16, domain.StatusCode)
	Scores                map[string]uint16
}

func (m MockReceiptRepository) WriteReceiptScore(ctx context.Context, id string, points uint16) domain.StatusCode {
	return m.WriteReceiptScoreMock(ctx, id, points, m.Scores)
}

func (m MockReceiptRepository) ReadReceiptScore(ctx context.Context, id string) (uint16, domain.StatusCode) {
	return m.ReadReceiptScoreMock(ctx, id, m.Scores)
}
