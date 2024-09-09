package main

import (
	"fmt"

	"github.com/shopspring/decimal"
)

type rateAlgorithm interface {
	calculate(month int, scenario ScenarioSummary) RateSummary
}

func listRatesWithAlgorithm(scenario Scenario) []RateSummary {
	var rates []RateSummary

	return calculateForRates(scenario, rates)
}

func calculateForRates(scenario Scenario, rates []RateSummary) []RateSummary {
	for i := len(rates); i < scenario.Loan.Length.Months(); i++ {
		rate := scenario.RateAlgorithm.calculate(i, ScenarioSummary{Scenario: scenario, Rates: rates})

		projectedLoanLength := scenario.Loan.Length.AddMonths(-1)
		if scenario.Savings != nil {
			scenarioWithoutOverpay := scenario
			scenarioWithoutOverpay.Savings = nil
			ratesWithoutOverpay := calculateForRates(scenarioWithoutOverpay, append(rates, rate))
			projectedLoanLength = NewLoanLengthFromMonths(len(ratesWithoutOverpay))
		}

		rate.ProjectedLoanLength = projectedLoanLength
		if scenario.Savings != nil {
			fmt.Printf("Month: %v, ProjectedLength: %v\n", i, rate.ProjectedLoanLength)
		}
		rates = append(rates, rate)

		if rate.RemainingLoanToBePaid.LessThanOrEqual(decimal.NewFromFloat(0.01)) {
			break
		}
	}

	return rates
}
