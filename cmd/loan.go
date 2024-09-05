package main

import "github.com/shopspring/decimal"

type Loan struct {
	Value         decimal.Decimal
	Length        LoanLength
	InterestRates []InterestRate
}

func (l Loan) CalculateConstInstallment() decimal.Decimal {
	return l.Value.Div(l.Length.MonthsDecimal())
}

func (l Loan) FindCurrentInterestRate(month int) InterestRate {
	for i := len(l.InterestRates) - 1; i >= 0; i-- {
		if month >= l.InterestRates[i].sinceMonth {
			return l.InterestRates[i]
		}
	}

	return InterestRate{}
}

type InterestRate struct {
	yearPercent decimal.Decimal
	sinceMonth  int
}

func (r InterestRate) MonthPercent() decimal.Decimal {
	return r.yearPercent.Div(decimal.NewFromInt(12))
}
