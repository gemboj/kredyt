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
		Value:  decimal.NewFromInt(100_000),
		Length: NewLoanLengthFromMonths(100),
		InterestRates: []InterestConfig{
			{
				yearPercent: decimal.NewFromFloat(0.1),
				sinceMonth:  0,
			},
		},
	}

	//overpay := overPayFlatTotal(5000)
	overpay := OverpayConst{}
	savings := SavingsConst{ConstValue: decimal.NewFromInt(0)}

	var rates []RateSummary
	//periodRates := listRatesWithConstantPesimissitc(loan, overpay)
	//periodRates := listRatesWithDecreasing(loan, overpay)

	periodRates := listRatesWithAlgorithm(loan, RateAlgorithmDecreasing{}, overpay, savings)

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

func percent(p float64) decimal.Decimal {
	return decimal.NewFromFloat(p).Div(decimal.NewFromInt(100))
}
