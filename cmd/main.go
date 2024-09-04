package main

import (
	"encoding/json"
	"fmt"
	"math"
)

type RateChange struct {
	yearPercent float64
	sinceMonth  int
}

var rateChanges = []RateChange{
	{
		yearPercent: 0.067,
		sinceMonth:  0,
	},
	{
		yearPercent: 0.0766,
		sinceMonth:  60,
	},
}

func main() {
	credit := float64(330000)
	cl := NewCreditLengthFromYears(7)
	//overpay := overPayFlatTotal(5000)
	overpay := overPayConst(0)

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
	//fmt.Printf("Koszt Kredytu: %v\n", constInstallmentValue*float64(cl.Months())-credit)
	//fmt.Printf("Laczna kwota do splaty: %v\n", constInstallmentValue*float64(credit))

	//startTop := 59
	//top := 3
	//if top > len(installments) {
	//	top = len(installments)
	//}
	ratesJson, _ := json.MarshalIndent(ratesSummary, "", "  ")
	fmt.Printf("Lista Rat: %v\n", string(ratesJson))
}

type RateValue struct {
	Value      float64
	SinceMonth int
}

func overPayConst(c float64) func(float64, float64) float64 {
	return func(capital, _ float64) float64 {
		return capital + c
	}
}

func overPayFlatTotal(flatTotal float64) func(float64, float64) float64 {
	return func(capital, interest float64) float64 {
		if flatTotal < interest {
			return interest
		}

		return flatTotal - interest
	}
}

func listRatesWithConstant(initialConstantRateValue RateValue, credit float64, cl CreditLength, overpay func(float64, float64) float64) []Rate {
	constantRateValue := initialConstantRateValue
	remainingCreditToBePaid := credit

	var totalCapitalPaid, totalInterestPaid float64

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
		capital := round2(constantRateValue.Value - interest)
		capitalPaid := overpay(capital, interest)

		if capitalPaid > remainingCreditToBePaid {
			capitalPaid = remainingCreditToBePaid
		}

		totalCapitalPaid += capitalPaid
		totalInterestPaid += interest

		remainingCreditToBePaid = credit - totalCapitalPaid

		overpaid := capitalPaid + interest - constantRateValue.Value
		if overpaid < 0 {
			overpaid = 0
		}

		rates = append(rates, Rate{
			Value:                   capitalPaid + interest,
			Overpaid:                overpaid,
			CapitalCurrentMonth:     capitalPaid,
			InterestCurrentMonth:    interest,
			ConstRateCurrentMonth:   constantRateValue.Value,
			CurrentMonth:            i,
			TotalCapitalPaid:        totalCapitalPaid,
			TotalInterestPaid:       totalInterestPaid,
			RemainingCreditToBePaid: remainingCreditToBePaid,
		})

		if remainingCreditToBePaid <= 0.01 {
			break
		}
	}

	return rates
}

func listRatesWithDecreasing(initialCapitalValue float64, credit float64, cl CreditLength, overpay func(float64, float64) float64) []Rate {
	remainingCreditToBePaid := credit

	var totalCapitalPaid, totalInterestPaid float64

	var rates []Rate

	for i := 0; i < cl.Months(); i++ {
		rateChange := findCurrentRateChange(i)

		interest := currentInterest(remainingCreditToBePaid, rateChange.yearPercent)
		capital := initialCapitalValue
		capitalPaid := overpay(capital, interest)

		if capitalPaid > remainingCreditToBePaid {
			capitalPaid = remainingCreditToBePaid
		}

		totalCapitalPaid += capitalPaid
		totalInterestPaid += interest

		remainingCreditToBePaid = credit - totalCapitalPaid

		overpaid := capitalPaid - initialCapitalValue
		if overpaid < 0 {
			overpaid = 0
		}

		rates = append(rates, Rate{
			Value:                   capitalPaid + interest,
			Overpaid:                overpaid,
			CapitalCurrentMonth:     capitalPaid,
			InterestCurrentMonth:    interest,
			CurrentMonth:            i,
			TotalCapitalPaid:        totalCapitalPaid,
			TotalInterestPaid:       totalInterestPaid,
			RemainingCreditToBePaid: remainingCreditToBePaid,
		})

		if remainingCreditToBePaid <= 0.01 {
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

func decreasingRateValue(credit, yearPercent float64, cl CreditLength) float64 {
	return round2(credit * yearPercent / (12 * (1 - math.Pow(12/(12+yearPercent), float64(cl.Months())))))
}

func constRateValue(credit, yearPercent float64, cl CreditLength) float64 {
	return round2(credit * yearPercent / (12 * (1 - math.Pow(12/(12+yearPercent), float64(cl.Months())))))
}

func constantCreditValue(v float64, cl CreditLength) float64 {
	return round2(v / float64(cl.Months()))
}

func currentInterest(v, yearPercent float64) float64 {
	return round2(v * yearPercent / 12)
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
