package main

import (
	"encoding/json"

	"github.com/shopspring/decimal"
)

type RateSummary struct {
	InitalRate          Rate
	PaidRate            Rate
	Total               Rate
	MiscCosts           []decimal.Decimal
	MiscCostsTotal      []decimal.Decimal
	ProjectedLoanLength LoanLength

	SavingsLeftThisMonth decimal.Decimal

	CurrentMonth          int
	RemainingLoanToBePaid decimal.Decimal
}

func (r *RateSummary) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		InitalRate            Rate
		PaidRate              Rate
		Total                 Rate
		MiscCosts             []decimal.Decimal
		MiscCostsTotal        []decimal.Decimal
		ProjectedLoanLength   LoanLength
		SavingsLeftThisMonth  decimal.Decimal
		Overpaid              decimal.Decimal
		CurrentMonth          int
		RemainingLoanToBePaid decimal.Decimal
	}{
		InitalRate:            r.InitalRate,
		PaidRate:              r.PaidRate,
		Total:                 r.Total,
		MiscCosts:             r.MiscCosts,
		MiscCostsTotal:        r.MiscCostsTotal,
		ProjectedLoanLength:   r.ProjectedLoanLength,
		SavingsLeftThisMonth:  round(r.SavingsLeftThisMonth),
		Overpaid:              round(r.PaidRate.Loan.Sub(r.InitalRate.Loan)),
		CurrentMonth:          r.CurrentMonth,
		RemainingLoanToBePaid: round(r.RemainingLoanToBePaid),
	})
}

type Rate struct {
	Loan     decimal.Decimal
	Interest decimal.Decimal
}

func (r Rate) Total() decimal.Decimal {
	return r.Loan.Add(r.Interest)
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

func roundSlice(ds []decimal.Decimal) []decimal.Decimal {
	var out []decimal.Decimal
	for _, d := range ds {
		out = append(out, d.Round(2))
	}

	return out
}
