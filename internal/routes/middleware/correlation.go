package middleware

import (
	"github.com/catgoose/promolog"
	"github.com/labstack/echo/v4"
)

// GetRequestID retrieves the request ID from the request context.
// The request ID is set by promolog.CorrelationMiddleware on every request.
func GetRequestID(c echo.Context) string {
	return promolog.GetRequestID(c.Request().Context())
}
