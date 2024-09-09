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

var (
	baselineLoan = Loan{
		Mortgage: decimal.NewFromInt(930_000),
		Value:    decimal.NewFromInt(330_000),
		Length:   NewLoanLengthFromMonths(180),
	}

	baselineAlgorithm = RateAlgorithmConstantPessimistic{}

	loanBaselineVelo = Loan{
		Mortgage: baselineLoan.Mortgage,
		Value:    baselineLoan.Value,
		Length:   baselineLoan.Length,
		InterestRates: []InterestConfig{
			{
				yearPercent: percent(6.76),
				sinceMonth:  0,
			},
			{
				yearPercent: percent(7.42),
				sinceMonth:  60,
			},
		},
	}

	miscCostsBaselineVelo = []MiscCostsAlgorithm{
		MiscCostsFromRemainingLoan{Percentage: percent(0.030129), RecalculateBaseCostEveryMonth: 12},
		MiscCostsFromMortgage{Percentage: percent(0.0426), MonthPeriod: 12},
	}

	loanBaselineING = Loan{
		Mortgage: baselineLoan.Mortgage,
		Value:    baselineLoan.Value,
		Length:   baselineLoan.Length,
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

	miscCostsBaselineING = []MiscCostsAlgorithm{
		MiscCostsFromRemainingLoan{Percentage: percent(0.035), UpToMonth: 36},
		MiscCostsFromLoan{Percentage: percent(0.0096)},
		MiscCostsSingle{Cost: decimal.NewFromInt(560)},
	}

	loanBaselinemBank = Loan{
		Mortgage: baselineLoan.Mortgage,
		Value:    baselineLoan.Value,
		Length:   baselineLoan.Length,
		InterestRates: []InterestConfig{
			{
				yearPercent: percent(6.7),
				sinceMonth:  0,
			},
			{
				yearPercent: percent(7.95),
				sinceMonth:  60,
			},
		},
	}

	miscCostsBaselinemBank = []MiscCostsAlgorithm{
		MiscCostsFromRemainingLoan{Percentage: percent(0.05), UpToMonth: 60},
		MiscCostsFromMortgage{Percentage: percent(0.0065)},
		MiscCostsSingle{Cost: decimal.NewFromInt(400)},
	}
)

func main() {
	scenarios := []Scenario{
		{
			Name:          "Velo baseline",
			Loan:          loanBaselineVelo,
			RateAlgorithm: baselineAlgorithm,
			Overpay:       Overpay{Commission: decimal.NewFromInt(200)},
			Savings:       SavingsFlatTotal{},
			MiscCosts:     miscCostsBaselineVelo,
		},
		findOptimalOverpayScenarioWithCommision(
			Scenario{
				Name:          "Velo optimal overpay 5000",
				Loan:          loanBaselineVelo,
				RateAlgorithm: baselineAlgorithm,
				Overpay:       Overpay{Commission: decimal.NewFromInt(200)},
				Savings:       SavingsFlatTotal{Value: decimal.NewFromInt(5000)},
				MiscCosts:     miscCostsBaselineVelo,
			},
		).Scenario,
		{
			Name:          "ING baseline",
			Loan:          loanBaselineING,
			RateAlgorithm: baselineAlgorithm,
			Overpay:       Overpay{},
			Savings:       SavingsConst{},
			MiscCosts:     miscCostsBaselineING,
		},
		{
			Name:          "ING overpay 5000",
			Loan:          loanBaselineING,
			RateAlgorithm: baselineAlgorithm,
			Overpay:       Overpay{},
			Savings:       SavingsFlatTotal{Value: decimal.NewFromInt(5000)},
			MiscCosts:     miscCostsBaselineING,
		},
		{
			Name:          "mBank baseline",
			Loan:          loanBaselinemBank,
			RateAlgorithm: baselineAlgorithm,
			Overpay:       Overpay{},
			Savings:       SavingsFlatTotal{},
			MiscCosts:     miscCostsBaselinemBank,
		},
		{
			Name:          "mBank overpay 5000",
			Loan:          loanBaselinemBank,
			RateAlgorithm: baselineAlgorithm,
			Overpay:       Overpay{},
			Savings:       SavingsFlatTotal{Value: decimal.NewFromInt(5000)},
			MiscCosts:     miscCostsBaselinemBank,
		},
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
	Name          string
	Loan          Loan
	Overpay       Overpay
	Savings       SavingsAlgorithm
	RateAlgorithm rateAlgorithm
	MiscCosts     []MiscCostsAlgorithm
}

type ScenarioSummary struct {
	Rates    []RateSummary
	Scenario Scenario
}

func displayScenarioComparion(scenarios []ScenarioSummary) {
	fmt.Printf("Length (months): %v\n", baselineLoan.Length.Months())
	fmt.Printf("%v\n", baselineAlgorithm)

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Name", "Total", "TotalInterest", "Months", "FirstRate", "MiscCosts"})
	for i, s := range scenarios {
		rateLast := s.Rates[len(s.Rates)-1]
		rateFirst := s.Rates[0]

		name := s.Scenario.Name
		if s.Scenario.Overpay.PeriodMonths != 0 {
			name += fmt.Sprintf(" /%d", s.Scenario.Overpay.PeriodMonths)
		}

		t.AppendRow([]interface{}{
			i,
			name,
			sum(append(rateLast.MiscCostsTotal, rateLast.Total.Interest)...).Round(2),
			rateLast.Total.Interest.Round(2),
			s.Scenario.Loan.Length.Months(),
			rateFirst.InitalRate.Total().Round(2),
			roundSlice(rateLast.MiscCostsTotal),
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

func sum(ds ...decimal.Decimal) decimal.Decimal {
	out := decimal.Zero
	for _, d := range ds {
		out = out.Add(d)
	}

	return out
}
