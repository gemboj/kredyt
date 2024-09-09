package main

import "github.com/shopspring/decimal"

type Overpay struct {
	// By default, if PeriodMonths == 0, overpay every month.
	// Periodmonths == 0 is the same as Periodmonths == 1
	PeriodMonths int

	// Commision cost paid for every overpay. Should be bigger than ConstValue.
	Commission decimal.Decimal
}

func (o Overpay) Overpay(month int, totalThisMonth, savingsToUse decimal.Decimal) (decimal.Decimal, decimal.Decimal) {
	periodMonths := o.PeriodMonths
	if periodMonths == 0 {
		periodMonths = 1
	}

	if (month+1)%(periodMonths) != 0 || savingsToUse.Equal(decimal.Zero) {
		return totalThisMonth, savingsToUse
	}

	return totalThisMonth.Add(savingsToUse).Sub(o.Commission), decimal.Zero
}
