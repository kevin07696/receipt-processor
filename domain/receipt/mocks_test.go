package receipt_test

import (
	"context"

	"github.com/kevin07696/receipt-processor/domain"
)

type MockReceiptRepository struct {
	WriteReceiptScoreMock func(ctx context.Context, id string, points int64, scores map[string]int64) domain.StatusCode
	ReadReceiptScoreMock  func(ctx context.Context, id string, scores map[string]int64) (int64, domain.StatusCode)
	Scores                map[string]int64
}

func (m MockReceiptRepository) WriteReceiptScore(ctx context.Context, id string, points int64) domain.StatusCode {
	return m.WriteReceiptScoreMock(ctx, id, points, m.Scores)
}

func (m MockReceiptRepository) ReadReceiptScore(ctx context.Context, id string) (int64, domain.StatusCode) {
	return m.ReadReceiptScoreMock(ctx, id, m.Scores)
}
