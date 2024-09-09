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

	overpaid := paidLoanThisMonth.Add(initialInterestThisMonth).Sub(constantRateValue)
	if overpaid.LessThan(decimal.NewFromInt(0)) {
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

func constRateValue(credit, yearPercent decimal.Decimal, cl LoanLength) decimal.Decimal {
	return credit.Mul(yearPercent).Div(decimal.NewFromInt(12).Mul(decimal.NewFromInt(1).Sub(decimal.NewFromInt(12).Div((yearPercent.Add(decimal.NewFromInt(12)))).Pow(cl.MonthsDecimal()))))
}

func monthInterest(totalCreditLeft, yearPercent decimal.Decimal) decimal.Decimal {
	return totalCreditLeft.Mul(yearPercent).Div(decimal.NewFromInt(12))
}

type RateValue struct {
	Value decimal.Decimal
}
