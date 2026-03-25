package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/catgoose/promolog"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestGetRequestID_ReturnsIDFromContext(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	ctx := context.WithValue(req.Context(), promolog.RequestIDKey, "test-id-123")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	assert.Equal(t, "test-id-123", GetRequestID(c))
}

func TestGetRequestID_EmptyWithoutMiddleware(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	assert.Empty(t, GetRequestID(c))
}
