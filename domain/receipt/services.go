package receipt

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"log/slog"
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
	ItemsMultiple       int64
	DescriptionMultiple int64
}

type Multipliers struct {
	Retailer       int64
	RoundTotal     int64
	DivisibleTotal int64
	Items          float64
	Description    float64
	PurchaseTime   int64
	PurchaseDate   int64
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
	var points int64
	points += rps.pointsForEachAlphaNumeric(ctx, request.Receipt.Retailer)
	points += rps.pointsForEachItemMultiples(ctx, len(request.Receipt.Items))
	points += rps.pointsIfRoundTotal(ctx, request.Receipt.Total)
	points += rps.pointsIfDivisibleTotal(ctx, request.Receipt.Total)
	points += rps.pointsIfOddPurchaseDate(ctx, request.Receipt.PurchaseDate)
	points += rps.pointsIfBetweenPurchaseTime(ctx, request.Receipt.PurchaseTime)
	points += rps.pointsForEachDivisibleItemDescription(ctx, request.Receipt.Items)

	slog.InfoContext(ctx, fmt.Sprintf("Total Points: %d", points))

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
	Points int64
}

func (rps ReceiptProcessorService) GetReceiptScore(ctx context.Context, request ReceiptScoreRequest) (ReceiptScoreResponse, domain.StatusCode) {
	points, status := rps.repository.ReadReceiptScore(ctx, request.ID)
	if status > 0 {
		return ReceiptScoreResponse{}, domain.ErrNotFound
	}

	return ReceiptScoreResponse{Points: points}, domain.StatusOK
}

func (rps ReceiptProcessorService) pointsForEachAlphaNumeric(ctx context.Context, name string) int64 {
	var alphaNums int64
	var buf bytes.Buffer
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			alphaNums++
		} else {
			buf.WriteRune(r)
		}
	}

	points := alphaNums * rps.mults.Retailer

	if points > 0 {
		var message string
		if buf.Len() > 0 {
			message = fmt.Sprintf("%d points - retailer name %s has %d alphanumeric characters note: '&' is not alphanumeric", points, name, alphaNums)
		} else {
			message = fmt.Sprintf("%d points - retailer name has %d characters", points, alphaNums)
		}
		slog.DebugContext(ctx, message)
	}

	return points
}

func (rps ReceiptProcessorService) pointsIfRoundTotal(ctx context.Context, total string) int64 {
	n := len(total)

	decimals := total[n-2:]

	if decimals > "00" {
		return 0
	}

	slog.DebugContext(ctx, fmt.Sprintf("%d points - total is a round dollar amount", rps.mults.RoundTotal))

	return rps.mults.RoundTotal
}

func (rps ReceiptProcessorService) pointsIfDivisibleTotal(ctx context.Context, total string) int64 {
	// This should not happen unless total is not properly validated
	currency, err := strconv.ParseFloat(total, 64)
	if err != nil {
		log.Fatalf("Failed to parse total, %s. Check validation: %v", total, err)
	}
	if currency == 0 {
		return 0
	}

	remainder := math.Mod(currency, rps.opts.TotalMultiple)

	if remainder > 0 {
		return 0
	}

	slog.DebugContext(ctx, fmt.Sprintf("%d points - total is a multiple of %.2f", rps.mults.DivisibleTotal, rps.opts.TotalMultiple))

	return rps.mults.DivisibleTotal
}

func (rps ReceiptProcessorService) pointsForEachItemMultiples(ctx context.Context, itemLength int) int64 {
	multiples := float64(itemLength) / float64(rps.opts.ItemsMultiple)

	slog.DebugContext(ctx, fmt.Sprintf("%d items (%d batches @ %.2f points each)", itemLength, rps.opts.ItemsMultiple, rps.mults.Items))
	return int64(math.Floor(multiples) * rps.mults.Items)
}

func (rps ReceiptProcessorService) pointsForEachDivisibleItemDescription(ctx context.Context, items []Item) int64 {
	var total int64
	var n, spaces int

Outerloop:
	for _, item := range items {
		var points int64
		var left, right int
		n = len(item.ShortDescription)
		for i := range item.ShortDescription {
			if item.ShortDescription[i] == byte(' ') {
				left++
			} else {
				break
			}
		}

		if spaces == n {
			continue Outerloop
		}

		for i := n - 1; i >= 0; i-- {
			if item.ShortDescription[i] == byte(' ') {
				right++
			} else {
				break
			}
		}

		trimmedLength := n - left - right
		if trimmedLength%int(rps.opts.DescriptionMultiple) != 0 {
			continue Outerloop
		}
		// This should not happen unless price is not properly validated
		price, err := strconv.ParseFloat(item.Price, 64)
		if err != nil {
			log.Fatalf("Failed to parse float, %s: %v", item.Price, err)
		}

		points = int64(math.Ceil(price * rps.mults.Description))

		trimmedDescription := item.ShortDescription[left : n-right]

		slog.DebugContext(ctx, fmt.Sprintf(`%d Points - "%s" is %d characters (a multiple of %d) item price of %s * %.2f = %.2f is rounded up is %d`,
			points, trimmedDescription, trimmedLength, rps.opts.DescriptionMultiple, item.Price, rps.mults.Description, price*rps.mults.Description, points))

		total += points
	}
	return total
}

func (rps ReceiptProcessorService) pointsIfOddPurchaseDate(ctx context.Context, date string) int64 {
	// This should not happen unless date is not validated properly
	dayNum, err := strconv.Atoi(date[8:])
	if err != nil {
		log.Fatalf("Failed to parse day %s. Check validation: %v", date, err)
	}
	if dayNum%2 == 0 {
		return 0
	}

	slog.DebugContext(ctx, fmt.Sprintf("%d points - purchase day is odd", rps.mults.PurchaseDate))

	return rps.mults.PurchaseDate
}

func (rps ReceiptProcessorService) pointsIfBetweenPurchaseTime(ctx context.Context, time string) int64 {
	if time[:2] == rps.opts.StartPurchaseTime[:2] && time[3:] == rps.opts.StartPurchaseTime[3:] {
		return 0
	}

	if rps.opts.StartPurchaseTime[:2] > time[:2] || time[:2] >= rps.opts.EndPurchaseTime[:2] {
		return 0
	}

	slog.DebugContext(ctx, fmt.Sprintf("%d points - %s is between %s and %s", rps.mults.PurchaseTime, time, rps.opts.StartPurchaseTime, rps.opts.EndPurchaseTime))

	return rps.mults.PurchaseTime
}
