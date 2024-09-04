package main

import "github.com/shopspring/decimal"

type Rate struct {
	Value                   decimal.Decimal
	CapitalCurrentMonth     decimal.Decimal
	Overpaid                decimal.Decimal
	ConstRateCurrentMonth   decimal.Decimal
	InterestCurrentMonth    decimal.Decimal
	CurrentMonth            int
	TotalCapitalPaid        decimal.Decimal
	TotalInterestPaid       decimal.Decimal
	RemainingCreditToBePaid decimal.Decimal
}
