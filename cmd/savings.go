package main

import "github.com/shopspring/decimal"

type SavingsAlgorithm interface {
	Savings(int, decimal.Decimal, decimal.Decimal) decimal.Decimal
}

// SavingsConst defines savings as a constant value added to LoanThisMonth every month.
// i.e.  savingsConst(1000) mean we will add 1000 to whatever value we needed to pay.
// if the rateThisMonth is 500, it means we will pay 1500 instead this month, thus the savings equals 1000.
type SavingsConst struct {
	ConstValue decimal.Decimal

	// By default, if PeriodMonths == 0, savings every month.
	// Periodmonths == 0 is the same as Periodmonths == 1
	PeriodMonths int
}

func (o SavingsConst) Savings(month int, _, _ decimal.Decimal) decimal.Decimal {
	periodMonths := o.PeriodMonths
	if periodMonths == 0 {
		periodMonths = 1
	}

	if month%periodMonths == 1 {
		return decimal.Zero
	}

	return o.ConstValue
}

// SavingsFlatTotal defines savings as a flat value that will be paid as LoanThisMonth.
// i.e.  savingsFlatTotal(2000) means we will pay 2000 in total this month.
// if the rateThisMonth is 500, it means we will pay 2000 of rate this month (including interest)
// of course the toal value paid needs to be higher than interest.
type SavingsFlatTotal struct {
	FlatTotalValue decimal.Decimal

	// By default, if PeriodMonths == 0, savings every month.
	// Periodmonths == 0 is the same as Periodmonths == 1
	PeriodMonths int
}

func (o SavingsFlatTotal) Savings(month int, loanThisMonth, interestThisMonth decimal.Decimal) decimal.Decimal {
	totalThisMonth := interestThisMonth.Add(loanThisMonth)

	periodMonths := o.PeriodMonths
	if periodMonths == 0 {
		periodMonths = 1
	}

	if month%periodMonths == 1 {
		return decimal.Zero
	}

	if o.FlatTotalValue.LessThan(totalThisMonth) {
		return decimal.Zero
	}

	return o.FlatTotalValue.Sub(totalThisMonth)
}
