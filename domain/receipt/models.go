package receipt

import "regexp"

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

const (
	idPattern          = "^\\S+$"
	retailerPattern    = "^[\\w\\s\\-&]+$"
	descriptionPattern = "^[\\w\\s\\-]+$"
	timePattern        = "^(0[0-9]|1[0-9]|2[0-3]):([0-5][0-9])$"
	datePattern        = "^[0-9]{4}-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])$"
	currencyPattern    = "^\\d+\\.\\d{2}$"
)

func (r Receipt) Validate() bool {
	for _, i := range r.Items {
		if !i.Validate() {
			return false
		}
	}
	return match(retailerPattern, r.Retailer) &&
		match(timePattern, r.PurchaseTime) &&
		match(datePattern, r.PurchaseDate) &&
		match(currencyPattern, r.Total)
}

func (i Item) Validate() bool {
	return match(descriptionPattern, i.ShortDescription) &&
		match(currencyPattern, i.Price)
}

func (id ID) Validate() bool {
	return match(idPattern, string(id))
}

func match(pattern, input string) bool {
	regex := regexp.MustCompile(pattern)
	return regex.MatchString(input)
}
