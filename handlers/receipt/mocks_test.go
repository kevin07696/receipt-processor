package receipt_test

import (
	"context"

	"github.com/kevin07696/receipt-processor/domain"
	"github.com/kevin07696/receipt-processor/domain/receipt"
)

type MockReceipt struct {
	Retailer     string     `json:"retailer" validate:"required,min=1"`
	PurchaseDate string     `json:"purchaseDate" validate:"required,len=10"`
	PurchaseTime string     `json:"purchaseTime" validate:"required,len=5"`
	Items        []MockItem `json:"items" validate:"required,dive,required"`
	Total        string     `json:"total" validate:"required"`
}

type MockItem struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}

type MockReceiptService struct {
	ProcessReceiptMock  func(ctx context.Context, request receipt.ReceiptProcessorRequest) (receipt.ReceiptProcessorResponse, domain.StatusCode)
	GetReceiptScoreMock func(ctx context.Context, request receipt.ReceiptScoreRequest) (receipt.ReceiptScoreResponse, domain.StatusCode)
	GenerateIDMock      func(ctx context.Context, input string) string
}

func (m *MockReceiptService) ProcessReceipt(ctx context.Context, request receipt.ReceiptProcessorRequest) (receipt.ReceiptProcessorResponse, domain.StatusCode) {
	return m.ProcessReceiptMock(ctx, request)
}
func (m MockReceiptService) GetReceiptScore(ctx context.Context, request receipt.ReceiptScoreRequest) (receipt.ReceiptScoreResponse, domain.StatusCode) {
	return m.GetReceiptScoreMock(ctx, request)
}
func (m MockReceiptService) GenerateID(ctx context.Context, input string) string {
	return m.GenerateIDMock(ctx, input)
}
