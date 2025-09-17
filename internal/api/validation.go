 package api

import (
	"fmt"
	"net/http"

	"github.com/Zisimopoulou/platform-go-challenge/internal/models"
	"github.com/Zisimopoulou/platform-go-challenge/internal/validation"   
	"github.com/go-playground/validator/v10"
)

 func validateStruct(s interface{}) error {
	return validation.ValidateStruct(s)
}

 func validateAsset(asset *models.RawAsset) error {
	return validation.ValidateAsset(asset)
}

 func validationErrorResponse(w http.ResponseWriter, err error) {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		errors := make(map[string]string)
		for _, fieldError := range validationErrors {
			errors[fieldError.Field()] = fmt.Sprintf(
				"failed validation for '%s' with value '%v'",
				fieldError.Tag(),
				fieldError.Value(),
			)
		}
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":   "validation failed",
			"details": errors,
		})
	} else {
		writeError(w, http.StatusBadRequest, err.Error())
	}
}
