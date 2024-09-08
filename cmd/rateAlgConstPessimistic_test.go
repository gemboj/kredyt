package main

import (
	"testing"

	"github.com/shopspring/decimal"
	"gotest.tools/v3/assert"
)

func Test_RateAlgConstPessimistic(t *testing.T) {
	loan := Loan{
		Value:  decimal.NewFromInt(100_000),
		Length: NewLoanLengthFromMonths(100),
		InterestRates: []InterestConfig{
			{
				yearPercent: percent(10),
				sinceMonth:  0,
			},
		},
	}

	overpay := OverpayConst{}
	savings := SavingsConst{}

	periodRates := listRatesWithAlgorithm(loan, RateAlgorithmConstantPessimistic{}, overpay, savings)

	assert.DeepEqual(t, decimal.NewFromFloat(47780.73), round(periodRates[len(periodRates)-1].Total.Interest))
}

func Test_RateAlgConstPessimistic_changingInterest(t *testing.T) {
	loan := Loan{
		Value:  decimal.NewFromInt(100_000),
		Length: NewLoanLengthFromMonths(100),
		InterestRates: []InterestConfig{
			{
				yearPercent: percent(10),
				sinceMonth:  0,
			},
			{
				yearPercent: percent(20),
				sinceMonth:  50,
			},
		},
	}

	overpay := OverpayConst{}
	savings := SavingsConst{}

	periodRates := listRatesWithAlgorithm(loan, RateAlgorithmConstantPessimistic{}, overpay, savings)

	assert.DeepEqual(t, decimal.NewFromFloat(63130.64), round(periodRates[len(periodRates)-1].Total.Interest))
}

func Test_RateAlgConstPessimistic_overpayConst(t *testing.T) {
	loan := Loan{
		Value:  decimal.NewFromInt(100_000),
		Length: NewLoanLengthFromMonths(100),
		InterestRates: []InterestConfig{
			{
				yearPercent: percent(10),
				sinceMonth:  0,
			},
		},
	}

	overpay := OverpayConst{}
	savings := SavingsConst{ConstValue: decimal.NewFromInt(1000)}

	periodRates := listRatesWithAlgorithm(loan, RateAlgorithmConstantPessimistic{}, overpay, savings)

	assert.DeepEqual(t, decimal.NewFromFloat(22403.94), round(periodRates[len(periodRates)-1].Total.Interest))
}
