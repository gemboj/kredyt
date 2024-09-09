package main

import "github.com/shopspring/decimal"

type RateAlgorithmDecreasing struct {
}

func (r RateAlgorithmDecreasing) String() string {
	return "RateAlgorithmDecreasing"
}

func (r RateAlgorithmDecreasing) calculate(month int, scenario ScenarioSummary) RateSummary {
	interestRate := scenario.Scenario.Loan.FindCurrentInterestRate(month)

	remainingLoanToBePaid := scenario.Scenario.Loan.Value
	totalLoanPaid := decimal.Zero
	totalInterestPaid := decimal.Zero
	totalSaved := decimal.Zero

	if len(scenario.Rates) > 0 {
		lastRateSummary := scenario.Rates[len(scenario.Rates)-1]

		remainingLoanToBePaid = lastRateSummary.RemainingLoanToBePaid
		totalLoanPaid = lastRateSummary.Total.Loan
		totalInterestPaid = lastRateSummary.Total.Interest
		totalSaved = lastRateSummary.SavingsLeftThisMonth
	}

	initialLoanThisMonth := scenario.Scenario.Loan.CalculateConstLoan()
	initialInterestThisMonth := remainingLoanToBePaid.Mul(interestRate.MonthPercent())

	savedThisMonth := scenario.Scenario.Savings.Savings(month, initialLoanThisMonth.Add(initialInterestThisMonth))

	var miscCostsOutputs []MiscCostsOutput
	for _, miscCost := range scenario.Scenario.MiscCosts {
		miscCostOutput := miscCost.Calculate(month, scenario)

		miscCostsOutputs = append(miscCostsOutputs, miscCostOutput)
	}

	miscCostsTotal := make([]decimal.Decimal, len(miscCostsOutputs))
	if len(scenario.Rates) > 0 && len(scenario.Rates[len(scenario.Rates)-1].MiscCostsTotal) > 0 {
		miscCostsTotal = scenario.Rates[len(scenario.Rates)-1].MiscCostsTotal
	}

	miscCosts := make([]decimal.Decimal, len(miscCostsOutputs))
	for i, miscCost := range miscCostsOutputs {
		if !miscCost.TotalOnly {
			miscCosts[i] = miscCost.Value
		} else {
			miscCosts[i] = decimal.Zero
		}

		miscCostsTotal[i] = miscCostsTotal[i].Add(miscCost.Value)
	}

	miscCostSum := sum(miscCosts...)

	totalPaidThisMonth, savingsLeftThisMonth := scenario.Scenario.Overpay.Overpay(month, initialLoanThisMonth.Add(initialInterestThisMonth).Add(miscCostSum), savedThisMonth.Add(totalSaved))
	paidLoanThisMonth := totalPaidThisMonth.Sub(initialInterestThisMonth).Sub(miscCostSum)

	if paidLoanThisMonth.GreaterThan(remainingLoanToBePaid) {
		paidLoanThisMonth = remainingLoanToBePaid
	}

	totalLoanPaid = totalLoanPaid.Add(paidLoanThisMonth)
	totalInterestPaid = totalInterestPaid.Add(initialInterestThisMonth)

	remainingLoanToBePaid = scenario.Scenario.Loan.Value.Sub(totalLoanPaid)

	overpaid := paidLoanThisMonth.Sub(initialLoanThisMonth)
	if overpaid.LessThan(decimal.Zero) {
		overpaid = decimal.Zero
	}

	return RateSummary{
		InitalRate: Rate{
			Loan:     initialLoanThisMonth,
			Interest: initialInterestThisMonth,
		},
		PaidRate: Rate{
			Loan:     paidLoanThisMonth,
			Interest: initialInterestThisMonth,
		},
		Total: Rate{
			Loan:     totalLoanPaid,
			Interest: totalInterestPaid,
		},
		MiscCosts:             miscCosts,
		MiscCostsTotal:        miscCostsTotal,
		SavingsLeftThisMonth:  savingsLeftThisMonth,
		CurrentMonth:          month,
		RemainingLoanToBePaid: remainingLoanToBePaid,
	}
}
