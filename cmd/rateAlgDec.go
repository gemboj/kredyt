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

	savedThisMonth := decimal.Zero
	if scenario.Scenario.Savings != nil {
		savedThisMonth = scenario.Scenario.Savings.Savings(month, initialLoanThisMonth, initialInterestThisMonth)
	}

	totalPaidThisMonth, savingsLeftThisMonth := scenario.Scenario.Overpay.Overpay(month, initialLoanThisMonth, initialInterestThisMonth, savedThisMonth.Add(totalSaved))
	paidLoanThisMonth := totalPaidThisMonth.Sub(initialInterestThisMonth)

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

	var miscCosts []decimal.Decimal
	for _, miscCost := range scenario.Scenario.MiscCosts {
		c := miscCost.Calculate(month, scenario)
		miscCosts = append(miscCosts, c)
	}

	miscCostsTotal := make([]decimal.Decimal, len(miscCosts))
	if len(scenario.Rates) > 0 && len(scenario.Rates[len(scenario.Rates)-1].MiscCostsTotal) > 0 {
		miscCostsTotal = scenario.Rates[len(scenario.Rates)-1].MiscCostsTotal
	}

	for i, miscCost := range miscCosts {
		miscCostsTotal[i] = miscCostsTotal[i].Add(miscCost)
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
