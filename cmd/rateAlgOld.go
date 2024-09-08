package main

import "github.com/shopspring/decimal"

// Pessimistic algorithm takes into account the remaining number of months according to the original deal, when recalculating changing interest rate in the middle of loan.
// e.g. when taking loan for 100 months, and interest changes on month #40, we calculate the new constant rate using arguments:
// 60 months left, remaining loan to be paid, new interest rate
//
// optimistic one would instead calculate how many months were actually left,  taking into account any overpay, and then use the remaining credit length to calculate new constant rate.
func listRatesWithConstantPesimissitc(loan Loan, overpay OverpayAlgorithm) []RateSummary {
	remainingLoanToBePaid := loan.Value

	var totalLoanPaid, totalInterestPaid decimal.Decimal

	var rateSummaries []RateSummary

	for i := 0; i < loan.Length.Months(); i++ {
		interestRate := loan.FindCurrentInterestRate(i)

		constantRateValue := RateValue{
			Value:      constRateValue(loan.Value, interestRate.yearPercent, loan.Length),
			SinceMonth: 0,
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

		rateSummaries = append(rateSummaries, RateSummary{
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
			CurrentMonth:          i,
			RemainingLoanToBePaid: remainingLoanToBePaid,
		})

		if remainingLoanToBePaid.LessThanOrEqual(decimal.NewFromFloat(0.01)) {
			break
		}
	}

	return rateSummaries
}

func listRatesWithDecreasing(loan Loan, overpay OverpayAlgorithm) []RateSummary {
	remainingLoanToBePaid := loan.Value

	var totalLoanPaid, totalInterestPaid decimal.Decimal

	var rateSummaries []RateSummary

	for i := 0; i < loan.Length.Months(); i++ {
		interestRate := loan.FindCurrentInterestRate(i)

		initialInterestThisMonth := remainingLoanToBePaid.Mul(interestRate.MonthPercent())
		initialLoanThisMonth := loan.CalculateConstLoan()

		totalPaidThisMonth := overpay.Overpay(initialLoanThisMonth, initialInterestThisMonth)
		paidLoanThisMonth := totalPaidThisMonth.Sub(initialInterestThisMonth)

		if paidLoanThisMonth.GreaterThan(remainingLoanToBePaid) {
			paidLoanThisMonth = remainingLoanToBePaid
		}

		totalLoanPaid = totalLoanPaid.Add(paidLoanThisMonth)
		totalInterestPaid = totalInterestPaid.Add(initialInterestThisMonth)

		remainingLoanToBePaid = loan.Value.Sub(totalLoanPaid)

		overpaid := paidLoanThisMonth.Sub(loan.CalculateConstLoan())
		if overpaid.LessThan(decimal.Zero) {
			overpaid = decimal.Zero
		}

		rateSummaries = append(rateSummaries, RateSummary{
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
			CurrentMonth:          i,
			RemainingLoanToBePaid: remainingLoanToBePaid,
		})

		if remainingLoanToBePaid.LessThanOrEqual(decimal.NewFromFloat(0.01)) {
			break
		}
	}

	return rateSummaries
}
