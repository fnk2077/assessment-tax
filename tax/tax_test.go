package tax

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

type StubTax struct {
	taxCalculate    TaxResponse
	changeDeduction error
	err             error
}

func (s *StubTax) TaxCalculate(TaxRequest) (TaxResponse, error) {
	return s.taxCalculate, nil
}

func (s *StubTax) ChangeDeduction(amount float64, deductionType string) error {
	return s.changeDeduction
}

func TestTaxCalculator(t *testing.T) {

	t.Run("Test tax calculator with total income 150000.0 (รายได้ 0 - 150,000 ได้รับการยกเว้น)", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/tax/calculations", io.NopCloser(strings.NewReader(
			`{
			"totalIncome": 150000.0,
			"wht": 0.0,
			"allowances": [
			  {
				"allowanceType": "donation",
				"amount": 0.0
			  }
			]
		  }`,
		)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		expected := TaxResponse{
			Tax: 0.0,
		}

		stubTax := StubTax{
			taxCalculate: expected,
		}

		handler := New(&stubTax)
		err := handler.TaxCalculateHandler(c)
		if err != nil {
			t.Errorf("expect nil but got %v", err)
		}
		actual := rec.Body.String()
		if rec.Code != http.StatusOK {
			t.Errorf("expect %d but got %d", http.StatusOK, rec.Code)
		}
		var got TaxResponse
		if err := json.Unmarshal([]byte(actual), &got); err != nil {
			t.Errorf("expect nil but got %v", err)
		}
		assert.Equal(t, expected, got)
	})
}
