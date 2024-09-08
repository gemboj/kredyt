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
	overpay := OverpayConst{ConstValue: decimal.NewFromInt(0)}

	var rates []RateSummary
	//periodRates := listRatesWithConstantPesimissitc(loan, overpay)
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
