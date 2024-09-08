package main

import "github.com/shopspring/decimal"

type OverpayAlgorithm interface {
	Overpay(decimal.Decimal, decimal.Decimal) decimal.Decimal
}

// OverPayConst defines overpay as a constant value added to LoanThisMonth every month.
// i.e.  overPayConst(1000) mean we will add 1000 to whatever value we needed to pay.
// if the rateThisMonth is 500, it means we will pay 1500 instead this month, thus the overpay equals 1000.
type OverpayConst struct {
	ConstValue decimal.Decimal
}

func (o OverpayConst) Overpay(loanThisMonth, interest decimal.Decimal) decimal.Decimal {
	return loanThisMonth.Add(interest).Add(o.ConstValue)
}

// OverPayFlatTotal defines overpay as a flat value that will be paid as LoanThisMonth.
// i.e.  overPayFlatTotal(2000) means we will pay 2000 in total this month.
// if the rateThisMonth is 500, it means we will pay 2000 of rate this month (including interest)
// of course the toal value paid needs to be higher than interest.
type OverpayFlatTotal struct {
	FlatTotalValue decimal.Decimal
}

func (o OverpayFlatTotal) Overpay(_, interest decimal.Decimal) decimal.Decimal {
	if o.FlatTotalValue.LessThan(interest) {
		return interest
	}

	return o.FlatTotalValue
}
