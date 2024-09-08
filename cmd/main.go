package main

/*
Terminology:

Loan - monetary value of the loan taken (excluding any additional fees, like interests, etc.)
LoanCurrentMonth - amount of paid/to be paid of the loan in the current month
Interest - monetary value of Interest to be paid
InterestRates - percentage value used to calculate monetary value of Interest based on Loan
Rate - month's worth of money used to paid the Loan, includes Interest.

*/

import (
	"encoding/json"
	"fmt"

	"github.com/shopspring/decimal"
)

func main() {
	loan := Loan{
		Value:  decimal.NewFromInt(330000),
		Length: NewLoanLengthFromYears(7),
		InterestRates: []InterestConfig{
			{
				yearPercent: decimal.NewFromFloat(0.067),
				sinceMonth:  0,
			},
			{
				yearPercent: decimal.NewFromFloat(0.0766),
				sinceMonth:  60,
			},
		},
	}

	//overpay := overPayFlatTotal(5000)
	overpay := overPayConst(decimal.NewFromInt(0))

	var rates []RateSummary

	//constRateValue := constRateValue(credit, rateChanges[0].yearPercent, cl)
	//periodRates := listRatesWithConstant(RateValue{Value: constRateValue, SinceMonth: rateChanges[0].sinceMonth}, credit, cl, overpay)

	periodRates := listRatesWithDecreasing(loan, overpay)

	rates = append(rates, periodRates...)

	var ratesSummary []RateSummary
	for _, rateChange := range loan.InterestRates {
		if len(rates) < rateChange.sinceMonth-1 {
			break
		}

		if rateChange.sinceMonth != 0 {
			ratesSummary = append(ratesSummary, rates[rateChange.sinceMonth-1])
		}

		ratesSummary = append(ratesSummary, rates[rateChange.sinceMonth])
	}
	ratesSummary = append(ratesSummary, rates[len(rates)-1])

	//fmt.Printf("Kwota kredytu: %v\n", credit)
	//fmt.Printf("Rata stala: %v\n", constInstallmentValue)
	//fmt.Printf("Koszt Kredytu: %v\n", constInstallmentValue*decimal.Decimal(cl.Months())-credit)
	//fmt.Printf("Laczna kwota do splaty: %v\n", constInstallmentValue*decimal.Decimal(credit))

	//startTop := 59
	//top := 3
	//if top > len(installments) {
	//	top = len(installments)
	//}
	ratesJson, _ := json.MarshalIndent(ratesSummary, "", "  ")
	fmt.Printf("RateList: %v\n", string(ratesJson))
}

type RateValue struct {
	Value      decimal.Decimal
	SinceMonth int
}

// overPayConst defines overpay as a constant value added to LoanThisMonth every month.
// i.e.  overPayConst(1000) mean we will add 1000 to whatever value we needed to pay.
// if the loanValueThisMonth is 500, it means we will pay 1500 of loan this month, thus the overpay equals 1000.
func overPayConst(c decimal.Decimal) func(decimal.Decimal, decimal.Decimal) decimal.Decimal {
	return func(loanThisMonth, _ decimal.Decimal) decimal.Decimal {
		return loanThisMonth.Add(c)
	}
}

// overPayFlatTotal defines overpay as a flat value that will be paid as LoanThisMonth.
// i.e.  overPayFlatTotal(2000) means we will pay 2000 in total this month.
// if the rateThisMonth is 500, it means we will pay 2000 of rate this month (including interest)
// of course the toal value paid needs to be higher than interest.
func overPayFlatTotal(flatTotal decimal.Decimal) func(decimal.Decimal, decimal.Decimal) decimal.Decimal {
	return func(_, interestThisMonth decimal.Decimal) decimal.Decimal {
		if flatTotal.LessThan(interestThisMonth) {
			return interestThisMonth
		}

		return flatTotal.Sub(interestThisMonth)
	}
}

