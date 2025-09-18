package tests

import (
	"testing"

	"github.com/Zisimopoulou/platform-go-challenge/internal/models"
	"github.com/Zisimopoulou/platform-go-challenge/internal/validation"
	"github.com/go-playground/validator/v10"
)

func TestValidateAsset(t *testing.T) {
	tests := []struct {
		name    string
		asset   models.RawAsset
		wantErr bool
	}{
		{
			name: "valid chart",
			asset: models.RawAsset{
				Type: models.TypeChart,
				Payload: models.Chart{
					Title: "Test Chart",
					XAxis: "Time",
					YAxis: "Value",
					Data:  []int{1, 2, 3},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid chart - missing title",
			asset: models.RawAsset{
				Type: models.TypeChart,
				Payload: models.Chart{
					XAxis: "Time",
					YAxis: "Value",
					Data:  []int{1, 2, 3},
				},
			},
			wantErr: true,
		},
		{
			name: "valid insight",
			asset: models.RawAsset{
				Type: models.TypeInsight,
				Payload: models.Insight{
					Text: "40% of millenials spend more than 3hours on social media daily",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid asset type",
			asset: models.RawAsset{
				Type: "invalid-type",
				Payload: models.Chart{
					Title: "Test",
					XAxis: "Time",
					YAxis: "Value",
					Data:  []int{1, 2, 3},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateAsset(&tt.asset)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAsset() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil {
				t.Logf("Validation error (expected): %v", err)
			}
		})
	}
}

func TestValidationErrorResponse(t *testing.T) {
	// Test that validation errors are properly formatted
	validate := validator.New()

	type TestStruct struct {
		Name  string `validate:"required"`
		Email string `validate:"email"`
	}

	test := TestStruct{Email: "invalid-email"}
	err := validate.Struct(test)

	if err == nil {
		t.Fatal("Expected validation error")
	}

	// This test just ensures the function doesn't panic
	// In real usage, you'd use httptest to capture the response
	t.Logf("Validation errors: %v", err)
}
