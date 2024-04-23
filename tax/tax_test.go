package tax

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTaxCalculator(t *testing.T) {

	t.Run("Test tax calculator with total income 149000 (รายได้ 0 - 150,000 ได้รับการยกเว้น)", func(t *testing.T) {
		//Arrange
		want := 0.0
		req := TaxRequest{
			TotalIncome: 149000.0,
		} 

		deduction := Deduction{
			Personal: 60000.0,
		}

		//Act
		got := taxCalculator(req , deduction).Tax

		//Assert
		assert.Equal(t, want, got)
	})

	t.Run("Test tax calculator with total income 210001 (150,001 - 500,000 อัตราภาษี 10%)", func(t *testing.T) {
		//Arrange
		want := 0.1
		req := TaxRequest{
			TotalIncome: 210001.0,
		}

		deduction := Deduction{
			Personal: 60000.0,
		}

		//Act
		got := taxCalculator(req, deduction).Tax

		//Assert
		assert.Equal(t, want, got)
	})

	t.Run("Test tax calculator with total income 500,000 wth 25,000 (150,001 - 500,000 อัตราภาษี 10%)", func(t *testing.T){
		//Arrange
		want := 4000.0
		req := TaxRequest{
			TotalIncome: 500000.0,
			Wht: 25000.0,
		}
		deduction := Deduction{
			Personal: 60000.0,
		}
		//Act
		got := taxCalculator(req, deduction).Tax

		//Assert
		assert.Equal(t, want, got)
	})

	t.Run("Test tax calculator with total income 500,000 donation amount 200,000 (150,001 - 500,000 อัตราภาษี 10%)", func(t *testing.T){
		//Arrange
		want := 19000.0
		req := TaxRequest{
			TotalIncome: 500000.0,
			Allowances: []Allowance{
				{
					AllowanceType: "donation",
					Amount: 200000.0,
				},
			},
		}

		deduction := Deduction{
			Personal: 60000.0,
		}

		//Act
		got := taxCalculator(req, deduction).Tax

		//Assert
		assert.Equal(t, want, got)
	})

	t.Run("Test tax calculator with total income 500,000 donation amount 50,000 (150,001 - 500,000 อัตราภาษี 10%)", func(t *testing.T){
		//Arrange
		want := 24000.0
		req := TaxRequest{
			TotalIncome: 500000.0,
			Allowances: []Allowance{
				{
					AllowanceType: "donation",
					Amount: 50000.0,
				},
			},
		}

		deduction := Deduction{
			Personal: 60000.0,
		}

		//Act
		got := taxCalculator(req, deduction).Tax

		//Assert
		assert.Equal(t, want, got)
	})


	t.Run("Test tax calculator with total income 500,000 k-receipt amount 100,000(max 50,000) (150,001 - 500,000 อัตราภาษี 10%)", func(t *testing.T){
		//Arrange
		want := 24000.0
		req := TaxRequest{
			TotalIncome: 500000.0,
			Allowances: []Allowance{
				{
					AllowanceType: "k-receipt",
					Amount: 100000.0,
				},
			},
		}

		deduction := Deduction{
			Personal: 60000.0,
			MaxKReceipt: 50000.0,
		}

		//Act
		got := taxCalculator(req, deduction).Tax

		//Assert
		assert.Equal(t, want, got)
	})

	t.Run("Test tax calculator with total income 500,000 k-receipt amount 40,000(max 50,000) (150,001 - 500,000 อัตราภาษี 10%)", func(t *testing.T){
		//Arrange
		want := 25000.0
		req := TaxRequest{
			TotalIncome: 500000.0,
			Allowances: []Allowance{
				{
					AllowanceType: "k-receipt",
					Amount: 40000.0,
				},
			},
		}

		deduction := Deduction{
			Personal: 60000.0,
			MaxKReceipt: 50000.0,
		}

		//Act
		got := taxCalculator(req, deduction).Tax

		//Assert
		assert.Equal(t, want, got)
	})

	t.Run("Test tax calculator with total income 500,000 k-receipt wth 25,000 amount 40,000(max 50,000) (150,001 - 500,000 อัตราภาษี 10%)", func(t *testing.T){
		//Arrange
		want := 21000.0
		req := TaxRequest{
			TotalIncome: 400000.0,
			Wht: 25000.0,
			Allowances: []Allowance{
				{
					AllowanceType: "k-receipt",
					Amount: 50000.0,
				},
				{
					AllowanceType: "donation",
					Amount: 100000.0,
				},
			},
		}

		deduction := Deduction{
			Personal: 60000.0,
			MaxKReceipt: 50000.0,
		}

		//Act
		got := taxCalculator(req, deduction).TaxRefund

		//Assert
		assert.Equal(t, want, got)
	})
}
