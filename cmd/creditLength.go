package main

type CreditLength struct {
	months int
}

func (cl CreditLength) Months() int {
	return cl.months
}

func (cl CreditLength) Years() int {
	return cl.months / 12
}

func (cl CreditLength) AddMonths(count int) CreditLength {
	return CreditLength{months: cl.months + count}
}

func NewCreditLengthFromYears(year int) CreditLength {
	return CreditLength{months: year * 12}
}

func NewCreditLengthFromMonths(months int) CreditLength {
	return CreditLength{months: months}
}
