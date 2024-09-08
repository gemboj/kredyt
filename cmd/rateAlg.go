package main

import "github.com/shopspring/decimal"

type rateAlgorithm interface {
	calculate(month int, scenario ScenarioSummary) RateSummary
}

func listRatesWithAlgorithm(scenario Scenario) []RateSummary {
	var rates []RateSummary

	for i := 0; i < scenario.Loan.Length.Months(); i++ {
		rate := scenario.RateAlgorithm.calculate(i, ScenarioSummary{Scenario: scenario, Rates: rates})
		rates = append(rates, rate)

		if rate.RemainingLoanToBePaid.LessThanOrEqual(decimal.NewFromFloat(0.01)) {
			break
		}
	}

	return rates
}
