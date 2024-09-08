package main

import (
	"fmt"

	"github.com/shopspring/decimal"
)

type MiscCostsAlgorithm interface {
	Calculate(int, ScenarioSummary) decimal.Decimal
}

type MiscCostsFromLoan struct {
	Percentage decimal.Decimal
}

func (m MiscCostsFromLoan) Calculate(month int, s ScenarioSummary) decimal.Decimal {
	return s.Scenario.Loan.Value.Mul(m.Percentage)
}

type MiscCostsSingle struct {
	Cost decimal.Decimal
}

func (m MiscCostsSingle) Calculate(month int, _ ScenarioSummary) decimal.Decimal {
	if month != 0 {
		return decimal.Zero
	}

	return m.Cost
}

type MiscCostsFromRemainingLoan struct {
	Percentage decimal.Decimal

	MonthPeriod int
	UpToMonth   int
}

func (m MiscCostsFromRemainingLoan) Calculate(month int, s ScenarioSummary) decimal.Decimal {
	monthPeriod := m.MonthPeriod
	if monthPeriod == 0 {
		monthPeriod = 1
	}

	if month >= m.UpToMonth {
		return decimal.Zero
	}

	if (month+1)%monthPeriod != 0 {
		return decimal.Zero
	}

	value := s.Scenario.Loan.Value
	if len(s.Rates) != 0 {
		value = s.Rates[len(s.Rates)-1].RemainingLoanToBePaid
	}

	cost := value.Mul(m.Percentage)
	fmt.Printf("month: %d, cost: %s \n", month, cost.Round(2))
	return cost
}
