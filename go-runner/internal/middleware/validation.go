package middleware

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	
	// Register custom validators
	validate.RegisterValidation("notempty", notEmpty)
	validate.RegisterValidation("port", isValidPort)
	validate.RegisterValidation("hexcolor", isHexColor)
}

// ValidationMiddleware validates request body
func ValidationMiddleware(model interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		var requestBody interface{}
		
		// Parse JSON body
		if err := c.ShouldBindJSON(&requestBody); err != nil {
			HandleError(c, NewError(http.StatusBadRequest, "Invalid JSON format", err.Error()))
			return
		}

		// Convert to target model
		jsonData, _ := json.Marshal(requestBody)
		if err := json.Unmarshal(jsonData, model); err != nil {
			HandleError(c, NewError(http.StatusBadRequest, "Invalid request format", err.Error()))
			return
		}

		// Validate the model
		if err := validate.Struct(model); err != nil {
			var validationErrors []ValidationError
			
			for _, err := range err.(validator.ValidationErrors) {
				validationErrors = append(validationErrors, ValidationError{
					Field:   err.Field(),
					Tag:     err.Tag(),
					Value:   err.Value().(string),
					Message: getValidationMessage(err),
				})
			}
			
			HandleValidationError(c, validationErrors)
			return
		}

		// Store validated model in context
		c.Set("validated_model", model)
		c.Next()
	}
}

// Custom validation functions
func notEmpty(fl validator.FieldLevel) bool {
	value := fl.Field()
	switch value.Kind() {
	case reflect.String:
		return strings.TrimSpace(value.String()) != ""
	case reflect.Slice, reflect.Array:
		return value.Len() > 0
	case reflect.Ptr:
		return !value.IsNil()
	default:
		return true
	}
}

func isValidPort(fl validator.FieldLevel) bool {
	port := fl.Field().Int()
	return port > 0 && port <= 65535
}

func isHexColor(fl validator.FieldLevel) bool {
	color := fl.Field().String()
	if len(color) != 7 || !strings.HasPrefix(color, "#") {
		return false
	}
	
	// Check if all characters after # are valid hex
	for _, char := range color[1:] {
		if !((char >= '0' && char <= '9') || 
			 (char >= 'A' && char <= 'F') || 
			 (char >= 'a' && char <= 'f')) {
			return false
		}
	}
	return true
}

// Get validation error message
func getValidationMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return err.Field() + " is required"
	case "min":
		return err.Field() + " must be at least " + err.Param()
	case "max":
		return err.Field() + " must be at most " + err.Param()
	case "email":
		return err.Field() + " must be a valid email address"
	case "url":
		return err.Field() + " must be a valid URL"
	case "notempty":
		return err.Field() + " cannot be empty"
	case "port":
		return err.Field() + " must be a valid port number (1-65535)"
	case "hexcolor":
		return err.Field() + " must be a valid hex color (e.g., #FF0000)"
	default:
		return err.Field() + " is invalid"
	}
}

// ValidateQueryParams validates query parameters
func ValidateQueryParams(rules map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var validationErrors []ValidationError
		
		for field, rule := range rules {
			value := c.Query(field)
			if value == "" && strings.Contains(rule, "required") {
				validationErrors = append(validationErrors, ValidationError{
					Field:   field,
					Tag:     "required",
					Value:   value,
					Message: field + " is required",
				})
			}
		}
		
		if len(validationErrors) > 0 {
			HandleValidationError(c, validationErrors)
			return
		}
		
		c.Next()
	}
}
