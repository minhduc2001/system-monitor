package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string      `json:"error"`
	Message string      `json:"message"`
	Code    int         `json:"code"`
	Details interface{} `json:"details,omitempty"`
	Trace   string      `json:"trace,omitempty"`
}

// ErrorHandler is a global error handler middleware
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the error
				log.Printf("Panic recovered: %v\n%s", err, debug.Stack())

				// Create error response
				errorResp := ErrorResponse{
					Error:   "Internal Server Error",
					Message: "An unexpected error occurred",
					Code:    http.StatusInternalServerError,
					Trace:   string(debug.Stack()),
				}

				// Send error response
				c.JSON(http.StatusInternalServerError, errorResp)
				c.Abort()
			}
		}()

		c.Next()
	}
}

// CustomError represents a custom application error
type CustomError struct {
	Code    int
	Message string
	Details interface{}
}

func (e *CustomError) Error() string {
	return e.Message
}

// NewError creates a new custom error
func NewError(code int, message string, details interface{}) *CustomError {
	return &CustomError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// HandleError handles custom errors and sends appropriate response
func HandleError(c *gin.Context, err error) {
	switch e := err.(type) {
	case *CustomError:
		errorResp := ErrorResponse{
			Error:   http.StatusText(e.Code),
			Message: e.Message,
			Code:    e.Code,
			Details: e.Details,
		}
		c.JSON(e.Code, errorResp)
	default:
		errorResp := ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		}
		c.JSON(http.StatusInternalServerError, errorResp)
	}
}

// Common error types
var (
	ErrNotFound     = &CustomError{Code: http.StatusNotFound, Message: "Resource not found"}
	ErrBadRequest   = &CustomError{Code: http.StatusBadRequest, Message: "Bad request"}
	ErrUnauthorized = &CustomError{Code: http.StatusUnauthorized, Message: "Unauthorized"}
	ErrForbidden    = &CustomError{Code: http.StatusForbidden, Message: "Forbidden"}
	ErrConflict     = &CustomError{Code: http.StatusConflict, Message: "Resource conflict"}
	ErrValidation   = &CustomError{Code: http.StatusUnprocessableEntity, Message: "Validation failed"}
)

// ValidationError represents validation errors
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// HandleValidationError handles validation errors
func HandleValidationError(c *gin.Context, errors []ValidationError) {
	errorResp := ErrorResponse{
		Error:   "Validation Failed",
		Message: "The request contains invalid data",
		Code:    http.StatusUnprocessableEntity,
		Details: errors,
	}
	c.JSON(http.StatusUnprocessableEntity, errorResp)
}
