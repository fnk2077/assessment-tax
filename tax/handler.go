package tax

import (
	"encoding/csv"
	"errors"

	"io"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

var ErrNotFound = errors.New("not found")

type Handler struct {
	store Storer
}

type Storer interface {
	TaxCalculate(TaxRequest) (TaxResponse, error)
	TaxCSVCalculate([]TaxCSVRequest) (TaxCSVResponse, error)
	ChangeDeduction(float64, string) error
}

func New(db Storer) *Handler {
	return &Handler{store: db}
}

type Err struct {
	Message string `json:"message"`
}

// TaxCalculate from tax request.
//
// @Summary Calculate tax from request
// @Description Calculate tax from request based on the provided data
// @Tags tax
// @Accept json
// @Produce json
// @Param request body TaxRequest true "Tax data"
// @Success 201 {object} TaxRequest
// @Router /tax/calculations [post]
// @Failure 400 {object} Err
// @Failure 500 {object} Err
func (h *Handler) TaxCalculateHandler(c echo.Context) error {
	var req TaxRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: "Invalid request body"})
	}

	err := validateInput(req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: err.Error()})
	}

	resp, err := h.store.TaxCalculate(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: "Internal server error"})
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) ChangeDeductionHandler(c echo.Context) error {
	req := struct {
		Amount float64 `json:"amount"`
	}{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: "Invalid request body"})
	}

	deductionType := c.Param("type")
	var response map[string]float64
	if deductionType == "personal" {
		response = map[string]float64{"personalDeduction": req.Amount}

		if req.Amount <= 10000 {
			return c.JSON(http.StatusBadRequest, Err{Message: "Amount must be more than 10,000"})
		}

		if req.Amount > 100000 {
			return c.JSON(http.StatusBadRequest, Err{Message: "Amount must not exceed 100,000"})
		}
	} else if deductionType == "k-receipt" {
		response = map[string]float64{"kReceipt": req.Amount}

		if req.Amount <= 0 {
			return c.JSON(http.StatusBadRequest, Err{Message: "Amount must be more than 0"})
		}

		if req.Amount > 100000 {
			return c.JSON(http.StatusBadRequest, Err{Message: "Amount must not exceed 100,000"})
		}
	} else {
		return c.JSON(http.StatusBadRequest, Err{Message: "Invalid deduction type"})
	}

	h.store.ChangeDeduction(req.Amount, deductionType)

	return c.JSON(http.StatusOK, response)
}

func (h *Handler) TaxCVSCalculateHandler(c echo.Context) error {
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

	taxCSVResponse, err := h.store.TaxCSVCalculate(taxCSVRequests)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: "Internal server error"})
	}

	return c.JSON(http.StatusOK, taxCSVResponse)
}

func validateInput(req TaxRequest) error {
	if req.TotalIncome < 0.0 {
		return errors.New("total income must be more than 0")
	}
	if req.Wht < 0.0 {
		return errors.New("wht must be more than 0")
	}
	return nil
}
