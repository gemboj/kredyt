package main

import "github.com/shopspring/decimal"

type RateAlgorithmDecreasing struct {
}

func (r RateAlgorithmDecreasing) calculate(month int, loan Loan, overpay Overpay, savings SavingsAlgorithm, rateSummaries []RateSummary) RateSummary {
	interestRate := loan.FindCurrentInterestRate(month)

	remainingLoanToBePaid := loan.Value
	totalLoanPaid := decimal.Zero
	totalInterestPaid := decimal.Zero
	totalSaved := decimal.Zero

	if len(rateSummaries) > 0 {
		lastRateSummary := rateSummaries[len(rateSummaries)-1]

		remainingLoanToBePaid = lastRateSummary.RemainingLoanToBePaid
		totalLoanPaid = lastRateSummary.Total.Loan
		totalInterestPaid = lastRateSummary.Total.Interest
		totalSaved = lastRateSummary.SavingsLeftThisMonth
	}

	initialLoanThisMonth := loan.CalculateConstLoan()
	initialInterestThisMonth := remainingLoanToBePaid.Mul(interestRate.MonthPercent())

	savedThisMonth := savings.Savings(month, initialLoanThisMonth, initialInterestThisMonth)

	totalPaidThisMonth, savingsLeftThisMonth := overpay.Overpay(month, initialLoanThisMonth, initialInterestThisMonth, savedThisMonth.Add(totalSaved))
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
		SavingsLeftThisMonth:  savingsLeftThisMonth,
		CurrentMonth:          month,
		RemainingLoanToBePaid: remainingLoanToBePaid,
	}
}
