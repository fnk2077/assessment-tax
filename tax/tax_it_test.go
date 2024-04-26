//go:build integration

package tax

import (
	"encoding/json"
	"fmt"

	"net/http"
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

		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:%d/tax/calculations", serverPort), strings.NewReader(reqBody))
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
}
