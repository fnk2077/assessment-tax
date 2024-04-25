package postgres

import (
	"math"

	"github.com/fnk2077/assessment-tax/tax"
)

type Allowance struct {
	AllowanceType string  `json:"allowanceType"`
	Amount        float64 `json:"amount"`
}

type TaxLevel struct {
	Level     string  `json:"level"`
	Tax       float64 `json:"tax"`
	TaxRefund float64 `json:"taxRefund,omitempty"`
}

type TaxRequest struct {
	TotalIncome float64     `json:"totalIncome"`
	Wht         float64     `json:"wht"`
	Allowances  []Allowance `json:"allowances"`
}

type TaxResponse struct {
	Tax       float64    `json:"tax"`
	TaxRefund float64    `json:"taxRefund,omitempty"`
	TaxLevels []TaxLevel `json:"taxLevel"`
}

type TaxCSVRequest struct {
	TotalIncome float64 `json:"totalIncome"`
	Wht         float64 `json:"wht"`
	Donation    float64 `json:"donation"`
}

type TaxCSVResponse struct {
	Taxes []TaxCSVResponseDetail `json:"taxes"`
}

type TaxCSVResponseDetail struct {
	TotalIncome float64 `json:"totalIncome"`
	Tax         float64 `json:"tax"`
	TaxRefund   float64 `json:"taxRefund,omitempty"`
}

func (p *Postgres) ChangeDeduction(amount float64, deductionType string) error {

	if deductionType == "personal" {
		query := `INSERT INTO deductions (personal, max_kreceipt) 
          SELECT $1, max_kreceipt FROM deductions ORDER BY id DESC LIMIT 1`

		_, err := p.Db.Exec(query, amount)
		if err != nil {
			return err
		}
	} else if deductionType == "k-receipt" {
		query := `INSERT INTO deductions (max_kreceipt, personal) 
		SELECT $1, personal FROM deductions ORDER BY id DESC LIMIT 1`

		_, err := p.Db.Exec(query, amount)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Postgres) TaxCalculate(req tax.TaxRequest) (tax.TaxResponse, error) {
	var personalDeduction float64
	var maxKReceiptDeduction float64
	p.Db.QueryRow(`SELECT personal FROM deductions ORDER BY id DESC LIMIT 1`).Scan(&personalDeduction)
	p.Db.QueryRow(`SELECT max_kreceipt FROM deductions ORDER BY id DESC LIMIT 1`).Scan(&maxKReceiptDeduction)

	taxResponse := taxCalculator(req, personalDeduction, maxKReceiptDeduction)

	return taxResponse, nil
}

func (p *Postgres) TaxCSVCalculate(reqs []tax.TaxCSVRequest) (tax.TaxCSVResponse, error) {
	var taxCSVResponse tax.TaxCSVResponse
	var personalDeduction float64
	var maxKReceiptDeduction float64
	p.Db.QueryRow(`SELECT personal FROM deductions ORDER BY id DESC LIMIT 1`).Scan(&personalDeduction)
	p.Db.QueryRow(`SELECT max_kreceipt FROM deductions ORDER BY id DESC LIMIT 1`).Scan(&maxKReceiptDeduction)

	personalDeduction = 60000.0
	maxKReceiptDeduction = 50000.0

	for _, req := range reqs {
		var taxCSVResponseDetail tax.TaxCSVResponseDetail
		taxRequest := tax.TaxRequest{
			TotalIncome: req.TotalIncome,
			Wht:         req.Wht,
			Allowances: []tax.Allowance{
				{
					AllowanceType: "donation",
					Amount:        req.Donation,
				},
			},
		}
		taxResponse:= taxCalculator(taxRequest, personalDeduction, maxKReceiptDeduction)

		taxCSVResponseDetail.TotalIncome = req.TotalIncome

		if (taxResponse.Tax >= 0) && (taxResponse.TaxRefund == 0.0) {
			taxCSVResponseDetail.Tax = taxResponse.Tax
		} else {
			taxCSVResponseDetail.TaxRefund = taxResponse.TaxRefund
		}
		taxCSVResponse.Taxes = append(taxCSVResponse.Taxes, taxCSVResponseDetail)
	}

	return taxCSVResponse, nil
}

func taxCalculator(req tax.TaxRequest, personalDeduction, maxKReceiptDeduction float64) tax.TaxResponse {
	const maxDontationDecuction = 100000.0
	var taxResponse tax.TaxResponse
	income := req.TotalIncome - personalDeduction

	if len(req.Allowances) > 0 {
		for _, allowance := range req.Allowances {
			if allowance.AllowanceType == "donation" {
				if allowance.Amount > maxDontationDecuction {
					income -= maxDontationDecuction
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
