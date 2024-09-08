package main

import "github.com/shopspring/decimal"

type RateAlgorithmDecreasing struct {
}

func (r RateAlgorithmDecreasing) calculate(month int, loan Loan, overpay OverpayAlgorithm, rateSummaries []RateSummary) RateSummary {
	initialLoanThisMonth := loan.CalculateConstLoan()

	interestRate := loan.FindCurrentInterestRate(month)

	remainingLoanToBePaid := loan.Value
	totalLoanPaid := decimal.Zero
	totalInterestPaid := decimal.Zero

	if len(rateSummaries) > 0 {
		lastRateSummary := rateSummaries[len(rateSummaries)-1]

		remainingLoanToBePaid = lastRateSummary.RemainingLoanToBePaid
		totalLoanPaid = lastRateSummary.Total.Loan
		totalInterestPaid = lastRateSummary.Total.Interest
	}

	initialInterestThisMonth := remainingLoanToBePaid.Mul(interestRate.MonthPercent())
	totalPaidThisMonth := overpay.Overpay(initialLoanThisMonth, initialInterestThisMonth)
	paidLoanThisMonth := totalPaidThisMonth.Sub(initialInterestThisMonth)

	if paidLoanThisMonth.GreaterThan(remainingLoanToBePaid) {
		paidLoanThisMonth = remainingLoanToBePaid
	}

	totalLoanPaid = totalLoanPaid.Add(paidLoanThisMonth)
	totalInterestPaid = totalInterestPaid.Add(initialInterestThisMonth)

	remainingLoanToBePaid = loan.Value.Sub(totalLoanPaid)

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
		CurrentMonth:          month,
		RemainingLoanToBePaid: remainingLoanToBePaid,
	}
}
