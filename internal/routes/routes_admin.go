// setup:feature:demo

package routes

import (
	"fmt"

	"catgoose/harmony/internal/demo"
	"catgoose/harmony/internal/routes/handler"
	"github.com/catgoose/tavern"
	"catgoose/harmony/web/views"

	"github.com/labstack/echo/v4"
)

type adminRoutes struct {
	db     *demo.DB
	actLog *demo.ActivityLog
	broker *tavern.SSEBroker
}

func (ar *appRoutes) initAdminRoutes(db *demo.DB, actLog *demo.ActivityLog, broker *tavern.SSEBroker) {
	a := &adminRoutes{db: db, actLog: actLog, broker: broker}
	ar.e.GET("/admin/sqlite", a.handleAdminPage)
	ar.e.POST("/admin/db/reinit", a.handleReinit)
}

func (a *adminRoutes) handleAdminPage(c echo.Context) error {
	info, _ := a.db.GetSchemaInfo(c.Request().Context())
	return handler.RenderBaseLayout(c, views.AdminPage(info))
}

func (a *adminRoutes) handleReinit(c echo.Context) error {
	if err := a.db.Reset(); err != nil {
		info, _ := a.db.GetSchemaInfo(c.Request().Context())
		return handler.RenderComponent(c, views.AdminDBStatus(info, fmt.Sprintf("Reset failed: %s", err), true))
	}
	info, _ := a.db.GetSchemaInfo(c.Request().Context())

	evt := a.actLog.Record("created", "database", 0, "Database", "re-initialized with seed data")
	BroadcastActivity(a.broker, evt)

	return handler.RenderComponent(c, views.AdminDBStatus(info, "Database re-initialized with seed data", false))
}
