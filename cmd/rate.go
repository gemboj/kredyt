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
		Overpaid:              r.PaidRate.Loan.Sub(r.InitalRate.Loan),
		CurrentMonth:          r.CurrentMonth,
		TotalLoanPaid:         r.TotalLoanPaid,
		TotalInterestPaid:     r.TotalInterestPaid,
		RemainingLoanToBePaid: r.RemainingLoanToBePaid,
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
		Total:    r.Loan.Add(r.Interest),
		Loan:     r.Loan,
		Interest: r.Interest,
	})
}
