package tax

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

var ErrNotFound = errors.New("not found")

type Handler struct {
	store Storer
}

type Storer interface {
	// PersonalDeduction() (error)
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
	return c.JSON(http.StatusOK, taxCalculator(req))
}

func taxCalculator(req TaxRequest) TaxResponse {
	income := req.TotalIncome
	income -= 60000

	taxRate := 0.0

    switch {
    case income <= 150000:
        taxRate = 0
    case income <= 500000:
        taxRate = 0.10
		income -= 150000
    case income <= 1000000:
        taxRate = 0.15
		income -= 500000
    case income <= 2000000:
		income -= 1000000
        taxRate = 0.20
    default:
		income -= 2000000
        taxRate = 0.35
    }
	tax := income * taxRate
	return TaxResponse{Tax: tax}
}