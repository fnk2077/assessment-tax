package middleware

import (
	"os"
	"testing"
)

func TestAuthMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		username       string
		password       string
		expectedResult bool
		expectedError  error
	}{
		{
			name:           "Valid credentials",
			username:       "admin",
			password:       "password",
			expectedResult: true,
			expectedError:  nil,
		},
		{
			name:           "Invalid credentials",
			username:       "invalid_user",
			password:       "invalid_password",
			expectedResult: false,
			expectedError:  nil,
		},
		{
			name:           "Empty username",
			username:       "",
			password:       "password",
			expectedResult: false,
			expectedError:  nil,
		},
		{
			name:           "Empty password",
			username:       "admin",
			password:       "",
			expectedResult: false,
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("ADMIN_USERNAME", "admin")
			os.Setenv("ADMIN_PASSWORD", "password")

			result, err := AuthMiddleware(tt.username, tt.password, nil)

			if result != tt.expectedResult {
				t.Errorf("Expected result %v but got %v", tt.expectedResult, result)
			}

			if err != tt.expectedError {
				t.Errorf("Expected error %v but got %v", tt.expectedError, err)
			}
		})
	}
}
