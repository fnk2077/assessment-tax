package tax

import (
	"testing"
	"github.com/stretchr/testify/assert"
)
func TestTaxCalculator(t *testing.T) {

	t.Run("Test tax calculator with total income 149000 (รายได้ 0 - 150,000 ได้รับการยกเว้น)", func(t *testing.T){
		//Arrange
		want := 0.0
		req := TaxRequest{
			TotalIncome: 149000.0,
		}

		//Act
		got := taxCalculator(req)

		//Assert
		assert.Equal(t, want, got.Tax)
	})

	t.Run("Test tax calculator with total income 210001 (150,001 - 500,000 อัตราภาษี 10%)", func(t *testing.T){
		//Arrange
		want := 0.1
		req := TaxRequest{
			TotalIncome: 210001.0,
		}

		//Act
		got := taxCalculator(req)

		//Assert
		assert.Equal(t, want, got.Tax)
	})
}