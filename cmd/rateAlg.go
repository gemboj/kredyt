package main

import "github.com/shopspring/decimal"

type rateAlgorithm interface {
	calculate(month int, loan Loan, overpay OverpayAlgorithm, savings SavingsAlgorithm, rateSummaries []RateSummary) RateSummary
}

func listRatesWithAlgorithm(loan Loan, alg rateAlgorithm, overpay OverpayAlgorithm, savings SavingsAlgorithm) []RateSummary {
	var rates []RateSummary

	for i := 0; i < loan.Length.Months(); i++ {
		rate := alg.calculate(i, loan, overpay, savings, rates)
		rates = append(rates, rate)

		if rate.RemainingLoanToBePaid.LessThanOrEqual(decimal.NewFromFloat(0.01)) {
			break
		}
	}

	return rates
}
