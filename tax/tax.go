package tax

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

type Deduction struct {
	Personal    float64 `json:"personalDeduction"`
	MaxKReceipt float64 `json:"kReceipt"`
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