func listRatesWithConstant(initialConstantRateValue RateValue, loan Loan, overpay func(decimal.Decimal, decimal.Decimal) decimal.Decimal) []RateSummary {
	constantRateValue := initialConstantRateValue
	remainingLoanToBePaid := loan.Value

	var totalLoanPaid, totalInterestPaid decimal.Decimal

	var rates []RateSummary

	for i := 0; i < loan.Length.Months(); i++ {
		interestRate := loan.FindCurrentInterestRate(i)

		if constantRateValue.SinceMonth != interestRate.sinceMonth {
			constantRateValue = RateValue{
				Value:      constRateValue(remainingLoanToBePaid, interestRate.yearPercent, loan.Length.AddMonths(-interestRate.sinceMonth)),
				SinceMonth: interestRate.sinceMonth,
			}
		}

		initialInterestThisMonth := monthInterest(remainingLoanToBePaid, interestRate.yearPercent)
		initialLoanThisMonth := constantRateValue.Value.Sub(initialInterestThisMonth)
		paidLoanThisMonth := overpay(initialLoanThisMonth, initialInterestThisMonth)

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

		rates = append(rates, RateSummary{
			InitalRate: Rate{
				Loan:     initialLoanThisMonth,
				Interest: initialInterestThisMonth,
			},
			PaidRate: Rate{
				Loan:     paidLoanThisMonth,
				Interest: initialInterestThisMonth,
			},
			CurrentMonth:          i,
			TotalLoanPaid:         totalLoanPaid,
			TotalInterestPaid:     totalInterestPaid,
			RemainingLoanToBePaid: remainingLoanToBePaid,
		})

		if remainingLoanToBePaid.LessThanOrEqual(decimal.NewFromFloat(0.01)) {
			break
		}
	}

	return rates
}

func listRatesWithDecreasing(loan Loan, overpay func(decimal.Decimal, decimal.Decimal) decimal.Decimal) []RateSummary {
	remainingLoanToBePaid := loan.Value

	var totalLoanPaid, totalInterestPaid decimal.Decimal

	var rates []RateSummary

	for i := 0; i < loan.Length.Months(); i++ {
		interestRate := loan.FindCurrentInterestRate(i)

		initialInterestThisMonth := remainingLoanToBePaid.Mul(interestRate.MonthPercent())
		initialLoanThisMonth := loan.CalculateConstLoan()
		paidLoanThisMonth := overpay(initialLoanThisMonth, initialInterestThisMonth)

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

		rates = append(rates, RateSummary{
			InitalRate: Rate{
				Loan:     initialLoanThisMonth,
				Interest: initialInterestThisMonth,
			},
			PaidRate: Rate{
				Loan:     paidLoanThisMonth,
				Interest: initialInterestThisMonth,
			},
			CurrentMonth:          i,
			TotalLoanPaid:         totalLoanPaid,
			TotalInterestPaid:     totalInterestPaid,
			RemainingLoanToBePaid: remainingLoanToBePaid,
		})

		if remainingLoanToBePaid.LessThanOrEqual(decimal.NewFromFloat(0.01)) {
			break
		}
	}

	return rates
}

type rateAlgorithm interface {
	calculate(month int) RateSummary
}

type rateAlgorithmDecreasing struct {
}

type rateAlgorithmConstant struct {
}

func (r rateAlgorithmConstant) calculate(month int, loan Loan, remainingLoanToBePaid decimal.Decimal, overpay func(decimal.Decimal, decimal.Decimal) decimal.Decimal, totalLoanPaid, totalInterestPaid decimal.Decimal) RateSummary {
	return RateSummary{}
}

func (r rateAlgorithmDecreasing) calculate(month int, loan Loan, remainingLoanToBePaid decimal.Decimal, overpay func(decimal.Decimal, decimal.Decimal) decimal.Decimal, totalLoanPaid, totalInterestPaid decimal.Decimal) RateSummary {
	initialLoanThisMonth := loan.CalculateConstLoan()

	interestRate := loan.FindCurrentInterestRate(month)

	initialInterestThisMonth := remainingLoanToBePaid.Mul(interestRate.MonthPercent())
	paidLoanThisMonth := overpay(initialLoanThisMonth, initialInterestThisMonth)

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
		CurrentMonth:          month,
		TotalLoanPaid:         totalLoanPaid,
		TotalInterestPaid:     totalInterestPaid,
		RemainingLoanToBePaid: remainingLoanToBePaid,
	}
}

func listRatesWithAlgorithm(cl LoanLength, alg rateAlgorithm) []RateSummary {
	var rates []RateSummary

	for i := 0; i < cl.Months(); i++ {
		rate := alg.calculate(i)
		rates = append(rates, rate)

		if rate.RemainingLoanToBePaid.LessThanOrEqual(decimal.NewFromFloat(0.01)) {
			break
		}
	}

	return rates
}

func constRateValue(credit, yearPercent decimal.Decimal, cl LoanLength) decimal.Decimal {
	return credit.Mul(yearPercent).Div(decimal.NewFromInt(12).Mul(decimal.NewFromInt(1).Sub(decimal.NewFromInt(12).Div((yearPercent.Add(decimal.NewFromInt(12)))).Pow(cl.MonthsDecimal()))))
}

func monthInterest(totalCreditLeft, yearPercent decimal.Decimal) decimal.Decimal {
	return totalCreditLeft.Mul(yearPercent).Div(decimal.NewFromInt(12))
}
