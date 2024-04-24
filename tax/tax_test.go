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
	return s.taxCalculate, s.err
}

func (s *StubTax) ChangeDeduction(amount float64, deductionType string) error {
	return s.changeDeduction
}

func TestTaxCalculate(t *testing.T) {

	t.Run("Test tax calculate with total income 150000.0 (รายได้ 0 - 150,000 ได้รับการยกเว้น)", func(t *testing.T) {
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

	t.Run("Test tax calculate with total income -150000.0 should return error", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/tax/calculations", io.NopCloser(strings.NewReader(
			`{
			"totalIncome": -150000.0,
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

		stubError := StubTax{err: echo.ErrBadRequest}
		handler := New(&stubError)
		handler.TaxCalculateHandler(c)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status code %d but got %v", http.StatusBadRequest, rec.Code)
		}

		var responseBody map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &responseBody); err != nil {
			t.Errorf("error decoding response body: %v", err)
		}

		errorMessage, ok := responseBody["message"].(string)
		if !ok {
			t.Error("expected 'message' key in response body")
		}

		expectedErrorMessage := "total income must be more than 0"
		if errorMessage != expectedErrorMessage {
			t.Errorf("expected '%s' but got '%s'", expectedErrorMessage, errorMessage)
		}
	})

	t.Run("Test tax calculate with total wht -25000.0 should return error", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/tax/calculations", io.NopCloser(strings.NewReader(
			`{
			"totalIncome": 150000.0,
			"wht": -25000.0,
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

		stubError := StubTax{err: echo.ErrBadRequest}
		handler := New(&stubError)
		handler.TaxCalculateHandler(c)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status code %d but got %v", http.StatusBadRequest, rec.Code)
		}

		var responseBody map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &responseBody); err != nil {
			t.Errorf("error decoding response body: %v", err)
		}

		errorMessage, ok := responseBody["message"].(string)
		if !ok {
			t.Error("expected 'message' key in response body")
		}

		expectedErrorMessage := "wht must be more than 0"
		if errorMessage != expectedErrorMessage {
			t.Errorf("expected '%s' but got '%s'", expectedErrorMessage, errorMessage)
		}
	})

	t.Run("Test tax calculate with total money 150000.0 should return error", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/tax/calculations", io.NopCloser(strings.NewReader(
			`{
			"Money": 150000.0,
		  }`,
		)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		stubError := StubTax{err: echo.ErrBadRequest}
		handler := New(&stubError)
		handler.TaxCalculateHandler(c)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status code %d but got %v", http.StatusBadRequest, rec.Code)
		}

		var responseBody map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &responseBody); err != nil {
			t.Errorf("error decoding response body: %v", err)
		}

		errorMessage, ok := responseBody["message"].(string)
		if !ok {
			t.Error("expected 'message' key in response body")
		}

		expectedErrorMessage := "Invalid request body"
		if errorMessage != expectedErrorMessage {
			t.Errorf("expected '%s' but got '%s'", expectedErrorMessage, errorMessage)
		}
	})

	t.Run("Test tax calculate with total income 150000.0 should return error", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/tax/calculations", io.NopCloser(strings.NewReader(
			`{
			"totalIncome": 150000.0,
			"wht": 25000.0,
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

		stubError := StubTax{err: echo.ErrInternalServerError}
		handler := New(&stubError)
		handler.TaxCalculateHandler(c)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status code %d but got %v", http.StatusInternalServerError, rec.Code)
		}

		var responseBody map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &responseBody); err != nil {
			t.Errorf("error decoding response body: %v", err)
		}

		errorMessage, ok := responseBody["message"].(string)
		if !ok {
			t.Error("expected 'message' key in response body")
		}

		expectedErrorMessage := "Internal server error"
		if errorMessage != expectedErrorMessage {
			t.Errorf("expected '%s' but got '%s'", expectedErrorMessage, errorMessage)
		}
	})
}
