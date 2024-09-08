package main

/*
Terminology:

Loan - monetary value of the loan taken (excluding any additional fees, like interests, etc.)
LoanCurrentMonth - amount of paid/to be paid of the loan in the current month
Interest - monetary value of Interest to be paid
InterestRates - percentage value used to calculate monetary value of Interest based on Loan
Rate - month's worth of money used to paid the Loan, includes Interest.

*/

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/shopspring/decimal"
)

func main() {
	loanBaseline := Loan{
		Value:  decimal.NewFromInt(330_000),
		Length: NewLoanLengthFromMonths(120),
		InterestRates: []InterestConfig{
			{
				yearPercent: percent(7.07),
				sinceMonth:  0,
			},
			{
				yearPercent: percent(7.59),
				sinceMonth:  60,
			},
		},
	}

	optimalScenarioWithComission := findOptimalOverpayScenarioWithCommision(
		Scenario{
			Loan:          loanBaseline,
			RateAlgorithm: RateAlgorithmConstantPessimistic{},
			Overpay:       Overpay{Commission: decimal.NewFromInt(200)},
			Savings:       SavingsFlatTotal{Value: decimal.NewFromInt(5000)},
		},
	)

	scenarios := []Scenario{
		{
			Loan:          loanBaseline,
			RateAlgorithm: RateAlgorithmConstantPessimistic{},
			Overpay:       Overpay{},
			Savings:       SavingsConst{Value: decimal.NewFromInt(0)},
		},
		{
			Loan:          loanBaseline,
			RateAlgorithm: RateAlgorithmConstantPessimistic{},
			Overpay:       Overpay{Commission: decimal.NewFromInt(200)},
			Savings:       SavingsFlatTotal{},
		},
		{
			Loan:          loanBaseline,
			RateAlgorithm: RateAlgorithmConstantPessimistic{},
			Overpay:       Overpay{},
			Savings:       SavingsFlatTotal{Value: decimal.NewFromInt(5000)},
		},
		optimalScenarioWithComission.Scenario,
	}

	scenarioSummaries := []ScenarioSummary{}
	for _, scenario := range scenarios {
		ratesSummary := listRatesWithAlgorithm(scenario)

		scenarioSummaries = append(scenarioSummaries, ScenarioSummary{
			Rates:    ratesSummary,
			Scenario: scenario,
		})
	}

	displayScenarioComparion(scenarioSummaries)
	//	displayScenarioDetails(scenarioSummaries)
	//	displayScenarioInDepth(scenarioSummaries)
}

type Scenario struct {
	Loan          Loan
	Overpay       Overpay
	Savings       SavingsAlgorithm
	RateAlgorithm rateAlgorithm
}

type ScenarioSummary struct {
	Rates    []RateSummary
	Scenario Scenario
}

func displayScenarioComparion(scenarios []ScenarioSummary) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Total", "Months", "FirstRate", "FirstRateOverpay", "OverpayPeriod"})
	for i, s := range scenarios {
		rateLast := s.Rates[len(s.Rates)-1]
		rateFirst := s.Rates[0]
		t.AppendRow([]interface{}{
			i,
			rateLast.Total.Interest.Round(2),
			s.Scenario.Loan.Length.Months(),
			rateFirst.InitalRate.Total().Round(2),
			rateFirst.PaidRate.Total().Sub(rateFirst.InitalRate.Total()).Round(2),
			s.Scenario.Overpay.PeriodMonths,
		})
	}

	t.Render()
}

func displayScenarioInDepth(scenarios []ScenarioSummary) {
	for _, s := range scenarios {
		rates := s.Rates[:]

		ratesJson, _ := json.MarshalIndent(rates, "", "  ")
		fmt.Printf("RateList: %v\n", string(ratesJson))
	}
}

func displayScenarioDetails(scenarios []ScenarioSummary) {
	for _, s := range scenarios {
		rates := s.Rates
		loan := s.Scenario.Loan

		var ratesSummary []RateSummary
		for _, rateChange := range loan.InterestRates {
			if len(rates) < rateChange.sinceMonth-1 {
				break
			}

			if rateChange.sinceMonth != 0 {
				ratesSummary = append(ratesSummary, rates[rateChange.sinceMonth-1])
			}

			ratesSummary = append(ratesSummary, rates[rateChange.sinceMonth])
		}
		ratesSummary = append(ratesSummary, rates[len(rates)-1])

		ratesJson, _ := json.MarshalIndent(ratesSummary, "", "  ")
		fmt.Printf("RateList: %v\n", string(ratesJson))
	}
}

func percent(p float64) decimal.Decimal {
	return decimal.NewFromFloat(p).Div(decimal.NewFromInt(100))
}

func findOptimalOverpayScenarioWithCommision(scenario Scenario) ScenarioSummary {
	months := 1
	scenario.Overpay.PeriodMonths = months

	previousResult := listRatesWithAlgorithm(scenario)

	for {
		months++

		scenario.Overpay.PeriodMonths = months
		result := listRatesWithAlgorithm(scenario)

		currentTotalInterest := result[len(result)-1].Total.Interest
		previousTotalInterest := previousResult[len(previousResult)-1].Total.Interest
		if currentTotalInterest.GreaterThan(previousTotalInterest) {
			scenario.Overpay.PeriodMonths = months - 1
			return ScenarioSummary{
				Rates:    previousResult,
				Scenario: scenario,
			}
		}

		previousResult = result
	}
}
