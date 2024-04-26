package calculator

import (
	"math"

	"github.com/fnk2077/assessment-tax/tax"
)

func TaxCalculator(req tax.TaxRequest, personalDeduction, maxKReceiptDeduction float64) tax.TaxResponse {
	const maxDonationDecuction = 100000.0
	var taxResponse tax.TaxResponse
	income := req.TotalIncome - personalDeduction

	if len(req.Allowances) > 0 {
		for _, allowance := range req.Allowances {
			if allowance.AllowanceType == "donation" {
				if allowance.Amount > maxDonationDecuction {
					income -= maxDonationDecuction
				} else {
					income -= allowance.Amount
				}
			}
			if allowance.AllowanceType == "k-receipt" {
				if allowance.Amount > maxKReceiptDeduction {
					income -= maxKReceiptDeduction
				} else {
					income -= allowance.Amount
				}
			}
		}
	}

	taxLevels := []struct {
		min   float64
		max   float64
		rate  float64
		level string
	}{
		{0, 150000, 0, "0 - 150,000"},
		{150000, 500000, 0.10, "150,001 - 500,000"},
		{500000, 1000000, 0.15, "500,001 - 1,000,000"},
		{1000000, 2000000, 0.20, "1,000,001 - 2,000,000"},
		{2000000, math.MaxFloat64, 0.30, "2,000,001 ขึ้นไป"},
	}

	totalTax := 0.0
	for _, bracket := range taxLevels {
		if income > bracket.min && income <= bracket.max {
			taxResponse.TaxLevels = append(taxResponse.TaxLevels, tax.TaxLevel{
				Level: bracket.level,
				Tax:   ((income - bracket.min) * bracket.rate),
			})
			totalTax += ((income - bracket.min) * bracket.rate)
		} else if income <= bracket.min {
			taxResponse.TaxLevels = append(taxResponse.TaxLevels, tax.TaxLevel{
				Level: bracket.level,
				Tax:   0.0,
			})
		} else {
			taxResponse.TaxLevels = append(taxResponse.TaxLevels, tax.TaxLevel{
				Level: bracket.level,
				Tax:   ((bracket.max - bracket.min) * bracket.rate),
			})
			totalTax += ((bracket.max - bracket.min) * bracket.rate)
		}
	}

	if totalTax-req.Wht >= 0 {
		taxResponse.Tax = totalTax - req.Wht
	} else {
		taxResponse.TaxRefund = -(totalTax - req.Wht)
	}

	return taxResponse
}
