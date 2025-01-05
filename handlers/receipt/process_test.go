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
	dReceipt "github.com/kevin07696/receipt-processor/domain/receipt"
	hReceipt "github.com/kevin07696/receipt-processor/handlers/receipt"
	"github.com/stretchr/testify/assert"
)

var receiptAPI dReceipt.IReceiptProcessorService = &MockReceiptService{
	ProcessReceiptMock: func(ctx context.Context, request dReceipt.ReceiptProcessorRequest) (dReceipt.ReceiptProcessorResponse, domain.StatusCode) {
		return dReceipt.ReceiptProcessorResponse{ID: "ID"}, domain.StatusOK
	},
	GetReceiptScoreMock: func(ctx context.Context, request dReceipt.ReceiptScoreRequest) (dReceipt.ReceiptScoreResponse, domain.StatusCode) {
		return dReceipt.ReceiptScoreResponse{}, domain.StatusOK
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

	handler := hReceipt.ProcessReceipt(receiptAPI)

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
		request          dReceipt.Receipt
		service          dReceipt.IReceiptProcessorService
		expectedResponse dReceipt.ReceiptProcessorResponse
		expectedCode     int
	}{
		{
			name: "GivenAValidRequest_ReturnStatusOK",
			request: dReceipt.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2024-01-01",
				PurchaseTime: "14:00",
				Items: []dReceipt.Item{
					{ShortDescription: "desc", Price: "2.00"},
				},
				Total: "2.00",
			},
			service: receiptAPI,
			expectedCode:     http.StatusOK,
			expectedResponse: dReceipt.ReceiptProcessorResponse{ID: "ID"},
		},
		{
			name: "GivenAValidRequest_ReturnServiceError",
			request: dReceipt.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2024-01-01",
				PurchaseTime: "14:00",
				Items: []dReceipt.Item{
					{ShortDescription: "desc", Price: "2.00"},
				},
				Total: "2.00",
			},
			service: &MockReceiptService{
				ProcessReceiptMock: func(ctx context.Context, request dReceipt.ReceiptProcessorRequest) (dReceipt.ReceiptProcessorResponse, domain.StatusCode) {
					return dReceipt.ReceiptProcessorResponse{}, domain.ErrInternal
				},
				GenerateIDMock: func(ctx context.Context, input string) string {
					return ""
				},
			},
			expectedCode:     http.StatusInternalServerError,
			expectedResponse: dReceipt.ReceiptProcessorResponse{},
		},
		{
			name: "GivenAnEmptyRetailer_ReturnBadRequestError",
			request: dReceipt.Receipt{
				PurchaseDate: "2024-01-01",
				PurchaseTime: "14:00",
				Items: []dReceipt.Item{
					{ShortDescription: "desc", Price: "2.00"},
				},
				Total: "2.00",
			},
			service: receiptAPI,
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "GivenAnEmptyRetailer_ReturnBadRequestError",
			request: dReceipt.Receipt{
				Retailer:     "!@#$%^*()+",
				PurchaseDate: "2024-01-01",
				PurchaseTime: "14:00",
				Items: []dReceipt.Item{
					{ShortDescription: "desc", Price: "2.00"},
				},
				Total: "2.00",
			},
			service: receiptAPI,
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "GivenAnEmptyPurchaseDate_ReturnBadRequestError",
			request: dReceipt.Receipt{
				Retailer:     "Target",
				PurchaseDate: "",
				PurchaseTime: "14:00",
				Items: []dReceipt.Item{
					{ShortDescription: "desc", Price: "2.00"},
				},
				Total: "2.00",
			},
			service: receiptAPI,
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "GivenAnInvalidPurchaseDate_ReturnBadRequestError",
			request: dReceipt.Receipt{
				Retailer:     "Target",
				PurchaseDate: "202A-01-01",
				PurchaseTime: "14:00",
				Items: []dReceipt.Item{
					{ShortDescription: "desc", Price: "2.00"},
				},
				Total: "2.00",
			},
			service: receiptAPI,
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "GivenAnInvalidPurchaseDateLength_ReturnBadRequestError",
			request: dReceipt.Receipt{
				Retailer:     "Target",
				PurchaseDate: "24-01-01",
				PurchaseTime: "14:00",
				Items: []dReceipt.Item{
					{ShortDescription: "desc", Price: "2.00"},
				},
				Total: "2.00",
			},
			service: receiptAPI,
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "GivenAnEmptyPurchaseTime_ReturnBadRequestError",
			request: dReceipt.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2024-01-01",
				PurchaseTime: "",
				Items: []dReceipt.Item{
					{ShortDescription: "desc", Price: "2.00"},
				},
				Total: "2.00",
			},
			service: receiptAPI,
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "GivenAnInvalidPurchaseTime_ReturnBadRequestError",
			request: dReceipt.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2024-01-01",
				PurchaseTime: "14:0A",
				Items: []dReceipt.Item{
					{ShortDescription: "desc", Price: "2.00"},
				},
				Total: "2.00",
			},
			service: receiptAPI,
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "GivenAnInvalidPurchaseTimeLength_ReturnBadRequestError",
			request: dReceipt.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2024-01-01",
				PurchaseTime: "14:00:01",
				Items: []dReceipt.Item{
					{ShortDescription: "desc", Price: "2.00"},
				},
				Total: "2.00",
			},
			service: receiptAPI,
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "GivenAnEmptyItemList_ReturnBadRequestError",
			request: dReceipt.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2024-01-01",
				PurchaseTime: "14:00",
				Items:        []dReceipt.Item{},
				Total:        "0",
			},
			service: receiptAPI,
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "GivenAnEmptyItemDescription_ReturnBadRequestError",
			request: dReceipt.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2024-01-01",
				PurchaseTime: "14:00",
				Items: []dReceipt.Item{
					{ShortDescription: "", Price: "2.00"},
				},
				Total: "2.00",
			},
			service: receiptAPI,
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "GivenAnEmptyPrice_ReturnBadRequestError",
			request: dReceipt.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2024-01-01",
				PurchaseTime: "14:00",
				Items: []dReceipt.Item{
					{ShortDescription: "desc", Price: ""},
				},
				Total: "2.00",
			},
			service: receiptAPI,
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "GivenAnInvalidPrice_ReturnBadRequestError",
			request: dReceipt.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2024-01-01",
				PurchaseTime: "14:00",
				Items: []dReceipt.Item{
					{ShortDescription: "desc", Price: "$7.00"},
				},
				Total: "2.00",
			},
			service: receiptAPI,
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "GivenAnEmptyTotal_ReturnBadRequestError",
			request: dReceipt.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2024-01-01",
				PurchaseTime: "14:00",
				Items: []dReceipt.Item{
					{ShortDescription: "desc", Price: "7.00"},
				},
				Total: "",
			},
			service: receiptAPI,
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "GivenAnInvalidTotal_ReturnBadRequestError",
			request: dReceipt.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2024-01-01",
				PurchaseTime: "14:00",
				Items: []dReceipt.Item{
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
			handler := hReceipt.ProcessReceipt(tt.service)
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
