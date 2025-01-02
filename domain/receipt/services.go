package receipt

import (
	"context"
	"log"
	"math"
	"strconv"
	"unicode"

	"github.com/kevin07696/receipt-processor/domain"
)

type Options struct {
	GenerateID          func(input string) string
	StartPurchaseTime   string
	EndPurchaseTime     string
	TotalMultiple       float64
	ItemsMultiple       uint16
	DescriptionMultiple uint16
}

type Multipliers struct {
	Receipt        uint16
	RoundTotal     uint16
	DivisibleTotal uint16
	Items          float64
	Description    float64
	PurchaseTime   uint16
	PurchaseDate   uint16
}

type ReceiptProcessorService struct {
	repository IReceiptProcessorRepository
	opts       Options
	mults      Multipliers
}

func NewReceiptProcessorService(repository IReceiptProcessorRepository, opts Options, mults Multipliers) ReceiptProcessorService {
	return ReceiptProcessorService{
		repository: repository,
		opts:       opts,
		mults:      mults,
	}
}

type ReceiptProcessorRequest struct {
	Receipt Receipt
	ID      string
}

type ReceiptProcessorResponse struct {
	ID string
}

func (rps *ReceiptProcessorService) GenerateID(ctx context.Context, input string) string {
	return rps.opts.GenerateID(input)
}

func (rps *ReceiptProcessorService) ProcessReceipt(ctx context.Context, request ReceiptProcessorRequest) (ReceiptProcessorResponse, domain.StatusCode) {
	var points uint16
	points += rps.pointsForEachAlphaNumeric(request.Receipt.Retailer)
	points += rps.pointsForEachDivisibleItemDescription(request.Receipt.Items)
	points += rps.pointsForEachItemMultiples(len(request.Receipt.Items))
	points += rps.pointsIfRoundTotal(request.Receipt.Total)
	points += rps.pointsIfDivisibleTotal(request.Receipt.Total)
	points += rps.pointsIfOddPurchaseDate(request.Receipt.PurchaseDate)
	points += rps.pointsIfBetweenPurchaseTime(request.Receipt.PurchaseTime)

	status := rps.repository.WriteReceiptScore(ctx, request.ID, points)
	if status > 0 {
		return ReceiptProcessorResponse{}, status
	}

	return ReceiptProcessorResponse{ID: request.ID}, domain.StatusOK
}

type ReceiptScoreRequest struct {
	ID string
}

type ReceiptScoreResponse struct {
	Points uint16
}

func (rps ReceiptProcessorService) GetReceiptScore(ctx context.Context, request ReceiptScoreRequest) (ReceiptScoreResponse, domain.StatusCode) {
	points, status := rps.repository.ReadReceiptScore(ctx, request.ID)
	if status > 0 {
		return ReceiptScoreResponse{}, domain.ErrNotFound
	}

	return ReceiptScoreResponse{Points: points}, domain.StatusOK
}

func (rps ReceiptProcessorService) pointsForEachAlphaNumeric(name string) uint16 {
	var points uint16
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			points++
		}
	}

	return points * rps.mults.Receipt
}

func (rps ReceiptProcessorService) pointsIfRoundTotal(total string) uint16 {
	n := len(total)
	var decimalPosition int
	for i := n - 1; i > -1; i-- {
		if total[i] == byte('.') {
			decimalPosition = i
		}
	}

	if decimalPosition > 0 && total[decimalPosition+1:] > "00" {
		return 0
	}

	return rps.mults.RoundTotal
}

func (rps ReceiptProcessorService) pointsIfDivisibleTotal(total string) uint16 {
	currency, err := strconv.ParseFloat(total, 64)
	if err != nil {
		log.Fatalf("Failed to parse total, %s: %v", total, err)
	}
	if currency == 0 {
		return 0
	}

	remainder := math.Mod(currency, rps.opts.TotalMultiple)

	if remainder > 0 {
		return 0
	}
	return rps.mults.DivisibleTotal
}

func (rps ReceiptProcessorService) pointsForEachItemMultiples(itemLength int) uint16 {
	multiples := float64(itemLength) / float64(rps.opts.ItemsMultiple)
	return uint16(math.Floor(multiples) * rps.mults.Items)
}

func (rps ReceiptProcessorService) pointsForEachDivisibleItemDescription(items []Item) uint16 {
	var points uint16
	var n, spaces int

Outerloop:
	for _, item := range items {
		spaces = 0
		n = len(item.ShortDescription)
		for i := range item.ShortDescription {
			if item.ShortDescription[i] == byte(' ') {
				spaces++
			} else {
				break
			}
		}

		if spaces == n {
			continue Outerloop
		}

		for i := n - 1; i >= 0; i-- {
			if item.ShortDescription[i] == byte(' ') {
				spaces++
			} else {
				break
			}
		}

		trimmedLength := n - spaces
		if trimmedLength%int(rps.opts.DescriptionMultiple) != 0 {
			continue Outerloop
		}

		price, err := strconv.ParseFloat(item.Price, 64)
		if err != nil {
			log.Fatalf("Failed to parse float, %s: %v", item.Price, err)
		}

		points += uint16(math.Ceil(price * rps.mults.Description))
	}
	return points
}

func (rps ReceiptProcessorService) pointsIfOddPurchaseDate(date string) uint16 {
	dayNum, err := strconv.Atoi(date[8:])
	if err != nil {
		log.Fatalf("Failed to parse day. Check validation method: %v", err)
	}
	if dayNum%2 == 0 {
		return 0
	}

	return rps.mults.PurchaseDate
}

func (rps ReceiptProcessorService) pointsIfBetweenPurchaseTime(time string) uint16 {
	if time[:2] == rps.opts.StartPurchaseTime[:2] && time[3:] == rps.opts.StartPurchaseTime[3:] {
		return 0
	}

	if rps.opts.StartPurchaseTime[:2] > time[:2] || time[:2] >= rps.opts.EndPurchaseTime[:2] {
		return 0
	}

	return rps.mults.PurchaseTime
}
