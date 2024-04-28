package tax

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

type StubTax struct {
	taxCalculate    TaxResponse
	taxCSVCalculate TaxCSVResponse
	changeDeduction error
	err             error
}

func (s *StubTax) TaxCalculate(TaxRequest) (TaxResponse, error) {
	return s.taxCalculate, s.err
}

func (s *StubTax) ChangeDeduction(amount float64, deductionType string) error {
	return s.changeDeduction
}

func (s *StubTax) TaxCSVCalculate([]TaxCSVRequest) (TaxCSVResponse, error) {
	return s.taxCSVCalculate, s.err
}
func TestTaxCalculate(t *testing.T) {

	t.Run("Income 150000.0 should return 0", func(t *testing.T) {
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

	t.Run("Income -150000.0 should return error", func(t *testing.T) {
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

	t.Run("WHT -25000.0 should return error", func(t *testing.T) {
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

	t.Run("Donation -25000.0 should return error", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/tax/calculations", io.NopCloser(strings.NewReader(
			`{
			"totalIncome": 150000.0,
			"wht": 25000.0,
			"allowances": [
			  {
				"allowanceType": "donation",
				"amount": -25000.0
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

		expectedErrorMessage := "allowance amount must be equal or more than 0"
		if errorMessage != expectedErrorMessage {
			t.Errorf("expected '%s' but got '%s'", expectedErrorMessage, errorMessage)
		}
	})

	t.Run("Wrong allowance type should return error", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/tax/calculations", io.NopCloser(strings.NewReader(
			`{
			"totalIncome": 150000.0,
			"wht": 25000.0,
			"allowances": [
			  {
				"allowanceType": "dota2 47 community",
				"amount": 25000.0
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

		expectedErrorMessage := "invalid allowance type"
		if errorMessage != expectedErrorMessage {
			t.Errorf("expected '%s' but got '%s'", expectedErrorMessage, errorMessage)
		}
	})

	t.Run("Money 150000.0 should return error(Invalid request body)", func(t *testing.T) {
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

	t.Run("Test tax calculate return InternalServerError", func(t *testing.T) {
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

func TestChangeDeduction(t *testing.T) {
	t.Run("Change Personal deduction amount 100,000.00 should return 100,000.00", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/", io.NopCloser(strings.NewReader(
			`{
				"amount": 100000.0
			  }`,
		)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/admin/deductions/:type")
		c.SetParamNames("type")
		c.SetParamValues("personal")

		stubTax := StubTax{
			changeDeduction: nil,
		}
		handler := New(&stubTax)
		handler.ChangeDeductionHandler(c)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status code %d but got %v", http.StatusOK, rec.Code)
		}

		var responseBody map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &responseBody); err != nil {
			t.Errorf("error decoding response body: %v", err)
		}

		message, ok := responseBody["personalDeduction"].(float64)
		if !ok {
			t.Error("expected 'personalDeduction' key in response body", responseBody)
		}

		expectedMessage := 100000.0
		if message != expectedMessage {
			t.Errorf("expected '%f' but got '%f'", expectedMessage, message)
		}
	})

	t.Run("Change Personal deduction amount 10,001.00 should return 10,001.0", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/", io.NopCloser(strings.NewReader(
			`{
				"amount": 10001.0
			  }`,
		)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/admin/deductions/:type")
		c.SetParamNames("type")
		c.SetParamValues("personal")

		stubTax := StubTax{
			changeDeduction: nil,
		}
		handler := New(&stubTax)
		handler.ChangeDeductionHandler(c)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status code %d but got %v", http.StatusOK, rec.Code)
		}

		var responseBody map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &responseBody); err != nil {
			t.Errorf("error decoding response body: %v", err)
		}

		message, ok := responseBody["personalDeduction"].(float64)
		if !ok {
			t.Error("expected 'personalDeduction' key in response body", responseBody)
		}

		expectedMessage := 10001.0
		if message != expectedMessage {
			t.Errorf("expected '%f' but got '%f'", expectedMessage, message)
		}
	})

	t.Run("Change Personal deduction amount 100,001.00 should return error (exceed 100,000.0)", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/", io.NopCloser(strings.NewReader(
			`{
				"amount": 100001.0
			  }`,
		)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/admin/deductions/:type")
		c.SetParamNames("type")
		c.SetParamValues("personal")

		stubError := StubTax{
			changeDeduction: echo.ErrBadRequest,
		}
		handler := New(&stubError)
		handler.ChangeDeductionHandler(c)

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

		expectedErrorMessage := "Amount must not exceed 100,000"
		if errorMessage != expectedErrorMessage {
			t.Errorf("expected '%s' but got '%s'", expectedErrorMessage, errorMessage)
		}
	})

	t.Run("Change Personal deduction amount 10,000 (must be more than 10.000)", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/", io.NopCloser(strings.NewReader(
			`{
				"amount": 10000.0
			  }`,
		)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/admin/deductions/:type")
		c.SetParamNames("type")
		c.SetParamValues("personal")

		stubError := StubTax{
			changeDeduction: echo.ErrBadRequest,
		}
		handler := New(&stubError)
		handler.ChangeDeductionHandler(c)

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

		expectedErrorMessage := "Amount must be more than 10,000"
		if errorMessage != expectedErrorMessage {
			t.Errorf("expected '%s' but got '%s'", expectedErrorMessage, errorMessage)
		}
	})
	t.Run("Change Max K-receipt deduction amount 1.00 should return 1.00", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/", io.NopCloser(strings.NewReader(
			`{
				"amount": 1.0
			  }`,
		)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/admin/deductions/:type")
		c.SetParamNames("type")
		c.SetParamValues("k-receipt")

		stubTax := StubTax{
			changeDeduction: nil,
		}
		handler := New(&stubTax)
		handler.ChangeDeductionHandler(c)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status code %d but got %v", http.StatusOK, rec.Code)
		}

		var responseBody map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &responseBody); err != nil {
			t.Errorf("error decoding response body: %v", err)
		}

		message, ok := responseBody["kReceipt"].(float64)
		if !ok {
			t.Error("expected 'kReceipt' key in response body", responseBody)
		}

		expectedMessage := 1.0
		if message != expectedMessage {
			t.Errorf("expected '%f' but got '%f'", expectedMessage, message)
		}
	})

	t.Run("Change Max K-receipt deduction amount 100,001.00 should return error (exceed 100,000.0)", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/", io.NopCloser(strings.NewReader(
			`{
				"amount": 100001.0
			  }`,
		)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/admin/deductions/:type")
		c.SetParamNames("type")
		c.SetParamValues("k-receipt")

		stubError := StubTax{
			changeDeduction: echo.ErrBadRequest,
		}
		handler := New(&stubError)
		handler.ChangeDeductionHandler(c)

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

		expectedErrorMessage := "Amount must not exceed 100,000"
		if errorMessage != expectedErrorMessage {
			t.Errorf("expected '%s' but got '%s'", expectedErrorMessage, errorMessage)
		}
	})

	t.Run("Change Max K-receipt deduction amount 0.00 should return error (must be more than 0.0)", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/", io.NopCloser(strings.NewReader(
			`{
				"amount": 0.0
			  }`,
		)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/admin/deductions/:type")
		c.SetParamNames("type")
		c.SetParamValues("k-receipt")

		stubError := StubTax{
			changeDeduction: echo.ErrBadRequest,
		}
		handler := New(&stubError)
		handler.ChangeDeductionHandler(c)

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

		expectedErrorMessage := "Amount must be more than 0"
		if errorMessage != expectedErrorMessage {
			t.Errorf("expected '%s' but got '%s'", expectedErrorMessage, errorMessage)
		}
	})

	t.Run("Invalid deduction request body", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/", io.NopCloser(strings.NewReader(
			`{
				"money": "asd",
			  }`,
		)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/admin/deductions/:type")
		c.SetParamNames("type")
		c.SetParamValues("k-receipt")

		stubError := StubTax{
			changeDeduction: echo.ErrBadRequest,
		}
		handler := New(&stubError)
		handler.ChangeDeductionHandler(c)

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

	t.Run("Invalid deduction type", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/", io.NopCloser(strings.NewReader(
			`{
				"amount": 100.0
			  }`,
		)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/admin/deductions/:type")
		c.SetParamNames("type")
		c.SetParamValues("test")

		stubError := StubTax{
			changeDeduction: echo.ErrBadRequest,
		}
		handler := New(&stubError)
		handler.ChangeDeductionHandler(c)

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

		expectedErrorMessage := "Invalid deduction type"
		if errorMessage != expectedErrorMessage {
			t.Errorf("expected '%s' but got '%s'", expectedErrorMessage, errorMessage)
		}
	})
}

