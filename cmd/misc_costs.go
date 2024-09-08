package main

import "github.com/shopspring/decimal"

type MiscCostsAlgorithm interface {
	Calculate(ScenarioSummary) decimal.Decimal
}

type MiscCostsFromLoan struct {
	Percentage decimal.Decimal
}

func (m MiscCostsFromLoan) Calculate(s ScenarioSummary) decimal.Decimal {
	return s.Scenario.Loan.Value.Mul(m.Percentage)
}
