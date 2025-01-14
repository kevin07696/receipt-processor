package receipt_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kevin07696/receipt-processor/domain"
	receiptDomain "github.com/kevin07696/receipt-processor/domain/receipt"
	receiptHandler "github.com/kevin07696/receipt-processor/handlers/receipt"
	"github.com/stretchr/testify/assert"
)

var receiptAPI receiptDomain.IReceiptProcessorService = &MockReceiptService{
	ProcessReceiptMock: func(ctx context.Context, request receiptDomain.ReceiptProcessorRequest) (receiptDomain.ReceiptProcessorResponse, domain.StatusCode) {
		return receiptDomain.ReceiptProcessorResponse{ID: "ID"}, domain.StatusOK
	},
	GetReceiptScoreMock: func(ctx context.Context, request receiptDomain.ReceiptScoreRequest) (receiptDomain.ReceiptScoreResponse, domain.StatusCode) {
		return receiptDomain.ReceiptScoreResponse{}, domain.StatusOK
	},
	GenerateIDMock: func(ctx context.Context, input string) string {
		return "ID"
	},
}

func TestUnmarshallingRequestBody(t *testing.T) {
	tests := []struct {
		name         string
		requestBody  string
		expectedCode int
	}{
		{
			name:         "GivenAValidRequest_ReturnStatusOK",
			requestBody:  "{ \"retailer\": \"Walgreens\", \"purchaseDate\": \"2022-01-02\", \"purchaseTime\": \"08:13\", \"total\": \"2.65\", \"items\": [ {\"shortDescription\": \"Pepsi - 12-oz\", \"price\": \"1.25\"}, {\"shortDescription\": \"Dasani\", \"price\": \"1.40\"} ] }",
			expectedCode: http.StatusOK,
		},
		{
			name:         "GivenAMalformedRequest_ReturnBadRequestError",
			requestBody:  "{ \"retailer\": \"Walgreens\", \"purchaseDate\": \"2022-01-02\", \"purchaseTime\": \"08:13\", \"total\": \"2.65\", \"items\": [ {\"shortDescription\": \"Pepsi - 12-oz\", \"price\": \"1.25\"}, {\"shortDescription\": \"Dasani\", \"price\": \"1.40\"} ]",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "GivenAnEmptyRequest_ReturnBadRequestError",
			requestBody:  "",
			expectedCode: http.StatusBadRequest,
		},
	}

	handler := receiptHandler.ProcessReceipt(receiptAPI)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request, err := http.NewRequest(http.MethodPost, "/receipts/process", strings.NewReader(tt.requestBody))
			if err != nil {
				t.Fatalf("Failed to build request: %v", err)
			}

			responseRecorder := httptest.NewRecorder()
			handler.ServeHTTP(responseRecorder, request)

			assert.Equal(t, tt.expectedCode, responseRecorder.Code)
		})
	}
}

