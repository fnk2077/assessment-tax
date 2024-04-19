package tax

import (
	"errors"
	"math"
	"net/http"

	"github.com/labstack/echo/v4"
)

var ErrNotFound = errors.New("not found")

type Handler struct {
	store Storer
}

type Storer interface {
	PersonalDeduction() (Deduction)
	ChangePersonalDeduction(float64)
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
	deductions := h.store.PersonalDeduction()
	return c.JSON(http.StatusOK, taxCalculator(req, deductions))
}

func (h *Handler) ChangePersonalDeduction(c echo.Context) error {
    req := struct {
		Amount float64 `json:"amount"`
	}{}
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, Err{Message: "Invalid request body"})
    }
    h.store.ChangePersonalDeduction(req.Amount)
    response := map[string]float64{"personalDeduction": req.Amount}
    return c.JSON(http.StatusOK, response)
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

	for _, bracket := range taxLevels {
		if income > bracket.min && income <= bracket.max {
			taxResponse.Tax = ((income - bracket.min) * bracket.rate) - req.Wht
			taxResponse.TaxLevels = append(taxResponse.TaxLevels, TaxLevel{
				Level: bracket.level,
				Tax: ((income - bracket.min) * bracket.rate) - req.Wht,
			})
		} else {
			taxResponse.TaxLevels = append(taxResponse.TaxLevels, TaxLevel{
				Level: bracket.level,
				Tax: 0.0,
			})
		}
	}

	return taxResponse
}