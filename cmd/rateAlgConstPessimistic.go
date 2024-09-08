package main

import "github.com/shopspring/decimal"

type rateAlgorithmConstantPessimistic struct {
}

func (r rateAlgorithmConstantPessimistic) calculate(month int, loan Loan, overpay OverpayAlgorithm, rateSummaries []RateSummary) RateSummary {
	interestRate := loan.FindCurrentInterestRate(month)

	constantRateValue := RateValue{
		Value:      constRateValue(loan.Value, interestRate.yearPercent, loan.Length),
		SinceMonth: 0,
	}

	remainingLoanToBePaid := loan.Value
	totalLoanPaid := decimal.Zero
	totalInterestPaid := decimal.Zero

	if len(rateSummaries) > 0 {
		lastRateSummary := rateSummaries[len(rateSummaries)]
		remainingLoanToBePaid = lastRateSummary.RemainingLoanToBePaid
		totalLoanPaid = lastRateSummary.Total.Loan
		totalInterestPaid = lastRateSummary.Total.Interest
	}

	if len(rateSummaries) != 0 && interestRate.sinceMonth > 0 && len(rateSummaries) >= interestRate.sinceMonth {
		rateSummaryBeforeInterestRateChange := rateSummaries[interestRate.sinceMonth-1]

		constantRateValue = RateValue{
			Value:      constRateValue(rateSummaryBeforeInterestRateChange.RemainingLoanToBePaid, interestRate.yearPercent, loan.Length.AddMonths(-interestRate.sinceMonth)),
			SinceMonth: interestRate.sinceMonth,
		}
	}

	initialInterestThisMonth := monthInterest(remainingLoanToBePaid, interestRate.yearPercent)
	initialLoanThisMonth := constantRateValue.Value.Sub(initialInterestThisMonth)
	totalPaidThisMonth := overpay.Overpay(initialLoanThisMonth, initialInterestThisMonth)
	paidLoanThisMonth := totalPaidThisMonth.Sub(initialInterestThisMonth)

	if paidLoanThisMonth.GreaterThan(remainingLoanToBePaid) {
		paidLoanThisMonth = remainingLoanToBePaid
	}

	totalLoanPaid = totalLoanPaid.Add(paidLoanThisMonth)
	totalInterestPaid = totalInterestPaid.Add(initialInterestThisMonth)

	remainingLoanToBePaid = loan.Value.Sub(totalLoanPaid)

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
