package apiresponse

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequestIDContextKey must match middleware that sets the Gin context value.
const RequestIDContextKey = "request_id"

type Envelope struct {
	Success bool           `json:"success"`
	Data    any            `json:"data,omitempty"`
	Error   *APIError      `json:"error,omitempty"`
	Meta    map[string]any `json:"meta,omitempty"`
	TraceID string         `json:"trace_id,omitempty"`
}

type APIError struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Fields  []FieldError `json:"fields,omitempty"`
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func TraceID(c *gin.Context) string {
	if value, ok := c.Get(RequestIDContextKey); ok {
		if id, ok := value.(string); ok {
			return id
		}
	}
	return c.GetHeader("X-Request-ID")
}

func OK(c *gin.Context, data any) {
	Respond(c, http.StatusOK, data, nil)
}

func Respond(c *gin.Context, status int, data any, meta map[string]any) {
	c.JSON(status, Envelope{
		Success: true,
		Data:    data,
		Meta:    meta,
		TraceID: TraceID(c),
	})
}

func Error(c *gin.Context, status int, code string, message string) {
	c.JSON(status, Envelope{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
		TraceID: TraceID(c),
	})
}

func ValidationError(c *gin.Context, fields []FieldError) {
	c.JSON(http.StatusUnprocessableEntity, Envelope{
		Success: false,
		Error: &APIError{
			Code:    "validation_error",
			Message: "The request contains invalid fields.",
			Fields:  fields,
		},
		TraceID: TraceID(c),
	})
}
