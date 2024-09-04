package main

import (
	"encoding/json"
	"fmt"

	"github.com/shopspring/decimal"
)

type RateChange struct {
	yearPercent decimal.Decimal
	sinceMonth  int
}

var rateChanges = []RateChange{
	{
		yearPercent: decimal.NewFromFloat(0.067),
		sinceMonth:  0,
	},
	{
		yearPercent: decimal.NewFromFloat(0.0766),
		sinceMonth:  60,
	},
}

func main() {
	credit := decimal.NewFromInt(330000)
	cl := NewCreditLengthFromYears(7)
	//overpay := overPayFlatTotal(5000)
	overpay := overPayConst(decimal.NewFromInt(0))

	var rates []Rate

	//constRateValue := constRateValue(credit, rateChanges[0].yearPercent, cl)
	//periodRates := listRatesWithConstant(RateValue{Value: constRateValue, SinceMonth: rateChanges[0].sinceMonth}, credit, cl, overpay)

	constCreditValue := constantCreditValue(credit, cl)
	periodRates := listRatesWithDecreasing(constCreditValue, credit, cl, overpay)

	rates = append(rates, periodRates...)

	var ratesSummary []Rate
	for _, rateChange := range rateChanges {
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

func listRatesWithConstant(initialConstantRateValue RateValue, credit decimal.Decimal, cl CreditLength, overpay func(decimal.Decimal, decimal.Decimal) decimal.Decimal) []Rate {
	constantRateValue := initialConstantRateValue
	remainingCreditToBePaid := credit

	var totalCapitalPaid, totalInterestPaid decimal.Decimal

	var rates []Rate

	for i := 0; i < cl.Months(); i++ {
		rateChange := findCurrentRateChange(i)

		if constantRateValue.SinceMonth != rateChange.sinceMonth {
			constantRateValue = RateValue{
				Value:      constRateValue(remainingCreditToBePaid, rateChange.yearPercent, cl.AddMonths(-rateChange.sinceMonth)),
				SinceMonth: rateChange.sinceMonth,
			}
		}

		interest := currentInterest(remainingCreditToBePaid, rateChange.yearPercent)
		capital := constantRateValue.Value.Sub(interest)
		capitalPaid := overpay(capital, interest)

		if capitalPaid.GreaterThan(remainingCreditToBePaid) {
			capitalPaid = remainingCreditToBePaid
		}

		totalCapitalPaid = totalCapitalPaid.Add(capitalPaid)
		totalInterestPaid = totalInterestPaid.Add(interest)

		remainingCreditToBePaid = credit.Sub(totalCapitalPaid)

		overpaid := capitalPaid.Add(interest).Sub(constantRateValue.Value)
		if overpaid.LessThan(decimal.NewFromInt(0)) {
			overpaid = decimal.Zero
		}

		rates = append(rates, Rate{
			Value:                   capitalPaid.Add(interest),
			Overpaid:                overpaid,
			CapitalCurrentMonth:     capitalPaid,
			InterestCurrentMonth:    interest,
			ConstRateCurrentMonth:   constantRateValue.Value,
			CurrentMonth:            i,
			TotalCapitalPaid:        totalCapitalPaid,
			TotalInterestPaid:       totalInterestPaid,
			RemainingCreditToBePaid: remainingCreditToBePaid,
		})

		if remainingCreditToBePaid.LessThanOrEqual(decimal.NewFromFloat(0.01)) {
			break
		}
	}

	return rates
}

func listRatesWithDecreasing(initialCapitalValue decimal.Decimal, credit decimal.Decimal, cl CreditLength, overpay func(decimal.Decimal, decimal.Decimal) decimal.Decimal) []Rate {
	remainingCreditToBePaid := credit

	var totalCapitalPaid, totalInterestPaid decimal.Decimal

	var rates []Rate

	for i := 0; i < cl.Months(); i++ {
		rateChange := findCurrentRateChange(i)

		interest := currentInterest(remainingCreditToBePaid, rateChange.yearPercent)
		capital := initialCapitalValue
		capitalPaid := overpay(capital, interest)

		if capitalPaid.GreaterThan(remainingCreditToBePaid) {
			capitalPaid = remainingCreditToBePaid
		}

		totalCapitalPaid = totalCapitalPaid.Add(capitalPaid)
		totalInterestPaid = totalInterestPaid.Add(interest)

		remainingCreditToBePaid = credit.Sub(totalCapitalPaid)

		overpaid := capitalPaid.Sub(initialCapitalValue)
		if overpaid.LessThan(decimal.Zero) {
			overpaid = decimal.Zero
		}

		rates = append(rates, Rate{
			Value:                   capitalPaid.Add(interest),
			Overpaid:                overpaid,
			CapitalCurrentMonth:     capitalPaid,
			InterestCurrentMonth:    interest,
			CurrentMonth:            i,
			TotalCapitalPaid:        totalCapitalPaid,
			TotalInterestPaid:       totalInterestPaid,
			RemainingCreditToBePaid: remainingCreditToBePaid,
		})

		if remainingCreditToBePaid.LessThanOrEqual(decimal.NewFromFloat(0.01)) {
			break
		}
	}

	return rates
}

func findCurrentRateChange(month int) RateChange {
	for i := len(rateChanges) - 1; i >= 0; i-- {
		if month >= rateChanges[i].sinceMonth {
			return rateChanges[i]
		}
	}

	return RateChange{}
}

func decreasingRateValue(credit, yearPercent decimal.Decimal, cl CreditLength) decimal.Decimal {
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

func constRateValue(credit, yearPercent decimal.Decimal, cl CreditLength) decimal.Decimal {
	return credit.Mul(yearPercent).Div(decimal.NewFromInt(12).Mul(decimal.NewFromInt(1).Sub(decimal.NewFromInt(12).Div((yearPercent.Add(decimal.NewFromInt(12)))).Pow(cl.MonthsDecimal()))))
}

func constantCreditValue(v decimal.Decimal, cl CreditLength) decimal.Decimal {
	return v.Div(cl.MonthsDecimal())
}

func currentInterest(v, yearPercent decimal.Decimal) decimal.Decimal {
	return v.Mul(yearPercent).Div(decimal.NewFromInt(12))
}
