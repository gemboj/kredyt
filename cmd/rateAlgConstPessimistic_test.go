package main

import (
	"testing"

	"github.com/shopspring/decimal"
	"gotest.tools/v3/assert"
)

func Test_RateAlgConstPessimistic(t *testing.T) {
	scenario := Scenario{
		Loan: Loan{
			Value:  decimal.NewFromInt(100_000),
			Length: NewLoanLengthFromMonths(100),
			InterestRates: []InterestConfig{
				{
					yearPercent: percent(10),
					sinceMonth:  0,
				},
			},
		},
		Overpay:       Overpay{},
		Savings:       SavingsConst{},
		RateAlgorithm: RateAlgorithmConstantPessimistic{},
	}

	periodRates := listRatesWithAlgorithm(scenario)

	assert.DeepEqual(t, decimal.NewFromFloat(47780.73), round(periodRates[len(periodRates)-1].Total.Interest))
}

func Test_RateAlgConstPessimistic_changingInterest(t *testing.T) {
	scenario := Scenario{
		Loan: Loan{
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
		},
		Overpay:       Overpay{},
		Savings:       SavingsConst{},
		RateAlgorithm: RateAlgorithmConstantPessimistic{},
	}

	periodRates := listRatesWithAlgorithm(scenario)

	assert.DeepEqual(t, decimal.NewFromFloat(63130.64), round(periodRates[len(periodRates)-1].Total.Interest))
}

func Test_RateAlgConstPessimistic_overpayConst(t *testing.T) {
	scenario := Scenario{
		Loan: Loan{
			Value:  decimal.NewFromInt(100_000),
			Length: NewLoanLengthFromMonths(100),
			InterestRates: []InterestConfig{
				{
					yearPercent: percent(10),
					sinceMonth:  0,
				},
			},
		},
		Overpay:       Overpay{},
		Savings:       SavingsConst{Value: decimal.NewFromInt(1000)},
		RateAlgorithm: RateAlgorithmConstantPessimistic{},
	}
	periodRates := listRatesWithAlgorithm(scenario)

	assert.DeepEqual(t, decimal.NewFromFloat(22403.94), round(periodRates[len(periodRates)-1].Total.Interest))
}
