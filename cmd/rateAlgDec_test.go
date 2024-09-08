package main

import (
	"testing"

	"github.com/shopspring/decimal"
	"gotest.tools/v3/assert"
)

func Test_RateAlgDec(t *testing.T) {
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

	overpay := OverpayConst{ConstValue: decimal.NewFromInt(0)}

	periodRates := listRatesWithAlgorithm(loan, RateAlgorithmDecreasing{}, overpay)

	assert.DeepEqual(t, decimal.NewFromFloat(42083.33), round(periodRates[len(periodRates)-1].Total.Interest))
}

func Test_RateAlgDec_changingInterest(t *testing.T) {
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

	overpay := OverpayConst{ConstValue: decimal.NewFromInt(0)}

	periodRates := listRatesWithAlgorithm(loan, RateAlgorithmDecreasing{}, overpay)

	assert.DeepEqual(t, decimal.NewFromFloat(52708.33), round(periodRates[len(periodRates)-1].Total.Interest))
}

func Test_RateAlgDec_overpayConst(t *testing.T) {
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

	overpay := OverpayConst{ConstValue: decimal.NewFromInt(1000)}

	periodRates := listRatesWithAlgorithm(loan, RateAlgorithmDecreasing{}, overpay)

	assert.DeepEqual(t, decimal.NewFromFloat(21250.00), round(periodRates[len(periodRates)-1].Total.Interest))
}
