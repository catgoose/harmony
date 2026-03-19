// setup:feature:session_settings

package middleware

import (
	"time"

	"catgoose/dothog/internal/domain"
	"catgoose/dothog/internal/logger"
	"catgoose/dothog/internal/repository"

	"github.com/labstack/echo/v4"
)

const (
	settingsContextKey = "sessionSettings"
	// sharedSessionUUID is used for all visitors so the demo behaves as a
	// single-user application — every browser reads/writes the same row.
	sharedSessionUUID = "00000000-0000-0000-0000-000000000000"
)

// SessionSettingsMiddleware loads the shared session settings row and stores
// it on the echo context. All visitors share the same settings.
func SessionSettingsMiddleware(repo repository.SessionSettingsRepository) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()

			settings, err := repo.GetByUUID(ctx, sharedSessionUUID)
			if err != nil {
				logger.WithContext(ctx).Error("Failed to load session settings", "error", err)
				settings = domain.NewDefaultSettings(sharedSessionUUID)
			}
			if settings == nil {
				settings = domain.NewDefaultSettings(sharedSessionUUID)
				if err := repo.Upsert(ctx, settings); err != nil {
					logger.WithContext(ctx).Error("Failed to create session settings", "error", err)
				}
			}

			// Touch if last update was more than 24 hours ago to keep the row fresh.
			if time.Since(settings.UpdatedAt) > 24*time.Hour {
				_ = repo.Touch(ctx, sharedSessionUUID)
			}

			c.Set(settingsContextKey, settings)
			return next(c)
		}
	}
}

// GetSessionSettings returns the session settings from the echo context.
func GetSessionSettings(c echo.Context) *domain.SessionSettings {
	if s, ok := c.Get(settingsContextKey).(*domain.SessionSettings); ok {
		return s
	}
	return domain.NewDefaultSettings(sharedSessionUUID)
}
