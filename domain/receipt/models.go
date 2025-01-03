package receipt

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
)

type Item struct {
	ShortDescription string `json:"shortDescription" validate:"description"`
	Price            string `json:"price" validate:"currency"`
}

type Receipt struct {
	Retailer     string `json:"retailer" validate:"retailer"`
	PurchaseDate string `json:"purchaseDate" validate:"date"`
	PurchaseTime string `json:"purchaseTime" validate:"time"`
	Items        []Item `json:"" validate:"required,min=1,dive,required"`
	Total        string `json:"" validate:"currency"`
}

type ID string

var (
	idPattern          = regexp.MustCompile(`^\S+$`)
	descriptionPattern = regexp.MustCompile(`^[\w\s\-]+$`)
	retailerPattern    = regexp.MustCompile(`^[\w\s\-&]+$`)
	timePattern        = regexp.MustCompile(`^(0[0-9]|1[0-9]|2[0-3]):([0-5][0-9])$`)
	datePattern        = regexp.MustCompile(`^[0-9]{4}-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])$`)
	currencyPattern    = regexp.MustCompile(`^\d+\.\d{2}$`)
)

func match(pattern *regexp.Regexp, value string) bool {
	return pattern.MatchString(value)
}

func (r Receipt) Validate(ctx context.Context) bool {
	var debugMessages []string

	for _, i := range r.Items {
		if !match(descriptionPattern, i.ShortDescription) {
			debugMessages = append(debugMessages, fmt.Sprintf("ShortDescription failed validation: %s", i.ShortDescription))
		}
		if !match(currencyPattern, i.Price) {
			debugMessages = append(debugMessages, fmt.Sprintf("Price failed validation: %s", i.Price))
		}
	}

	if !match(retailerPattern, r.Retailer) {
		debugMessages = append(debugMessages, fmt.Sprintf("Retailer failed validation: %s", r.Retailer))
	}

	if !match(timePattern, r.PurchaseTime) {
		debugMessages = append(debugMessages, fmt.Sprintf("PurchaseTime failed validation: %s", r.PurchaseTime))
	}

	if !match(datePattern, r.PurchaseDate) {
		debugMessages = append(debugMessages, fmt.Sprintf("PurchaseDate failed validation: %s", r.PurchaseDate))
	}

	if !match(currencyPattern, r.Total) {
		debugMessages = append(debugMessages, fmt.Sprintf("Total failed validation: %s", r.Total))
	}

	if len(debugMessages) > 0 {
		slog.DebugContext(ctx, "Receipt failed validation", slog.Any("ReceiptInvalidMsgs", debugMessages))
		return false
	}

	return true
}

func (id ID) Validate() bool {
	return match(idPattern, string(id))
}
