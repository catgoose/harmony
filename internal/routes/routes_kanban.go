// setup:feature:demo

package routes

import (
	"catgoose/harmony/internal/demo"
	"catgoose/harmony/internal/routes/handler"
	"catgoose/harmony/internal/routes/params"
	"github.com/catgoose/tavern"
	"catgoose/harmony/web/views"

	"github.com/labstack/echo/v4"
)

type kanbanRoutes struct {
	board  *demo.KanbanBoard
	actLog *demo.ActivityLog
	broker *tavern.SSEBroker
}

func (ar *appRoutes) initKanbanRoutes(board *demo.KanbanBoard, actLog *demo.ActivityLog, broker *tavern.SSEBroker) {
	k := &kanbanRoutes{board: board, actLog: actLog, broker: broker}
	ar.e.GET("/apps/kanban", k.handleKanbanPage)
	ar.e.PATCH("/apps/kanban/tasks/:id", k.handleMoveTask)
}

func (k *kanbanRoutes) handleKanbanPage(c echo.Context) error {
	tasks := k.board.AllTasks()
	return handler.RenderBaseLayout(c, views.KanbanPage(tasks))
}

func (k *kanbanRoutes) handleMoveTask(c echo.Context) error {
	id, err := params.ParseParamID(c, "id")
	if err != nil {
		return handler.HandleHypermediaError(c, 400, "Invalid task ID", err)
	}
	newStatus := c.FormValue("status")
	task, ok := k.board.MoveTask(id, newStatus)
	if !ok {
		return handler.HandleHypermediaError(c, 404, "Task not found or invalid status", nil)
	}
	evt := k.actLog.Record("moved", "task", id, task.Title, "moved to "+newStatus)
	BroadcastActivity(k.broker, evt)
	tasks := k.board.AllTasks()
	return handler.RenderComponent(c, views.KanbanBoard(tasks))
}
