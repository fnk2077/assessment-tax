package tax

import (
	"encoding/csv"
	"errors"

	"reflect"

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
// @Success 201 {object} TaxResponse "Returns the tax calculation"
// @Router /tax/calculations [post]
// @Failure 400 {object} Err "Bad Request"
// @Failure 500 {object} Err "Internal Server Error"
func (h *Handler) TaxCalculateHandler(c echo.Context) error {
	var req TaxRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: "Invalid request body"})
	}

	err := TaxRequestValidation(req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: err.Error()})
	}

	resp, err := h.store.TaxCalculate(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: "Internal server error"})
	}

	return c.JSON(http.StatusOK, resp)
}

// ChangeDeductionHandler changes deduction based on the provided data.
//
// @Summary Change deduction
// @Description Change deduction based on the provided data
// @Tags tax
// @Accept json
// @Produce json
// @Param type path string true "Type of deduction: personal or k-receipt"
// @Param amount body DeductionRequest true "Amount to be deducted"
// @Success 200 {object} map[string]float64 "Returns the updated deduction"
// @Router /admin/deductions/{type} [post]
// @Failure 400 {object} Err "Bad Request"
// @Failure 500 {object} Err "Internal Server Error"
func (h *Handler) ChangeDeductionHandler(c echo.Context) error {
	var deductionRequest DeductionRequest

	if err := c.Bind(&deductionRequest); err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: "Invalid request body"})
	}

	deductionType := c.Param("type")
	var response map[string]float64
	if deductionType == "personal" {
		response = map[string]float64{"personalDeduction": deductionRequest.Amount}

		if deductionRequest.Amount <= 10000 {
			return c.JSON(http.StatusBadRequest, Err{Message: "Amount must be more than 10,000"})
		}

		if deductionRequest.Amount > 100000 {
			return c.JSON(http.StatusBadRequest, Err{Message: "Amount must not exceed 100,000"})
		}
	} else if deductionType == "k-receipt" {
		response = map[string]float64{"kReceipt": deductionRequest.Amount}

		if deductionRequest.Amount <= 0 {
			return c.JSON(http.StatusBadRequest, Err{Message: "Amount must be more than 0"})
		}

		if deductionRequest.Amount > 100000 {
			return c.JSON(http.StatusBadRequest, Err{Message: "Amount must not exceed 100,000"})
		}
	} else {
		return c.JSON(http.StatusBadRequest, Err{Message: "Invalid deduction type"})
	}

	h.store.ChangeDeduction(deductionRequest.Amount, deductionType)

	return c.JSON(http.StatusOK, response)
}

// TaxCVSCalculateHandler calculates tax from CSV file.
//
// @Summary Calculate tax from CSV file
// @Description Calculate tax based on the data provided in a CSV file
// @Tags tax
// @Accept multipart/form-data
// @Param taxFile formData file true "CSV file containing tax data"
// @Success 200 {object} TaxCSVResponse "Returns the calculated tax"
// @Router /tax/calculations/upload-csv [post]
// @Failure 400 {object} Err "Bad Request"
// @Failure 500 {object} Err "Internal Server Error"
func (h *Handler) TaxCVSCalculateHandler(c echo.Context) error {
	var taxCSVRequests []TaxCSVRequest

	file, err := c.FormFile("taxFile")
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: "Invalid CSV file Key"})
	}

	src, err := file.Open()
	if err != nil || file.Filename != "taxes.csv" {
		return c.JSON(http.StatusBadRequest, Err{Message: "Invalid CSV file name or file not found"})
	}
	defer src.Close()

	reader := csv.NewReader(src)

	header, err := reader.Read()
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: "Invalid CSV file: missing header"})
	}

	expectedHeader := []string{"totalIncome", "wht", "donation"}
	if !reflect.DeepEqual(header, expectedHeader) {
		return c.JSON(http.StatusBadRequest, Err{Message: "Invalid CSV file: incorrect header format"})
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return c.JSON(http.StatusBadRequest, Err{Message: "Invalid CSV file"})
		}

		totalIncome, err := strconv.ParseFloat(record[0], 64)
		if err != nil {
			return err
		}
		if totalIncome < 0.0 {
			return c.JSON(http.StatusBadRequest, Err{Message: "total income must be more than 0"})
		}
		wht, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			return err
		}
		if wht < 0.0 {
			return c.JSON(http.StatusBadRequest, Err{Message: "wht must be more than 0"})
		}
		donation, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			return err
		}
		if donation < 0.0 {
			return c.JSON(http.StatusBadRequest, Err{Message: "donation amount must be equal or more than 0"})
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

func TaxRequestValidation(req TaxRequest) error {
	if req.TotalIncome < 0.0 {
		return errors.New("total income must be more than 0")
	}
	if req.Wht < 0.0 {
		return errors.New("wht must be more than 0")
	}

	for _, allowance := range req.Allowances {
		if allowance.Amount < 0.0 {
			return errors.New("allowance amount must be equal or more than 0")
		}
		if allowance.AllowanceType != "donation" && allowance.AllowanceType != "k-receipt" {
			return errors.New("invalid allowance type")
		}
	}

	return nil
}