func TestReceiptProcessorHandler(t *testing.T) {
	tests := []struct {
		name             string
		request          receiptDomain.Receipt
		service          receiptDomain.IReceiptProcessorService
		expectedResponse receiptDomain.ReceiptProcessorResponse
		expectedCode     int
	}{
		{
			name: "GivenAValidRequest_ReturnStatusOK",
			request: receiptDomain.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2024-01-01",
				PurchaseTime: "14:00",
				Items: []receiptDomain.Item{
					{ShortDescription: "desc", Price: "2.00"},
				},
				Total: "2.00",
			},
			service: receiptAPI,
			expectedCode:     http.StatusOK,
			expectedResponse: receiptDomain.ReceiptProcessorResponse{ID: "ID"},
		},
		{
			name: "GivenAValidRequest_ReturnServiceError",
			request: receiptDomain.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2024-01-01",
				PurchaseTime: "14:00",
				Items: []receiptDomain.Item{
					{ShortDescription: "desc", Price: "2.00"},
				},
				Total: "2.00",
			},
			service: &MockReceiptService{
				ProcessReceiptMock: func(ctx context.Context, request receiptDomain.ReceiptProcessorRequest) (receiptDomain.ReceiptProcessorResponse, domain.StatusCode) {
					return receiptDomain.ReceiptProcessorResponse{}, domain.ErrInternal
				},
				GenerateIDMock: func(ctx context.Context, input string) string {
					return ""
				},
			},
			expectedCode:     http.StatusInternalServerError,
			expectedResponse: receiptDomain.ReceiptProcessorResponse{},
		},
		{
			name: "GivenAnEmptyRetailer_ReturnBadRequestError",
			request: receiptDomain.Receipt{
				PurchaseDate: "2024-01-01",
				PurchaseTime: "14:00",
				Items: []receiptDomain.Item{
					{ShortDescription: "desc", Price: "2.00"},
				},
				Total: "2.00",
			},
			service: receiptAPI,
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "GivenAnEmptyRetailer_ReturnBadRequestError",
			request: receiptDomain.Receipt{
				Retailer:     "!@#$%^*()+",
				PurchaseDate: "2024-01-01",
				PurchaseTime: "14:00",
				Items: []receiptDomain.Item{
					{ShortDescription: "desc", Price: "2.00"},
				},
				Total: "2.00",
			},
			service: receiptAPI,
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "GivenAnEmptyPurchaseDate_ReturnBadRequestError",
			request: receiptDomain.Receipt{
				Retailer:     "Target",
				PurchaseDate: "",
				PurchaseTime: "14:00",
				Items: []receiptDomain.Item{
					{ShortDescription: "desc", Price: "2.00"},
				},
				Total: "2.00",
			},
			service: receiptAPI,
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "GivenAnInvalidPurchaseDate_ReturnBadRequestError",
			request: receiptDomain.Receipt{
				Retailer:     "Target",
				PurchaseDate: "202A-01-01",
				PurchaseTime: "14:00",
				Items: []receiptDomain.Item{
					{ShortDescription: "desc", Price: "2.00"},
				},
				Total: "2.00",
			},
			service: receiptAPI,
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "GivenAnInvalidPurchaseDateLength_ReturnBadRequestError",
			request: receiptDomain.Receipt{
				Retailer:     "Target",
				PurchaseDate: "24-01-01",
				PurchaseTime: "14:00",
				Items: []receiptDomain.Item{
					{ShortDescription: "desc", Price: "2.00"},
				},
				Total: "2.00",
			},
			service: receiptAPI,
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "GivenAnEmptyPurchaseTime_ReturnBadRequestError",
			request: receiptDomain.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2024-01-01",
				PurchaseTime: "",
				Items: []receiptDomain.Item{
					{ShortDescription: "desc", Price: "2.00"},
				},
				Total: "2.00",
			},
			service: receiptAPI,
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "GivenAnInvalidPurchaseTime_ReturnBadRequestError",
			request: receiptDomain.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2024-01-01",
				PurchaseTime: "14:0A",
				Items: []receiptDomain.Item{
					{ShortDescription: "desc", Price: "2.00"},
				},
				Total: "2.00",
			},
			service: receiptAPI,
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "GivenAnInvalidPurchaseTimeLength_ReturnBadRequestError",
			request: receiptDomain.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2024-01-01",
				PurchaseTime: "14:00:01",
				Items: []receiptDomain.Item{
					{ShortDescription: "desc", Price: "2.00"},
				},
				Total: "2.00",
			},
			service: receiptAPI,
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "GivenAnEmptyItemList_ReturnBadRequestError",
			request: receiptDomain.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2024-01-01",
				PurchaseTime: "14:00",
				Items:        []receiptDomain.Item{},
				Total:        "0",
			},
			service: receiptAPI,
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "GivenAnEmptyItemDescription_ReturnBadRequestError",
			request: receiptDomain.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2024-01-01",
				PurchaseTime: "14:00",
				Items: []receiptDomain.Item{
					{ShortDescription: "", Price: "2.00"},
				},
				Total: "2.00",
			},
			service: receiptAPI,
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "GivenAnEmptyPrice_ReturnBadRequestError",
			request: receiptDomain.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2024-01-01",
				PurchaseTime: "14:00",
				Items: []receiptDomain.Item{
					{ShortDescription: "desc", Price: ""},
				},
				Total: "2.00",
			},
			service: receiptAPI,
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "GivenAnInvalidPrice_ReturnBadRequestError",
			request: receiptDomain.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2024-01-01",
				PurchaseTime: "14:00",
				Items: []receiptDomain.Item{
					{ShortDescription: "desc", Price: "$7.00"},
				},
				Total: "2.00",
			},
			service: receiptAPI,
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "GivenAnEmptyTotal_ReturnBadRequestError",
			request: receiptDomain.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2024-01-01",
				PurchaseTime: "14:00",
				Items: []receiptDomain.Item{
					{ShortDescription: "desc", Price: "7.00"},
				},
				Total: "",
			},
			service: receiptAPI,
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "GivenAnInvalidTotal_ReturnBadRequestError",
			request: receiptDomain.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2024-01-01",
				PurchaseTime: "14:00",
				Items: []receiptDomain.Item{
					{ShortDescription: "desc", Price: "7.00"},
				},
				Total: "14.A",
			},
			service: receiptAPI,
			expectedCode: http.StatusBadRequest,
		},
	}

	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := receiptHandler.ProcessReceipt(tt.service)
			requestBody, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatalf("Failed to marshall request: %v", err)
			}

			request, err := http.NewRequest(http.MethodPost, "/receipts/process", bytes.NewBuffer(requestBody))
			if err != nil {
				t.Fatalf("Failed to build request: %v", err)
			}

			responseRecorder := httptest.NewRecorder()
			handler.ServeHTTP(responseRecorder, request)

			jsonResponse, err := json.Marshal(tt.expectedResponse)
			if err != nil {
				t.Fatalf("Failed to marshall response: %v", err)
			}

			assert.Equal(t, tt.expectedCode, responseRecorder.Code)
			if responseRecorder.Code == 200 {
				assert.Equal(t, jsonResponse, responseRecorder.Body.Bytes())
			}
		})
	}
}
