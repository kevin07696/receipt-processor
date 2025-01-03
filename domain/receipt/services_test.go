package receipt_test

import (
	"context"
	"crypto/sha256"
	"testing"

	"github.com/google/uuid"
	"github.com/kevin07696/receipt-processor/domain"
	"github.com/kevin07696/receipt-processor/domain/receipt"
	"github.com/stretchr/testify/assert"
)

var mults = receipt.Multipliers{
	Retailer:       1,
	RoundTotal:     50,
	DivisibleTotal: 25,
	Items:          5,
	Description:    0.2,
	PurchaseTime:   10,
	PurchaseDate:   6,
}

var opts = receipt.Options{
	GenerateID: func(input string) string {
		if len(input) == 0 {
			return uuid.NewString()
		}

		hash := sha256.Sum256([]byte(input))
		hashBytes := hash[:16]
		hashUUID, err := uuid.FromBytes(hashBytes)
		if err != nil {
			return uuid.NewString()
		}

		return hashUUID.String()
	},
	StartPurchaseTime:   "14:00",
	EndPurchaseTime:     "16:00",
	TotalMultiple:       0.25,
	ItemsMultiple:       2,
	DescriptionMultiple: 3,
}

var mockRepository = MockReceiptRepository{
	WriteReceiptScoreMock: func(ctx context.Context, id string, points uint16, scores map[string]uint16) domain.StatusCode {
		scores[id] = points
		return domain.StatusOK
	},
	ReadReceiptScoreMock: func(ctx context.Context, id string, scores map[string]uint16) (uint16, domain.StatusCode) {
		points, ok := scores[id]
		if !ok {
			return points, domain.ErrNotFound
		}
		return points, domain.StatusOK
	},
}

func TestProcessReceipt(t *testing.T) {
	request := receipt.ReceiptProcessorRequest{
		Receipt: receipt.Receipt{
			Retailer: "",
			Total:    "0.10",
			Items: []receipt.Item{
				{
					ShortDescription: "Mountain Dew 12PK",
					Price:            "6.49",
				},
			},
			PurchaseDate: "2022-01-02",
			PurchaseTime: "12:00",
		},
		ID: "id",
	}

	testCases := []struct {
		title            string
		mockRepository   receipt.IReceiptProcessorRepository
		expectedResponse receipt.ReceiptProcessorResponse
		expectedStatus   domain.StatusCode
	}{
		{
			title: "GivenAValidRequest_ReturnID",
			mockRepository: MockReceiptRepository{
				ReadReceiptScoreMock: func(ctx context.Context, id string, scores map[string]uint16) (uint16, domain.StatusCode) {
					return 0, domain.ErrNotFound
				},
				WriteReceiptScoreMock: func(ctx context.Context, id string, points uint16, scores map[string]uint16) domain.StatusCode {
					return domain.StatusOK
				},
			},
			expectedResponse: receipt.ReceiptProcessorResponse{ID: request.ID},
			expectedStatus:   domain.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			services := receipt.NewReceiptProcessorService(tc.mockRepository, opts, mults)

			response, status := services.ProcessReceipt(context.TODO(), request)

			assert.Equal(t, tc.expectedResponse, response)
			assert.Equal(t, tc.expectedStatus, status)
		})
	}
}

func TestGetRequest(t *testing.T) {
	scores := map[string]uint16{}
	id := opts.GenerateID("receipt data")
	scores[id] = 28

	testCases := []struct {
		title            string
		request          receipt.ReceiptScoreRequest
		expectedResponse receipt.ReceiptScoreResponse
		expectedStatus   domain.StatusCode
	}{
		{
			title:            "GivenAValidRequest_ReturnPoints",
			request:          receipt.ReceiptScoreRequest{ID: id},
			expectedResponse: receipt.ReceiptScoreResponse{Points: 28},
		},
		{
			title:            "GivenAnInvalidID_ReturnNotFoundError",
			request:          receipt.ReceiptScoreRequest{ID: "id"},
			expectedResponse: receipt.ReceiptScoreResponse{},
			expectedStatus:   domain.ErrNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			mockRepository.Scores = scores
			services := receipt.NewReceiptProcessorService(mockRepository, opts, mults)

			response, status := services.GetReceiptScore(context.TODO(), tc.request)

			assert.Equal(t, tc.expectedResponse, response)
			assert.Equal(t, tc.expectedStatus, status)
		})
	}
}

