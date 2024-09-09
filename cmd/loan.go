package main

import "github.com/shopspring/decimal"

type Loan struct {
	// total value of whatever you are buying
	Mortgage decimal.Decimal

	Value         decimal.Decimal
	Length        LoanLength
	InterestRates []InterestConfig `json:"-"`
}

func (l Loan) CalculateConstLoan() decimal.Decimal {
	return l.Value.Div(l.Length.MonthsDecimal())
}

func (l Loan) FindCurrentInterestRate(month int) InterestConfig {
	for i := len(l.InterestRates) - 1; i >= 0; i-- {
		if month >= l.InterestRates[i].sinceMonth {
			return l.InterestRates[i]
		}
	}

	return InterestConfig{}
}

type InterestConfig struct {
	yearPercent decimal.Decimal
	sinceMonth  int
}

func (r InterestConfig) MonthPercent() decimal.Decimal {
	return r.yearPercent.Div(decimal.NewFromInt(12))
}
