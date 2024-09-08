package main

import "github.com/shopspring/decimal"

type RateAlgorithmConstantPessimistic struct {
}

func (r RateAlgorithmConstantPessimistic) calculate(month int, scenario ScenarioSummary) RateSummary {
	interestRate := scenario.Scenario.Loan.FindCurrentInterestRate(month)

	constantRateValue := RateValue{
		Value:      constRateValue(scenario.Scenario.Loan.Value, interestRate.yearPercent, scenario.Scenario.Loan.Length),
		SinceMonth: 0,
	}

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

	if len(scenario.Rates) != 0 && interestRate.sinceMonth > 0 && len(scenario.Rates) >= interestRate.sinceMonth {
		rateSummaryBeforeInterestRateChange := scenario.Rates[interestRate.sinceMonth-1]

		constantRateValue = RateValue{
			Value:      constRateValue(rateSummaryBeforeInterestRateChange.RemainingLoanToBePaid, interestRate.yearPercent, scenario.Scenario.Loan.Length.AddMonths(-interestRate.sinceMonth)),
			SinceMonth: interestRate.sinceMonth,
		}
	}

	initialInterestThisMonth := monthInterest(remainingLoanToBePaid, interestRate.yearPercent)
	initialLoanThisMonth := constantRateValue.Value.Sub(initialInterestThisMonth)

	savedThisMonth := scenario.Scenario.Savings.Savings(month, initialLoanThisMonth, initialInterestThisMonth)

	totalPaidThisMonth, savingsLeftThisMonth := scenario.Scenario.Overpay.Overpay(month, initialLoanThisMonth, initialInterestThisMonth, savedThisMonth.Add(totalSaved))
	paidLoanThisMonth := totalPaidThisMonth.Sub(initialInterestThisMonth)

	if paidLoanThisMonth.GreaterThan(remainingLoanToBePaid) {
		paidLoanThisMonth = remainingLoanToBePaid
	}

	totalLoanPaid = totalLoanPaid.Add(paidLoanThisMonth)
	totalInterestPaid = totalInterestPaid.Add(initialInterestThisMonth)

	remainingLoanToBePaid = scenario.Scenario.Loan.Value.Sub(totalLoanPaid)

	overpaid := paidLoanThisMonth.Add(initialInterestThisMonth).Sub(constantRateValue.Value)
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
	Value      decimal.Decimal
	SinceMonth int
}
