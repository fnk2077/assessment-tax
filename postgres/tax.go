package postgres

import (
	"github.com/fnk2077/assessment-tax/pkg/calculator"
	"github.com/fnk2077/assessment-tax/tax"
)

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
	err := p.Db.QueryRow(`SELECT personal FROM deductions ORDER BY id DESC LIMIT 1`).Scan(&personalDeduction)
	if err != nil {
		return tax.TaxResponse{}, err
	}

	err = p.Db.QueryRow(`SELECT max_kreceipt FROM deductions ORDER BY id DESC LIMIT 1`).Scan(&maxKReceiptDeduction)
	if err != nil {
		return tax.TaxResponse{}, err
	}

	taxResponse := calculator.TaxCalculator(req, personalDeduction, maxKReceiptDeduction)

	return taxResponse, nil
}

func (p *Postgres) TaxCSVCalculate(reqs []tax.TaxCSVRequest) (tax.TaxCSVResponse, error) {
	var taxCSVResponse tax.TaxCSVResponse
	var personalDeduction float64
	var maxKReceiptDeduction float64
	err := p.Db.QueryRow(`SELECT personal FROM deductions ORDER BY id DESC LIMIT 1`).Scan(&personalDeduction)
	if err != nil {
		return tax.TaxCSVResponse{}, err
	}

	err = p.Db.QueryRow(`SELECT max_kreceipt FROM deductions ORDER BY id DESC LIMIT 1`).Scan(&maxKReceiptDeduction)
	if err != nil {
		return tax.TaxCSVResponse{}, err
	}

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
		taxResponse := calculator.TaxCalculator(taxRequest, personalDeduction, maxKReceiptDeduction)

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


