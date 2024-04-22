package tax

import (
	"encoding/csv"
	"errors"
	"io"
	"math"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

var ErrNotFound = errors.New("not found")

type Handler struct {
	store Storer
}

type Storer interface {
	GetDefaultDeduction() Deduction
	ChangePersonalDeduction(float64)
	ChangeKReceiptDeduction(float64)
}

func New(db Storer) *Handler {
	return &Handler{store: db}
}

type Err struct {
	Message string `json:"message"`
}

// TaxCalculate calculate tax from request.
//
// @Summary Calculate tax from request
// @Description Calculate tax from request based on the provided data
// @Tags tax
// @Accept json
// @Produce json
// @Param request body TaxRequest true "Tax data"
// @Success 201 {object} Tax
// @Router /tax/calculations [post]
// @Failure 400 {object} Err
// @Failure 500 {object} Err
func (h *Handler) TaxCalculate(c echo.Context) error {
	var req TaxRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: "Invalid request body"})
	}
	deductions := h.store.GetDefaultDeduction()
	return c.JSON(http.StatusOK, taxCalculator(req, deductions))
}

func (h *Handler) ChangePersonalDeduction(c echo.Context) error {
	req := struct {
		Amount float64 `json:"amount"`
	}{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: "Invalid request body"})
	}

	if req.Amount <= 10000 {
        return c.JSON(http.StatusBadRequest, Err{Message: "Amount must be more than 10,000"})
    }
	
	if req.Amount > 100000 {
		return c.JSON(http.StatusBadRequest, Err{Message: "Amount must not exceed 100,000"})
	}

	h.store.ChangePersonalDeduction(req.Amount)
	response := map[string]float64{"personalDeduction": req.Amount}
	return c.JSON(http.StatusOK, response)
}

func (h *Handler) ChangeKReciept(c echo.Context) error {
	req := struct {
		Amount float64 `json:"amount"`
	}{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: "Invalid request body"})
	}

	if req.Amount <= 0 {
        return c.JSON(http.StatusBadRequest, Err{Message: "Amount must be more than 0"})
    }

	if req.Amount > 100000 {
		return c.JSON(http.StatusBadRequest, Err{Message: "Amount must not exceed 100,000"})
	}


	h.store.ChangeKReceiptDeduction(req.Amount)
	response := map[string]float64{"kReceipt": req.Amount}
	return c.JSON(http.StatusOK, response)

}

func (h *Handler) ReadTaxCSV(c echo.Context) error {
	var taxCSVRequests []TaxCSVRequest

	file, err := c.FormFile("taxFile")
	if err != nil {
		return err
	}

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	reader := csv.NewReader(src)
	_, _ = reader.Read()

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		totalIncome, err := strconv.ParseFloat(record[0], 64)
		if err != nil {
			return err
		}
		wht, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			return err
		}
		donation, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			return err
		}

		taxCSVRequests = append(taxCSVRequests, TaxCSVRequest{
			TotalIncome: totalIncome,
			Wht:         wht,
			Donation:    donation,
		})
	}

	var taxCSVResponse TaxCSVResponse

	for _, taxCSVRequest := range taxCSVRequests {
		var taxCSVResponseDetail TaxCSVResponseDetail
		req := TaxRequest{
			TotalIncome: taxCSVRequest.TotalIncome,
			Wht:         taxCSVRequest.Wht,
			Allowances: []Allowance{
				{
					AllowanceType: "donation",
					Amount:        taxCSVRequest.Donation,
				},
			},
		}
		deductions := h.store.GetDefaultDeduction() //////////////
		taxResponse := taxCalculator(req, deductions)
		taxCSVResponseDetail.TotalIncome = taxCSVRequest.TotalIncome

		if (taxResponse.Tax >= 0) && (taxResponse.TaxRefund == 0.0) {
			taxCSVResponseDetail.Tax = taxResponse.Tax
		} else {
			taxCSVResponseDetail.TaxRefund = taxResponse.TaxRefund
		}
		taxCSVResponse.Taxes = append(taxCSVResponse.Taxes, taxCSVResponseDetail)
	}

	return c.JSON(http.StatusOK, taxCSVResponse)
}

func taxCalculator(req TaxRequest, deduction Deduction) TaxResponse {
	var taxResponse TaxResponse
	income := req.TotalIncome
	income -= deduction.Personal

	if len(req.Allowances) > 0 {
		for _, allowance := range req.Allowances {
			if allowance.AllowanceType == "donation" {
				if allowance.Amount > 100000.0 {
					income -= 100000.0
				} else {
					income -= allowance.Amount
				}
			}
			if allowance.AllowanceType == "k-receipt" {
				if allowance.Amount > deduction.MaxKReceipt {
					income -= deduction.MaxKReceipt
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

			taxResponse.TaxLevels = append(taxResponse.TaxLevels, TaxLevel{
				Level: bracket.level,
				Tax:   ((income - bracket.min) * bracket.rate),
			})
			totalTax += ((income - bracket.min) * bracket.rate)

		} else if income <= bracket.min {
			taxResponse.TaxLevels = append(taxResponse.TaxLevels, TaxLevel{
				Level: bracket.level,
				Tax:   0.0,
			})
		} else {
			taxResponse.TaxLevels = append(taxResponse.TaxLevels, TaxLevel{
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