func TestTaxCVSCalculate(t *testing.T) {

	t.Run("Test tax CSV calculate with total income 150000.0", func(t *testing.T) {
		e := echo.New()
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("taxFile", "taxes.csv")
		if err != nil {
			t.Errorf("create form file error: %v", err)
		}
		part.Write([]byte("totalIncome,wht,donation\n150000.0,0.0,0.0\n"))
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/tax/calculations/upload-csv", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		expected := TaxCSVResponse{
			Taxes: []TaxCSVResponseDetail{
				{
					TotalIncome: 150000.0,
					Tax:         0.0,
				},
			},
		}

		stubTax := StubTax{
			taxCSVCalculate: expected,
		}

		handler := New(&stubTax)
		err = handler.TaxCVSCalculateHandler(c)
		if err != nil {
			t.Errorf("expect nil but got %v", err)
		}
		actual := rec.Body.String()
		if rec.Code != http.StatusOK {
			t.Errorf("expect %d but got %d", http.StatusOK, rec.Code)
		}
		var got TaxCSVResponse
		if err := json.Unmarshal([]byte(actual), &got); err != nil {
			t.Errorf("expect nil but got %v", err)
		}
		assert.Equal(t, expected, got)
	})

	t.Run("Test tax CSV calculate with total income 150000.0 wht 5,000.0 expect TaxRefund", func(t *testing.T) {
		e := echo.New()
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("taxFile", "taxes.csv")
		if err != nil {
			t.Errorf("create form file error: %v", err)
		}
		part.Write([]byte("totalIncome,wht,donation\n150000.0,5000.0,0.0\n"))
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/tax/calculations/upload-csv", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		expected := TaxCSVResponse{
			Taxes: []TaxCSVResponseDetail{
				{
					TotalIncome: 150000.0,
					Tax:         0.0,
					TaxRefund:   5000.0,
				},
			},
		}

		stubTax := StubTax{
			taxCSVCalculate: expected,
		}

		handler := New(&stubTax)
		err = handler.TaxCVSCalculateHandler(c)
		if err != nil {
			t.Errorf("expect nil but got %v", err)
		}

		actual := rec.Body.String()
		if rec.Code != http.StatusOK {
			t.Errorf("expect %d but got %d", http.StatusOK, rec.Code)
		}
		var got TaxCSVResponse
		if err := json.Unmarshal([]byte(actual), &got); err != nil {
			t.Errorf("expect nil but got %v", err)
		}
	})

	t.Run("Test tax calculate csv with Wrong csv file Key", func(t *testing.T) {
		e := echo.New()
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("TaxTest", "taxes.csv")
		if err != nil {
			t.Errorf("create form file error: %v", err)
		}
		part.Write([]byte("totalIncome,wht,donation\n150000.0,0.0,0.0\n"))
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/tax/calculations/upload-csv", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		stubTax := StubTax{
			err: echo.ErrBadRequest,
		}

		handler := New(&stubTax)
		err = handler.TaxCVSCalculateHandler(c)

		if err != nil {
			t.Errorf("expect nil but got %v", err)
		}

		actual := rec.Body.String()
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expect %d but got %d", http.StatusBadRequest, rec.Code)
		}
		var got map[string]interface{}
		if err := json.Unmarshal([]byte(actual), &got); err != nil {
			t.Errorf("expect nil but got %v", err)
		}

		expected := map[string]interface{}{
			"message": "Invalid CSV file Key",
		}

		assert.Equal(t, expected, got)
	})

	t.Run("Test tax calculate csv with Wrong filename", func(t *testing.T) {
		e := echo.New()
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("taxFile", "taxesza.csv")
		if err != nil {
			t.Errorf("create form file error: %v", err)
		}
		part.Write([]byte("totalIncome,wht,donation\n150000.0,0.0,0.0\n"))
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/tax/calculations/upload-csv", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		stubTax := StubTax{
			err: echo.ErrBadRequest,
		}

		handler := New(&stubTax)
		err = handler.TaxCVSCalculateHandler(c)

		if err != nil {
			t.Errorf("expect nil but got %v", err)
		}

		actual := rec.Body.String()
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expect %d but got %d", http.StatusBadRequest, rec.Code)
		}
		var got map[string]interface{}
		if err := json.Unmarshal([]byte(actual), &got); err != nil {
			t.Errorf("expect nil but got %v", err)
		}

		expected := map[string]interface{}{
			"message": "Invalid CSV file name or file not found",
		}
		assert.Equal(t, expected, got)
	})

	t.Run("Test tax CSV calculate with Wrong TotalIncome value", func(t *testing.T) {
		e := echo.New()
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("taxFile", "taxes.csv")
		if err != nil {
			t.Errorf("create form file error: %v", err)
		}
		part.Write([]byte("totalIncome,wht,donation\n-123.00,0.0,0.0\n"))
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/tax/calculations/upload-csv", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		stubTax := StubTax{
			err: echo.ErrBadRequest,
		}

		handler := New(&stubTax)
		handler.TaxCVSCalculateHandler(c)

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

	t.Run("Test tax CSV calculate with Wrong WHT value", func(t *testing.T) {
		e := echo.New()
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("taxFile", "taxes.csv")
		if err != nil {
			t.Errorf("create form file error: %v", err)
		}
		part.Write([]byte("totalIncome,wht,donation\n123.00,-1230.0,0.0\n"))
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/tax/calculations/upload-csv", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		stubTax := StubTax{
			err: echo.ErrBadRequest,
		}

		handler := New(&stubTax)
		handler.TaxCVSCalculateHandler(c)

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

	t.Run("Test tax CSV calculate with Wrong Donation value", func(t *testing.T) {
		e := echo.New()
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("taxFile", "taxes.csv")
		if err != nil {
			t.Errorf("create form file error: %v", err)
		}
		part.Write([]byte("totalIncome,wht,donation\n123.00,0.0,-1.0\n"))
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/tax/calculations/upload-csv", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		stubTax := StubTax{
			err: echo.ErrBadRequest,
		}

		handler := New(&stubTax)
		handler.TaxCVSCalculateHandler(c)

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

		expectedErrorMessage := "donation amount must be equal or more than 0"
		if errorMessage != expectedErrorMessage {
			t.Errorf("expected '%s' but got '%s'", expectedErrorMessage, errorMessage)
		}

	})

	t.Run("Test tax CSV calculate with Wrong Type value", func(t *testing.T) {
		e := echo.New()
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("taxFile", "taxes.csv")
		if err != nil {
			t.Errorf("create form file error: %v", err)
		}
		part.Write([]byte("totalIncome,wht,donation\nabc,def,gdf.ads\n"))
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/tax/calculations/upload-csv", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		stubTaxError := StubTax{
			err: echo.ErrBadRequest,
		}

		handler := New(&stubTaxError)
		if err := handler.TaxCVSCalculateHandler(c); err == nil {
			t.Errorf("expected error but got nil")
		}

	})

	t.Run("Test tax CSV calculate with Wrong Type WHT", func(t *testing.T) {
		e := echo.New()
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("taxFile", "taxes.csv")
		if err != nil {
			t.Errorf("create form file error: %v", err)
		}
		part.Write([]byte("totalIncome,wht,donation\n500000.0,def,gdf\n"))
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/tax/calculations/upload-csv", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		stubTaxError := StubTax{
			err: echo.ErrBadRequest,
		}

		handler := New(&stubTaxError)
		if err := handler.TaxCVSCalculateHandler(c); err == nil {
			t.Errorf("expected error but got nil")
		}

	})

	t.Run("Test tax CSV calculate with Wrong Type Allowance", func(t *testing.T) {
		e := echo.New()
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("taxFile", "taxes.csv")
		if err != nil {
			t.Errorf("create form file error: %v", err)
		}
		part.Write([]byte("totalIncome,wht,donation\n500000.0,25000.0,gdf\n"))
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/tax/calculations/upload-csv", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		stubTaxError := StubTax{
			err: echo.ErrBadRequest,
		}

		handler := New(&stubTaxError)
		if err := handler.TaxCVSCalculateHandler(c); err == nil {
			t.Errorf("expected error but got nil")
		}

	})

	t.Run("Test tax CSV calculate with Wrong Header", func(t *testing.T) {
		e := echo.New()
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("taxFile", "taxes.csv")
		if err != nil {
			t.Errorf("create form file error: %v", err)
		}
		part.Write([]byte("asdsadsad,wht,allowances\n123.00,0.0,-1.0\n"))
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/tax/calculations/upload-csv", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		stubTax := StubTax{
			err: echo.ErrBadRequest,
		}

		handler := New(&stubTax)
		handler.TaxCVSCalculateHandler(c)

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

		expectedErrorMessage := "Invalid CSV file: incorrect header format"
		if errorMessage != expectedErrorMessage {
			t.Errorf("expected '%s' but got '%s'", expectedErrorMessage, errorMessage)
		}

	})

	t.Run("Test tax CSV calculate with Missing header", func(t *testing.T) {
		e := echo.New()
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("taxFile", "taxes.csv")
		if err != nil {
			t.Errorf("create form file error: %v", err)
		}
		part.Write([]byte(""))
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/tax/calculations/upload-csv", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		stubTax := StubTax{
			err: echo.ErrBadRequest,
		}

		handler := New(&stubTax)
		handler.TaxCVSCalculateHandler(c)

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

		expectedErrorMessage := "Invalid CSV file: missing header"
		if errorMessage != expectedErrorMessage {
			t.Errorf("expected '%s' but got '%s'", expectedErrorMessage, errorMessage)
		}

	})

	t.Run("Test tax CSV calculate with Invalid csv file", func(t *testing.T) {
		e := echo.New()
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("taxFile", "taxes.csv")
		if err != nil {
			t.Errorf("create form file error: %v", err)
		}
		part.Write([]byte("totalIncome,wht,donation\n500000.0,25000.0,0.0,123,456\n"))
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/tax/calculations/upload-csv", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		stubTax := StubTax{
			err: echo.ErrBadRequest,
		}

		handler := New(&stubTax)
		handler.TaxCVSCalculateHandler(c)

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

		expectedErrorMessage := "Invalid CSV file"
		if errorMessage != expectedErrorMessage {
			t.Errorf("expected '%s' but got '%s'", expectedErrorMessage, errorMessage)
		}

	})

	t.Run("Test tax CSV calculate return InternalServerError", func(t *testing.T) {
		e := echo.New()
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("taxFile", "taxes.csv")
		if err != nil {
			t.Errorf("create form file error: %v", err)
		}
		part.Write([]byte("totalIncome,wht,donation\n500000.0,25000.0,0.0\n"))
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/tax/calculations/upload-csv", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		stubError := StubTax{err: echo.ErrInternalServerError}
		handler := New(&stubError)
		handler.TaxCVSCalculateHandler(c)

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
