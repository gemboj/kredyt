package main

import "github.com/shopspring/decimal"

type LoanLength struct {
	months int
}

func (ll LoanLength) Months() int {
	return ll.months
}

func (ll LoanLength) MonthsDecimal() decimal.Decimal {
	return decimal.NewFromInt(int64(ll.months))
}

func (ll LoanLength) Years() int {
	return ll.months / 12
}

func (ll LoanLength) AddMonths(count int) LoanLength {
	return LoanLength{months: ll.months + count}
}

func NewLoanLengthFromYears(year int) LoanLength {
	return LoanLength{months: year * 12}
}

func NewLoanLengthFromMonths(months int) LoanLength {
	return LoanLength{months: months}
}
