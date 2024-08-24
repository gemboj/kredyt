package main

type Rate struct {
	Value                   float64
	CapitalCurrentMonth     float64
	Overpaid                float64
	ConstRateCurrentMonth   float64
	InterestCurrentMonth    float64
	CurrentMonth            int
	TotalCapitalPaid        float64
	TotalInterestPaid       float64
	RemainingCreditToBePaid float64
}
