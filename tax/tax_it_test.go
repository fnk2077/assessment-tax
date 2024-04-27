//go:build integration

package tax

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"testing"

	_ "github.com/joho/godotenv/autoload"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

const serverPort = 8080

func TestITTaxCalculate(t *testing.T) {

	t.Run("given user able to calculate tax should return tax", func(t *testing.T) {
		reqBody := `{
		"totalIncome": 500000.0,
		"wht": 0.0,
		"allowances": [
		  {
			"allowanceType": "donation",
			"amount": 0.0
		  }
		]
	  }`

		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:%d/tax/calculations", serverPort), io.NopCloser(strings.NewReader(reqBody)))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		taxResponse := TaxResponse{
			Tax: 29000,
			TaxLevels: []TaxLevel{
				{
					Level: "0 - 150,000",
					Tax:   0,
				},
				{
					Level: "150,001 - 500,000",
					Tax:   29000,
				},
				{
					Level: "500,001 - 1,000,000",
					Tax:   0,
				},
				{
					Level: "1,000,001 - 2,000,000",
					Tax:   0,
				},
				{
					Level: "2,000,001 ขึ้นไป",
					Tax:   0,
				},
			},
		}

		var result TaxResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			t.Errorf("error decoding response body: %v", err)
		}

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, taxResponse.Tax, result.Tax)
		assert.Equal(t, taxResponse.TaxLevels, result.TaxLevels)
	})

	t.Run("given user able to calculate tax with csv should return taxes", func(t *testing.T) {
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("taxFile", "taxes.csv")
		if err != nil {
			t.Errorf("create form file error: %v", err)
		}
		part.Write([]byte("totalIncome,wht,allowances\n500000.0,0.0,0.0\n"))
		writer.Close()
		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:%d/tax/calculations/upload-csv", serverPort), body)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Content-Type", writer.FormDataContentType())

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		taxCSVResponse := TaxCSVResponse{
			Taxes: []TaxCSVResponseDetail{
				{
					TotalIncome: 500000.0,
					Tax:         29000.0,
				},
			},
		}

		var result TaxCSVResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			t.Errorf("error decoding response body: %v", err)
		}

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, taxCSVResponse.Taxes, result.Taxes)
	})

	// t.Run("given admin able to change personal deduction", func(t *testing.T) {
	// 	reqBody := `{"amount": 60000.0}`

	// 	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:%d/admin/deductions/personal", serverPort), strings.NewReader(reqBody))
	// 	if err != nil {
	// 		t.Fatal(err)
	// 	}

	// 	// fmt.Println("Heloooooooooooooooooooo", os.Getenv("ADMIN_USERNAME"), os.Getenv("ADMIN_PASSWORD"))
	// 	// fmt.Println(os.Environ())
	// 	req.SetBasicAuth( os.Getenv("ADMIN_USERNAME"), os.Getenv("ADMIN_PASSWORD"))
	// 	// req.SetBasicAuth("adminTax", "admin!")
	
	// 	req.Header.Set("Content-Type", "application/json")
	
	// 	client := &http.Client{}
	// 	resp, err := client.Do(req)
	// 	if err != nil {
	// 		t.Fatal(err)
	// 	}
	// 	defer resp.Body.Close()
	
	// 	if resp.StatusCode != http.StatusOK {
	// 		t.Fatalf("unexpected status code: %d", resp.StatusCode)
	// 	}
	
	// 	body, err := io.ReadAll(resp.Body)
	// 	if err != nil {
	// 		t.Fatal(err)
	// 	}
	
	// 	var result map[string]float64
	// 	err = json.Unmarshal(body, &result)
	// 	if err != nil {
	// 		t.Errorf("error decoding response body: %v", err)
	// 	}
	
	// 	fmt.Println("Decoded result:", result)
	// })
}
