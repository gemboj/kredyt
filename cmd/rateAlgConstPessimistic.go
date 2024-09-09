package main

import "github.com/shopspring/decimal"

type RateAlgorithmConstantPessimistic struct {
}

func (r RateAlgorithmConstantPessimistic) String() string {
	return "RateAlgorithmConstantPessimistic"
}

func (r RateAlgorithmConstantPessimistic) calculate(month int, scenario ScenarioSummary) RateSummary {
	interestRate := scenario.Scenario.Loan.FindCurrentInterestRate(month)

	constantRateValue := constRateValue(scenario.Scenario.Loan.Value, interestRate.yearPercent, scenario.Scenario.Loan.Length)

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

	if len(scenario.Rates) != 0 {
		constantRateValue = scenario.Rates[len(scenario.Rates)-1].InitalRate.Total()

		if interestRate.sinceMonth > 0 && len(scenario.Rates) == interestRate.sinceMonth {
			rateSummaryBeforeInterestRateChange := scenario.Rates[interestRate.sinceMonth-1]

			constantRateValue = constRateValue(rateSummaryBeforeInterestRateChange.RemainingLoanToBePaid, interestRate.yearPercent, scenario.Scenario.Loan.Length.AddMonths(-interestRate.sinceMonth))
		}
	}

	initialInterestThisMonth := monthInterest(remainingLoanToBePaid, interestRate.yearPercent)
	initialLoanThisMonth := constantRateValue.Sub(initialInterestThisMonth)

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

	savedThisMonth := scenario.Scenario.Savings.Savings(month, initialLoanThisMonth.Add(initialInterestThisMonth).Add(miscCostSum))

	totalPaidThisMonth, savingsLeftThisMonth := scenario.Scenario.Overpay.Overpay(month, initialLoanThisMonth.Add(initialInterestThisMonth).Add(miscCostSum), savedThisMonth.Add(totalSaved))
	paidLoanThisMonth := totalPaidThisMonth.Sub(initialInterestThisMonth).Sub(miscCostSum)

	if paidLoanThisMonth.GreaterThan(remainingLoanToBePaid) {
		paidLoanThisMonth = remainingLoanToBePaid
	}

	totalLoanPaid = totalLoanPaid.Add(paidLoanThisMonth)
	totalInterestPaid = totalInterestPaid.Add(initialInterestThisMonth)

	remainingLoanToBePaid = scenario.Scenario.Loan.Value.Sub(totalLoanPaid)

	overpaid := paidLoanThisMonth.Add(initialInterestThisMonth).Sub(constantRateValue)
	if overpaid.LessThan(decimal.NewFromInt(0)) {
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

func constRateValue(credit, yearPercent decimal.Decimal, cl LoanLength) decimal.Decimal {
	return credit.Mul(yearPercent).Div(decimal.NewFromInt(12).Mul(decimal.NewFromInt(1).Sub(decimal.NewFromInt(12).Div((yearPercent.Add(decimal.NewFromInt(12)))).Pow(cl.MonthsDecimal()))))
}

func monthInterest(totalCreditLeft, yearPercent decimal.Decimal) decimal.Decimal {
	return totalCreditLeft.Mul(yearPercent).Div(decimal.NewFromInt(12))
}

type RateValue struct {
	Value decimal.Decimal
}
