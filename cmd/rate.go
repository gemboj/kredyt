package main

import (
	"encoding/json"

	"github.com/shopspring/decimal"
)

type RateSummary struct {
	InitalRate Rate
	PaidRate   Rate

	CurrentMonth          int
	TotalLoanPaid         decimal.Decimal
	TotalInterestPaid     decimal.Decimal
	RemainingLoanToBePaid decimal.Decimal
}

func (r *RateSummary) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		InitalRate            Rate
		PaidRate              Rate
		Overpaid              decimal.Decimal
		CurrentMonth          int
		TotalLoanPaid         decimal.Decimal
		TotalInterestPaid     decimal.Decimal
		RemainingLoanToBePaid decimal.Decimal
	}{
		InitalRate:            r.InitalRate,
		PaidRate:              r.PaidRate,
		Overpaid:              round(r.PaidRate.Loan.Sub(r.InitalRate.Loan)),
		CurrentMonth:          r.CurrentMonth,
		TotalLoanPaid:         round(r.TotalLoanPaid),
		TotalInterestPaid:     round(r.TotalInterestPaid),
		RemainingLoanToBePaid: round(r.RemainingLoanToBePaid),
	})
}

type Rate struct {
	Loan     decimal.Decimal
	Interest decimal.Decimal
}

func (r *Rate) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Total    decimal.Decimal
		Loan     decimal.Decimal
		Interest decimal.Decimal
	}{
		Total:    round(r.Loan.Add(r.Interest)),
		Loan:     round(r.Loan),
		Interest: round(r.Interest),
	})
}

func round(d decimal.Decimal) decimal.Decimal {
	return d.Round(2)
}