func TestProcessReceiptRules(t *testing.T) {
	testCases := []struct {
		title                 string
		request               receipt.ReceiptProcessorRequest
		expectedScoreResponse receipt.ReceiptScoreResponse
	}{
		{
			title: "GivenValidRequest_ReturnPoints",
			request: receipt.ReceiptProcessorRequest{
				Receipt: receipt.Receipt{
					Retailer:     "Target",
					PurchaseDate: "2022-01-01",
					PurchaseTime: "13:01",
					Items: []receipt.Item{
						{
							ShortDescription: "Mountain Dew 12PK",
							Price:            "6.49",
						},
						{
							ShortDescription: "Emils Cheese Pizza",
							Price:            "12.25",
						},
						{
							ShortDescription: "Knorr Creamy Chicken",
							Price:            "1.26",
						},
						{
							ShortDescription: "Doritos Nacho Cheese",
							Price:            "3.35",
						},
						{
							ShortDescription: "   Klarbrunn 12-PK 12 FL OZ  ",
							Price:            "12.00",
						},
					},
					Total: "35.00",
				},
			},
			expectedScoreResponse: receipt.ReceiptScoreResponse{Points: 28},
		},
		{
			title: "GivenAZeroRequest_Return0",
			request: receipt.ReceiptProcessorRequest{
				Receipt: receipt.Receipt{
					Retailer: "",
					Total:    "0.10",
					Items: []receipt.Item{
						{
							ShortDescription: "Mountain Dew 12PK",
							Price:            "6.49",
						},
					},
					PurchaseDate: "2022-01-02",
					PurchaseTime: "12:00",
				},
			},
			expectedScoreResponse: receipt.ReceiptScoreResponse{Points: 0},
		},
		// Alphanumeric Receipt
		{
			title: "GivenAnAlphaNumericName_Return0",
			request: receipt.ReceiptProcessorRequest{
				Receipt: receipt.Receipt{
					Retailer:     "",
					Total:        "0.10",
					Items:        []receipt.Item{},
					PurchaseDate: "2022-01-02",
					PurchaseTime: "12:00",
				},
			},
			expectedScoreResponse: receipt.ReceiptScoreResponse{Points: 0},
		},
		{
			title: "GivenAnAlphaNumericName_ReturnNumberOfAlphaNumerics",
			request: receipt.ReceiptProcessorRequest{
				Receipt: receipt.Receipt{
					Retailer:     "A3",
					Total:        "0.10",
					Items:        []receipt.Item{},
					PurchaseDate: "2022-01-02",
					PurchaseTime: "12:00",
				},
			},
			expectedScoreResponse: receipt.ReceiptScoreResponse{Points: 2},
		},
		{
			title: "GivenANonAlphaNumeric_ReturnNumberOfAlphaNumerics",
			request: receipt.ReceiptProcessorRequest{
				Receipt: receipt.Receipt{
					Retailer:     "&%A3! ",
					Total:        "0.10",
					Items:        []receipt.Item{},
					PurchaseDate: "2022-01-02",
					PurchaseTime: "12:00",
				},
			},
			expectedScoreResponse: receipt.ReceiptScoreResponse{Points: 2},
		},
		// Total
		{
			title: "GivenTotalIsZero_ReturnPoints",
			request: receipt.ReceiptProcessorRequest{
				Receipt: receipt.Receipt{
					Retailer:     "",
					Total:        "0.00",
					Items:        []receipt.Item{},
					PurchaseDate: "2022-01-02",
					PurchaseTime: "12:00",
				},
			},
			expectedScoreResponse: receipt.ReceiptScoreResponse{Points: 50},
		},
		{
			title: "GivenUnroundTotal_Return0",
			request: receipt.ReceiptProcessorRequest{
				Receipt: receipt.Receipt{
					Retailer:     "",
					Total:        "0.10",
					Items:        []receipt.Item{},
					PurchaseDate: "2022-01-02",
					PurchaseTime: "12:00",
				},
			},
			expectedScoreResponse: receipt.ReceiptScoreResponse{Points: 0},
		},
		{
			title: "GivenTotalIsRound_ReturnRound_DivisiblePoints",
			request: receipt.ReceiptProcessorRequest{
				Receipt: receipt.Receipt{
					Retailer:     "",
					Total:        "1.00",
					Items:        []receipt.Item{},
					PurchaseDate: "2022-01-02",
					PurchaseTime: "12:00",
				},
			},
			expectedScoreResponse: receipt.ReceiptScoreResponse{Points: 75},
		},
		{
			title: "GivenAnInDivisbleDecimalTotal_Return0",
			request: receipt.ReceiptProcessorRequest{
				Receipt: receipt.Receipt{
					Retailer:     "",
					Total:        "1.24",
					Items:        []receipt.Item{},
					PurchaseDate: "2022-01-02",
					PurchaseTime: "12:00",
				},
			},
			expectedScoreResponse: receipt.ReceiptScoreResponse{Points: 0},
		},
		{
			title: "GivenDivisbleDecimalTotal_ReturnDivisiblePoints",
			request: receipt.ReceiptProcessorRequest{
				Receipt: receipt.Receipt{
					Retailer:     "",
					Total:        "1.25",
					Items:        []receipt.Item{},
					PurchaseDate: "2022-01-02",
					PurchaseTime: "12:00",
				},
			},
			expectedScoreResponse: receipt.ReceiptScoreResponse{Points: 25},
		},
		// Items Length
		{
			title: "GivenAZeroItemsLength_Return0",
			request: receipt.ReceiptProcessorRequest{
				Receipt: receipt.Receipt{
					Retailer:     "",
					Total:        "0.10",
					Items:        []receipt.Item{},
					PurchaseDate: "2022-01-02",
					PurchaseTime: "12:00",
				},
			},
			expectedScoreResponse: receipt.ReceiptScoreResponse{Points: 0},
		},
		{
			title: "GivenADivisibleItemsLength_ReturnNumberOfMultiples",
			request: receipt.ReceiptProcessorRequest{
				Receipt: receipt.Receipt{
					Retailer: "",
					Total:    "0.10",
					Items: []receipt.Item{
						{
							ShortDescription: "",
							Price:            "6.49",
						},
						{
							ShortDescription: "",
							Price:            "12.25",
						},
					},
					PurchaseDate: "2022-01-02",
					PurchaseTime: "12:00",
				},
			},
			expectedScoreResponse: receipt.ReceiptScoreResponse{Points: 5},
		},
		{
			title: "GivenAnIndivisibleItemsLength_ReturnNumberOfMultiples",
			request: receipt.ReceiptProcessorRequest{
				Receipt: receipt.Receipt{
					Retailer: "",
					Total:    "0.10",
					Items: []receipt.Item{
						{
							ShortDescription: "",
							Price:            "6.49",
						},
						{
							ShortDescription: "",
							Price:            "12.25",
						},
						{
							ShortDescription: "",
							Price:            "1.26",
						},
					},
					PurchaseDate: "2022-01-02",
					PurchaseTime: "12:00",
				},
			},
			expectedScoreResponse: receipt.ReceiptScoreResponse{Points: 5},
		},
		// Description Length
		{
			title: "GivenASpaceDescriptionLength_Return0",
			request: receipt.ReceiptProcessorRequest{
				Receipt: receipt.Receipt{
					Retailer: "",
					Total:    "0.10",
					Items: []receipt.Item{
						{
							ShortDescription: "  ",
							Price:            "6.49",
						},
					},
					PurchaseDate: "2022-01-02",
					PurchaseTime: "12:00",
				},
			},
			expectedScoreResponse: receipt.ReceiptScoreResponse{Points: 0},
		},
		{
			title: "GivenADivisibleDescriptionLength_ReturnPriceTimesMultiplier",
			request: receipt.ReceiptProcessorRequest{
				Receipt: receipt.Receipt{
					Retailer: "",
					Total:    "0.10",
					Items: []receipt.Item{
						{
							ShortDescription: "cat",
							Price:            "100",
						},
					},
					PurchaseDate: "2022-01-02",
					PurchaseTime: "12:00",
				},
			},
			expectedScoreResponse: receipt.ReceiptScoreResponse{Points: 20},
		},
		{
			title: "GivenADivisibleDescriptionLength_ReturnPriceTimesMultiplier_CeilingRoundedPoints",
			request: receipt.ReceiptProcessorRequest{
				Receipt: receipt.Receipt{
					Retailer: "",
					Total:    "0.10",
					Items: []receipt.Item{
						{
							ShortDescription: "cat",
							Price:            "44",
						},
					},
					PurchaseDate: "2022-01-02",
					PurchaseTime: "12:00",
				},
			},
			expectedScoreResponse: receipt.ReceiptScoreResponse{Points: 9},
		},
		{
			title: "GivenADivisibleWithPaddingDescriptionLength_ReturnPriceTimesMultiplier",
			request: receipt.ReceiptProcessorRequest{
				Receipt: receipt.Receipt{
					Retailer: "",
					Total:    "0.10",
					Items: []receipt.Item{
						{
							ShortDescription: " cat ",
							Price:            "100",
						},
					},
					PurchaseDate: "2022-01-02",
					PurchaseTime: "12:00",
				},
			},
			expectedScoreResponse: receipt.ReceiptScoreResponse{Points: 20},
		},
		{
			title: "GivenAnIndivisibleDescriptionLength_Return0",
			request: receipt.ReceiptProcessorRequest{
				Receipt: receipt.Receipt{
					Retailer: "",
					Total:    "0.10",
					Items: []receipt.Item{
						{
							ShortDescription: "cats",
							Price:            "100",
						},
					},
					PurchaseDate: "2022-01-02",
					PurchaseTime: "12:00",
				},
			},
			expectedScoreResponse: receipt.ReceiptScoreResponse{Points: 0},
		},
		{
			title: "GivenADivisibleWithSpacesDescriptionLength_Return0",
			request: receipt.ReceiptProcessorRequest{
				Receipt: receipt.Receipt{
					Retailer: "",
					Total:    "0.10",
					Items: []receipt.Item{
						{
							ShortDescription: "cats cats",
							Price:            "100",
						},
					},
					PurchaseDate: "2022-01-02",
					PurchaseTime: "12:00",
				},
			},
			expectedScoreResponse: receipt.ReceiptScoreResponse{Points: 20},
		},
		{
			title: "GivenMultipleItemDescriptions_ReturnDesciption_ItemListMultiplePoints",
			request: receipt.ReceiptProcessorRequest{
				Receipt: receipt.Receipt{
					Retailer: "",
					Total:    "0.10",
					Items: []receipt.Item{
						{
							ShortDescription: "cats cats",
							Price:            "100",
						},
						{
							ShortDescription: "cats cat",
							Price:            "100",
						},
						{
							ShortDescription: " ",
							Price:            "100",
						},
					},
					PurchaseDate: "2022-01-02",
					PurchaseTime: "12:00",
				},
			},
			expectedScoreResponse: receipt.ReceiptScoreResponse{Points: 25},
		},
		// Purchase Date
		{
			title: "GivenOddDays_ReturnPoints",
			request: receipt.ReceiptProcessorRequest{
				Receipt: receipt.Receipt{
					Retailer: "",
					Total:    "0.10",
					Items: []receipt.Item{
						{
							ShortDescription: "Mountain Dew 12PK",
							Price:            "6.49",
						},
					},
					PurchaseDate: "2022-01-01",
					PurchaseTime: "12:00",
				},
			},
			expectedScoreResponse: receipt.ReceiptScoreResponse{Points: 6},
		},
		{
			title: "GivenEvenDays_Return0",
			request: receipt.ReceiptProcessorRequest{
				Receipt: receipt.Receipt{
					Retailer: "",
					Total:    "0.10",
					Items: []receipt.Item{
						{
							ShortDescription: "Mountain Dew 12PK",
							Price:            "6.49",
						},
					},
					PurchaseDate: "2022-01-02",
					PurchaseTime: "12:00",
				},
			},
			expectedScoreResponse: receipt.ReceiptScoreResponse{Points: 0},
		},
		// Purchase Time
		{
			title: "GivenBeforeDuration_Return0",
			request: receipt.ReceiptProcessorRequest{
				Receipt: receipt.Receipt{
					Retailer: "",
					Total:    "0.10",
					Items: []receipt.Item{
						{
							ShortDescription: "Mountain Dew 12PK",
							Price:            "6.49",
						},
					},
					PurchaseDate: "2022-01-02",
					PurchaseTime: "12:00",
				},
			},
			expectedScoreResponse: receipt.ReceiptScoreResponse{Points: 0},
		},
		{
			title: "GivenAtMinDuration_Return0",
			request: receipt.ReceiptProcessorRequest{
				Receipt: receipt.Receipt{
					Retailer: "",
					Total:    "0.10",
					Items: []receipt.Item{
						{
							ShortDescription: "Mountain Dew 12PK",
							Price:            "6.49",
						},
					},
					PurchaseDate: "2022-01-02",
					PurchaseTime: "14:00",
				},
			},
			expectedScoreResponse: receipt.ReceiptScoreResponse{Points: 0},
		},
		{
			title: "GivenAtMaxDuration_Return0",
			request: receipt.ReceiptProcessorRequest{
				Receipt: receipt.Receipt{
					Retailer: "",
					Total:    "0.10",
					Items: []receipt.Item{
						{
							ShortDescription: "Mountain Dew 12PK",
							Price:            "6.49",
						},
					},
					PurchaseDate: "2022-01-02",
					PurchaseTime: "16:00",
				},
			},
			expectedScoreResponse: receipt.ReceiptScoreResponse{Points: 0},
		},
		{
			title: "GivenBetweenDuration_ReturnPoints",
			request: receipt.ReceiptProcessorRequest{
				Receipt: receipt.Receipt{
					Retailer: "",
					Total:    "0.10",
					Items: []receipt.Item{
						{
							ShortDescription: "Mountain Dew 12PK",
							Price:            "6.49",
						},
					},
					PurchaseDate: "2022-01-02",
					PurchaseTime: "14:01",
				},
			},
			expectedScoreResponse: receipt.ReceiptScoreResponse{Points: 10},
		},
		{
			title: "GivenBetweenHourDurations_ReturnPoints",
			request: receipt.ReceiptProcessorRequest{
				Receipt: receipt.Receipt{
					Retailer: "",
					Total:    "0.10",
					Items: []receipt.Item{
						{
							ShortDescription: "Mountain Dew 12PK",
							Price:            "6.49",
						},
					},
					PurchaseDate: "2022-01-02",
					PurchaseTime: "15:00",
				},
			},
			expectedScoreResponse: receipt.ReceiptScoreResponse{Points: 10},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			mockRepository.Scores = map[string]uint16{}

			services := receipt.NewReceiptProcessorService(mockRepository, opts, mults)

			id := opts.GenerateID("")
			tc.request.ID = id

			services.ProcessReceipt(context.TODO(), tc.request)

			scoreResponse, _ := services.GetReceiptScore(context.TODO(), receipt.ReceiptScoreRequest{ID: id})

			assert.Equal(t, tc.expectedScoreResponse.Points, scoreResponse.Points)
		})
	}
}
