package calculator

import (
	"testing"

	"github.com/fnk2077/assessment-tax/tax"
	"github.com/stretchr/testify/assert"
)

func TestTaxCalculator(t *testing.T) {

	t.Run("Income 149,999.0 should return Tax 0.0", func(t *testing.T) {
		//Arrange
		want := 0.0
		req := tax.TaxRequest{
			TotalIncome: 149999.0,
		}

		//Act
		got := TaxCalculator(req, 60000.0, 50000.0)

		//Assert
		assert.Equal(t, want, got.Tax)
	})

	t.Run("Income 210,001.0 should return Tax 0.1", func(t *testing.T) {
		//Arrange
		want := 0.1
		req := tax.TaxRequest{
			TotalIncome: 210001.0,
		}

		//Act
		got := TaxCalculator(req, 60000.0, 50000.0)

		//Assert
		assert.Equal(t, want, got.Tax)
	})

	t.Run("Income 500,000 WTH 25,000 should return Tax 4000.0", func(t *testing.T) {
		//Arrange
		want := 4000.0
		req := tax.TaxRequest{
			TotalIncome: 500000.0,
			Wht:         25000.0,
		}
		//Act
		got := TaxCalculator(req, 60000.0, 50000.0)

		//Assert
		assert.Equal(t, want, got.Tax)
	})

	t.Run("Income 500,000.0 Donation 200,000.0 should return 19,000.0", func(t *testing.T) {
		//Arrange
		want := 19000.0
		req := tax.TaxRequest{
			TotalIncome: 500000.0,
			Allowances: []tax.Allowance{
				{
					AllowanceType: "donation",
					Amount:        200000.0,
				},
			},
		}

		//Act
		got := TaxCalculator(req, 60000.0, 50000.0)

		//Assert
		assert.Equal(t, want, got.Tax)
	})

	t.Run("Income 500,000.0 Donation 50,000.0 should return 24,000.0", func(t *testing.T) {
		//Arrange
		want := 24000.0
		req := tax.TaxRequest{
			TotalIncome: 500000.0,
			Allowances: []tax.Allowance{
				{
					AllowanceType: "donation",
					Amount:        50000.0,
				},
			},
		}

		//Act
		got := TaxCalculator(req, 60000.0, 50000.0)

		//Assert
		assert.Equal(t, want, got.Tax)
	})

	t.Run("Income 500,000.0 K-Receipt 100,000.0 should return 24,000.0", func(t *testing.T) {
		//Arrange
		want := 24000.0
		req := tax.TaxRequest{
			TotalIncome: 500000.0,
			Allowances: []tax.Allowance{
				{
					AllowanceType: "k-receipt",
					Amount:        100000.0,
				},
			},
		}

		//Act
		got := TaxCalculator(req, 60000.0, 50000.0)

		//Assert
		assert.Equal(t, want, got.Tax)
	})

	t.Run("Income 500,000.0 K-receipt 40,000.0 should return 25,000.0", func(t *testing.T) {
		//Arrange
		want := 25000.0
		req := tax.TaxRequest{
			TotalIncome: 500000.0,
			Allowances: []tax.Allowance{
				{
					AllowanceType: "k-receipt",
					Amount:        40000.0,
				},
			},
		}

		//Act
		got := TaxCalculator(req, 60000.0, 50000.0)

		//Assert
		assert.Equal(t, want, got.Tax)
	})

	t.Run("Income 400,000.0 WTH 25,000.0 Donation 100,000.0 K-receipt 40,000.0 should return 21,000.0", func(t *testing.T) {
		//Arrange
		want := 21000.0
		req := tax.TaxRequest{
			TotalIncome: 400000.0,
			Wht:         25000.0,
			Allowances: []tax.Allowance{
				{
					AllowanceType: "k-receipt",
					Amount:        50000.0,
				},
				{
					AllowanceType: "donation",
					Amount:        100000.0,
				},
			},
		}

		//Act
		got := TaxCalculator(req, 60000.0, 50000.0)

		//Assert
		assert.Equal(t, want, got.TaxRefund)
	})
}
