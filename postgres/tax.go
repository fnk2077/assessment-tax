package postgres

import (
	"github.com/fnk2077/assessment-tax/tax"
)

type Allowance struct {
	AllowanceType string  `json:"allowanceType"`
	Amount        float64 `json:"amount"`
}

type TaxLevel struct {
	Level string `json:"level"`
	Tax   float64 `json:"tax"`
}

type TaxRequest struct {
	TotalIncome float64     `json:"totalIncome"`
	Wht         float64     `json:"wht"`
	Allowances  []Allowance `json:"allowances"`
}

type TaxResponse struct {
	Tax float64 `json:"tax"`
	TaxLevels []TaxLevel `json:"taxLevel"`
}

type Deduction struct {
	Personal float64 `json:"personalDeduction"`
	KReceipt float64 `json:"kReceipt"`
}

func (p *Postgres) PersonalDeduction() tax.Deduction {
    var deduction tax.Deduction
    query := `SELECT personal_deduction FROM deductions ORDER BY id DESC LIMIT 1`
    p.Db.QueryRow(query).Scan(&deduction.Personal)
    return deduction
}

func (p *Postgres) ChangePersonalDeduction(deduction float64) {
	query := `INSERT INTO deductions (personal_deduction) VALUES ($1)`
	p.Db.Exec(query, deduction)
}

func (p *Postgres) ChangeKReceiptDeduction(deduction float64) {
	query := `INSERT INTO deductions (personal_deduction) VALUES ($1)`
	p.Db.Exec(query, deduction)
}