package receipt

import (
	"context"

	"github.com/kevin07696/receipt-processor/domain"
)

type ReceiptProcessorRepository struct {
	cache IRepository
}

func NewReceiptProcessorRepository(cache IRepository) *ReceiptProcessorRepository {
	return &ReceiptProcessorRepository{
		cache: cache,
	}
}

func (r *ReceiptProcessorRepository) WriteReceiptScore(ctx context.Context, id string, points int64) domain.StatusCode {
	return r.cache.Set(ctx, id, points)
}

func (r ReceiptProcessorRepository) ReadReceiptScore(ctx context.Context, id string) (int64, domain.StatusCode) {
	score, status := r.cache.Get(ctx, id)
	if status > 0 {
		return 0, status
	}

	return score.(int64), domain.StatusOK
}
