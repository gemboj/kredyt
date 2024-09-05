package main

import (
	"encoding/json"
	"fmt"

	"github.com/shopspring/decimal"
)

func main() {
	loan := Loan{
		Value:  decimal.NewFromInt(330000),
		Length: NewLoanLengthFromYears(7),
		InterestRates: []InterestRate{
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

	var rates []Rate

	//constRateValue := constRateValue(credit, rateChanges[0].yearPercent, cl)
	//periodRates := listRatesWithConstant(RateValue{Value: constRateValue, SinceMonth: rateChanges[0].sinceMonth}, credit, cl, overpay)

	periodRates := listRatesWithDecreasing(loan, overpay)

	rates = append(rates, periodRates...)

	var ratesSummary []Rate
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

func overPayConst(c decimal.Decimal) func(decimal.Decimal, decimal.Decimal) decimal.Decimal {
	return func(capital, _ decimal.Decimal) decimal.Decimal {
		return capital.Add(c)
	}
}

func overPayFlatTotal(flatTotal decimal.Decimal) func(decimal.Decimal, decimal.Decimal) decimal.Decimal {
	return func(capital, interest decimal.Decimal) decimal.Decimal {
		if flatTotal.LessThan(interest) {
			return interest
		}

		return flatTotal.Sub(interest)
	}
}

func listRatesWithConstant(initialConstantRateValue RateValue, loan Loan, overpay func(decimal.Decimal, decimal.Decimal) decimal.Decimal) []Rate {
	constantRateValue := initialConstantRateValue
	totalLoanLeft := loan.Value

	var totalCapitalPaid, totalInterestPaid decimal.Decimal

	var rates []Rate

	for i := 0; i < loan.Length.Months(); i++ {
		rateChange := loan.FindCurrentInterestRate(i)

		if constantRateValue.SinceMonth != rateChange.sinceMonth {
			constantRateValue = RateValue{
				Value:      constRateValue(totalLoanLeft, rateChange.yearPercent, loan.Length.AddMonths(-rateChange.sinceMonth)),
				SinceMonth: rateChange.sinceMonth,
			}
		}

		interest := monthInterest(totalLoanLeft, rateChange.yearPercent)
		capital := constantRateValue.Value.Sub(interest)
		capitalPaid := overpay(capital, interest)

		if capitalPaid.GreaterThan(totalLoanLeft) {
			capitalPaid = totalLoanLeft
		}

		totalCapitalPaid = totalCapitalPaid.Add(capitalPaid)
		totalInterestPaid = totalInterestPaid.Add(interest)

		totalLoanLeft = loan.Value.Sub(totalCapitalPaid)

		overpaid := capitalPaid.Add(interest).Sub(constantRateValue.Value)
		if overpaid.LessThan(decimal.NewFromInt(0)) {
			overpaid = decimal.Zero
		}

		rates = append(rates, Rate{
			Value:                 capitalPaid.Add(interest),
			Overpaid:              overpaid,
			CapitalCurrentMonth:   capitalPaid,
			InterestCurrentMonth:  interest,
			ConstRateCurrentMonth: constantRateValue.Value,
			CurrentMonth:          i,
			TotalCapitalPaid:      totalCapitalPaid,
			TotalInterestPaid:     totalInterestPaid,
			RemainingLoanToBePaid: totalLoanLeft,
		})

		if totalLoanLeft.LessThanOrEqual(decimal.NewFromFloat(0.01)) {
			break
		}
	}

	return rates
}

type rateAlgorithm interface {
	calculate(month int) Rate
}

type rateAlgorithmDecreasing struct {
}

type rateAlgorithmConstant struct {
}

func (r rateAlgorithmConstant) calculate(month int, loan Loan, remainingLoanToBePaid decimal.Decimal, overpay func(decimal.Decimal, decimal.Decimal) decimal.Decimal, totalLoanPaid, totalInterestPaid decimal.Decimal) Rate {
	return Rate{}
}

func (r rateAlgorithmDecreasing) calculate(month int, loan Loan, remainingLoanToBePaid decimal.Decimal, overpay func(decimal.Decimal, decimal.Decimal) decimal.Decimal, totalLoanPaid, totalInterestPaid decimal.Decimal) Rate {
	constInstallment := loan.CalculateConstInstallment()

	interestRate := loan.FindCurrentInterestRate(month)

	interestPaidThisMonth := remainingLoanToBePaid.Mul(interestRate.MonthPercent())
	loanPaidThisMonth := overpay(constInstallment, interestPaidThisMonth)

	if loanPaidThisMonth.GreaterThan(remainingLoanToBePaid) {
		loanPaidThisMonth = remainingLoanToBePaid
	}

	totalLoanPaid = totalLoanPaid.Add(loanPaidThisMonth)
	totalInterestPaid = totalInterestPaid.Add(interestPaidThisMonth)

	remainingLoanToBePaid = loan.Value.Sub(totalLoanPaid)

	overpaid := loanPaidThisMonth.Sub(constInstallment)
	if overpaid.LessThan(decimal.Zero) {
		overpaid = decimal.Zero
	}

	return Rate{
		Value:                 loanPaidThisMonth.Add(interestPaidThisMonth),
		Overpaid:              overpaid,
		CapitalCurrentMonth:   loanPaidThisMonth,
		InterestCurrentMonth:  interestPaidThisMonth,
		CurrentMonth:          month,
		TotalCapitalPaid:      totalLoanPaid,
		TotalInterestPaid:     totalInterestPaid,
		RemainingLoanToBePaid: remainingLoanToBePaid,
	}
}

func listRatesWithAlgorithm(cl LoanLength, alg rateAlgorithm) []Rate {
	var rates []Rate

	for i := 0; i < cl.Months(); i++ {
		rate := alg.calculate(i)
		rates = append(rates, rate)

		if rate.RemainingLoanToBePaid.LessThanOrEqual(decimal.NewFromFloat(0.01)) {
			break
		}
	}

	return rates
}

func listRatesWithDecreasing(loan Loan, overpay func(decimal.Decimal, decimal.Decimal) decimal.Decimal) []Rate {
	remainingLoanToBePaid := loan.Value

	var totalLoanPaid, totalInterestPaid decimal.Decimal

	var rates []Rate

	for i := 0; i < loan.Length.Months(); i++ {
		interestRate := loan.FindCurrentInterestRate(i)

		interestPaidThisMonth := remainingLoanToBePaid.Mul(interestRate.MonthPercent())
		loanPaidThisMonth := overpay(loan.CalculateConstInstallment(), interestPaidThisMonth)

		if loanPaidThisMonth.GreaterThan(remainingLoanToBePaid) {
			loanPaidThisMonth = remainingLoanToBePaid
		}

		totalLoanPaid = totalLoanPaid.Add(loanPaidThisMonth)
		totalInterestPaid = totalInterestPaid.Add(interestPaidThisMonth)

		remainingLoanToBePaid = loan.Value.Sub(totalLoanPaid)

		overpaid := loanPaidThisMonth.Sub(loan.CalculateConstInstallment())
		if overpaid.LessThan(decimal.Zero) {
			overpaid = decimal.Zero
		}

		rates = append(rates, Rate{
			Value:                 loanPaidThisMonth.Add(interestPaidThisMonth),
			Overpaid:              overpaid,
			CapitalCurrentMonth:   loanPaidThisMonth,
			InterestCurrentMonth:  interestPaidThisMonth,
			CurrentMonth:          i,
			TotalCapitalPaid:      totalLoanPaid,
			TotalInterestPaid:     totalInterestPaid,
			RemainingLoanToBePaid: remainingLoanToBePaid,
		})

		if remainingLoanToBePaid.LessThanOrEqual(decimal.NewFromFloat(0.01)) {
			break
		}
	}

	return rates
}

func decreasingRateValue(credit, yearPercent decimal.Decimal, cl LoanLength) decimal.Decimal {
	return credit.
		Mul(yearPercent).
		Div(
			decimal.NewFromInt(12).
				Mul(
					decimal.NewFromInt(1).
						Sub(
							decimal.NewFromInt(12).
								Div(decimal.NewFromInt(12).Add(yearPercent)).
								Pow(cl.MonthsDecimal()),
						),
				),
		)
}

func constRateValue(credit, yearPercent decimal.Decimal, cl LoanLength) decimal.Decimal {
	return credit.Mul(yearPercent).Div(decimal.NewFromInt(12).Mul(decimal.NewFromInt(1).Sub(decimal.NewFromInt(12).Div((yearPercent.Add(decimal.NewFromInt(12)))).Pow(cl.MonthsDecimal()))))
}

func monthInterest(totalCreditLeft, yearPercent decimal.Decimal) decimal.Decimal {
	return totalCreditLeft.Mul(yearPercent).Div(decimal.NewFromInt(12))
}
