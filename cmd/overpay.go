package main

import "github.com/shopspring/decimal"

type OverpayAlgorithm interface {
	Overpay(int, decimal.Decimal, decimal.Decimal) decimal.Decimal
}

// OverPayConst defines overpay as a constant value added to LoanThisMonth every month.
// i.e.  overPayConst(1000) mean we will add 1000 to whatever value we needed to pay.
// if the rateThisMonth is 500, it means we will pay 1500 instead this month, thus the overpay equals 1000.
type OverpayConst struct {
	ConstValue decimal.Decimal

	// By default, if PeriodMonths == 0, overpay every month.
	// Periodmonths == 0 is the same as Periodmonths == 1
	PeriodMonths int

	// Commision cost paid for every overpay. Should be bigger than ConstValue.
	Commission decimal.Decimal
}

func (o OverpayConst) Overpay(month int, loanThisMonth, interestThisMonth decimal.Decimal) decimal.Decimal {
	totalThisMonth := interestThisMonth.Add(loanThisMonth)

	periodMonths := o.PeriodMonths
	if periodMonths == 0 {
		periodMonths = 1
	}

	if month%periodMonths == 1 {
		return totalThisMonth
	}

	return totalThisMonth.Add(o.ConstValue).Sub(o.Commission)
}

// OverPayFlatTotal defines overpay as a flat value that will be paid as LoanThisMonth.
// i.e.  overPayFlatTotal(2000) means we will pay 2000 in total this month.
// if the rateThisMonth is 500, it means we will pay 2000 of rate this month (including interest)
// of course the toal value paid needs to be higher than interest.
type OverpayFlatTotal struct {
	FlatTotalValue decimal.Decimal

	// By default, if PeriodMonths == 0, overpay every month.
	// Periodmonths == 0 is the same as Periodmonths == 1
	PeriodMonths int

	// Commision cost paid for every overpay. Should be bigger than ConstValue.
	Commission decimal.Decimal
}

func (o OverpayFlatTotal) Overpay(month int, loanThisMonth, interestThisMonth decimal.Decimal) decimal.Decimal {
	totalThisMonth := interestThisMonth.Add(loanThisMonth)

	periodMonths := o.PeriodMonths
	if periodMonths == 0 {
		periodMonths = 1
	}

	if month%periodMonths == 1 {
		return totalThisMonth
	}

	if o.FlatTotalValue.LessThan(totalThisMonth) {
		return totalThisMonth
	}

	return o.FlatTotalValue.Sub(o.Commission)
}
