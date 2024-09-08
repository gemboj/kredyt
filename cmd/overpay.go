package main

import "github.com/shopspring/decimal"

type OverpayAlgorithm interface {
	Overpay(int, decimal.Decimal, decimal.Decimal, decimal.Decimal) (decimal.Decimal, decimal.Decimal)
}

type OverpayConst struct {
	// By default, if PeriodMonths == 0, overpay every month.
	// Periodmonths == 0 is the same as Periodmonths == 1
	PeriodMonths int

	// Commision cost paid for every overpay. Should be bigger than ConstValue.
	Commission decimal.Decimal
}

func (o OverpayConst) Overpay(month int, loanThisMonth, interestThisMonth, savingsToUse decimal.Decimal) (decimal.Decimal, decimal.Decimal) {
	totalThisMonth := interestThisMonth.Add(loanThisMonth)

	periodMonths := o.PeriodMonths
	if periodMonths == 0 {
		periodMonths = 1
	}

	if month%periodMonths == 1 {
		return totalThisMonth, savingsToUse
	}

	return totalThisMonth.Add(savingsToUse).Sub(o.Commission), decimal.Zero
}
