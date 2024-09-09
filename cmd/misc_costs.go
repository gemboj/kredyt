package main

import (
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

	RecalculateBaseCostEveryMonth int
	CalculateCostEveryMonth       int
	UpToMonth                     int
}

func (m MiscCostsFromRemainingLoan) Calculate(month int, s ScenarioSummary) decimal.Decimal {
	calculateCostEveryMonth := m.CalculateCostEveryMonth
	recalculateBaseCostEveryMonth := m.RecalculateBaseCostEveryMonth
	upToMonth := m.UpToMonth

	if calculateCostEveryMonth == 0 {
		calculateCostEveryMonth = 1
	}

	if recalculateBaseCostEveryMonth == 0 {
		recalculateBaseCostEveryMonth = 1
	}

	if upToMonth == 0 {
		upToMonth = 999999
	}

	if month >= upToMonth {
		return decimal.Zero
	}

	if (month+1)%calculateCostEveryMonth != 0 {
		return decimal.Zero
	}

	value := s.Scenario.Loan.Value
	if len(s.Rates) != 0 {
		//value = s.Rates[len(s.Rates)-1].RemainingLoanToBePaid
		basePeriod := len(s.Rates) / recalculateBaseCostEveryMonth

		rateIndex := recalculateBaseCostEveryMonth*basePeriod - 1
		if rateIndex >= 0 {
			baseRate := s.Rates[rateIndex]
			value = baseRate.RemainingLoanToBePaid
		}

	}

	cost := value.Mul(m.Percentage)
	return cost.Round(2)
}

type MiscCostsFromMortgage struct {
	Percentage decimal.Decimal

	MonthPeriod int
	UpToMonth   int
}

func (m MiscCostsFromMortgage) Calculate(month int, s ScenarioSummary) decimal.Decimal {
	monthPeriod := m.MonthPeriod
	upToMonth := m.UpToMonth

	if monthPeriod == 0 {
		monthPeriod = 1
	}

	if upToMonth == 0 {
		upToMonth = 999999
	}

	if month >= upToMonth {
		return decimal.Zero
	}

	if (month+1)%monthPeriod != 0 {
		return decimal.Zero
	}

	return s.Scenario.Loan.Mortgage.Mul(m.Percentage).Round(2)
}
