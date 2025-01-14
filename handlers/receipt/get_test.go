package receipt_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kevin07696/receipt-processor/domain"
	receiptDomain "github.com/kevin07696/receipt-processor/domain/receipt"
	receioptHandler "github.com/kevin07696/receipt-processor/handlers/receipt"
	"github.com/stretchr/testify/assert"
)

func TestGetScores(t *testing.T) {
	testCases := []struct {
		title            string
		id               string
		expectedResponse receiptDomain.ReceiptScoreResponse
		expectedCode     int
		receiptAPI       receiptDomain.IReceiptProcessorService
	}{
		{
			title:            "GivenAValidID_ReturnScore",
			id:               "af523d7a-e8d0-4af0-8bbd-d2340a4da5a4",
			expectedResponse: receiptDomain.ReceiptScoreResponse{Points: 65535},
			expectedCode:     http.StatusOK,
			receiptAPI: &MockReceiptService{
				GetReceiptScoreMock: func(ctx context.Context, request receiptDomain.ReceiptScoreRequest) (receiptDomain.ReceiptScoreResponse, domain.StatusCode) {
					return receiptDomain.ReceiptScoreResponse{Points: 65535}, domain.StatusOK
				},
			},
		},
		{
			title:        "GivenAInvalidID_ReturnBadRequestError",
			id:           "af523d7a",
			expectedCode: http.StatusBadRequest,
		},
		{
			title:        "GivenAValidID_ReturnNotFound",
			id:           "af523d7a-e8d0-4af0-8bbd-d2340a4da5a4",
			expectedCode: http.StatusNotFound,
			receiptAPI: &MockReceiptService{
				GetReceiptScoreMock: func(ctx context.Context, request receiptDomain.ReceiptScoreRequest) (receiptDomain.ReceiptScoreResponse, domain.StatusCode) {
					return receiptDomain.ReceiptScoreResponse{}, domain.ErrNotFound
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			handler := receioptHandler.GetScore(tc.receiptAPI)
			url := fmt.Sprintf("/receipts/%s/points", tc.id)

			request, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				t.Fatalf("Failed to build request: %v", err)
			}

			responseRecorder := httptest.NewRecorder()
			handler.ServeHTTP(responseRecorder, request)

			assert.Equal(t, tc.expectedCode, responseRecorder.Code)
			if responseRecorder.Code == http.StatusOK {
				jsonResponse, err := json.Marshal(tc.expectedResponse)
				if err != nil {
					t.Fatalf("Failed to marshal response: %v", err)
				}

				assert.Equal(t, jsonResponse, responseRecorder.Body.Bytes())
			}
		})
	}
}
