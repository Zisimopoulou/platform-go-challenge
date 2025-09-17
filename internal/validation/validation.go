package validation

import (
	"encoding/json"
	"fmt"

	"github.com/Zisimopoulou/platform-go-challenge/internal/models"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}

 func ValidateAsset(asset *models.RawAsset) error {
 	if err := validate.Struct(asset); err != nil {
		return fmt.Errorf("asset validation failed: %v", err)
	}

 	return validatePayload(asset)
}

func validatePayload(asset *models.RawAsset) error {
	switch asset.Type {
	case models.TypeChart:
		var chart models.Chart
		if err := unmarshalPayload(asset.Payload, &chart); err != nil {
			return fmt.Errorf("invalid chart payload: %v", err)
		}
		return validate.Struct(chart)
		
	case models.TypeInsight:
		var insight models.Insight
		if err := unmarshalPayload(asset.Payload, &insight); err != nil {
			return fmt.Errorf("invalid insight payload: %v", err)
		}
		return validate.Struct(insight)
		
	case models.TypeAudience:
		var audience models.Audience
		if err := unmarshalPayload(asset.Payload, &audience); err != nil {
			return fmt.Errorf("invalid audience payload: %v", err)
		}
		return validate.Struct(audience)
		
	default:
		return fmt.Errorf("unknown asset type: %s", asset.Type)
	}
}

 func unmarshalPayload(payload interface{}, target interface{}) error {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, target)
}