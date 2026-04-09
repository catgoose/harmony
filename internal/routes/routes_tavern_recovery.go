// setup:feature:demo

package routes

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"catgoose/harmony/internal/demo"
	"catgoose/harmony/internal/routes/handler"
	"catgoose/harmony/web/views"

	"github.com/catgoose/tavern"
	"github.com/labstack/echo/v4"
)

// Topics for the recovery lab. Each topic exercises a different recovery
// strategy on a single SSE connection.
const (
	topicRecoveryReplay   = "recovery/replay"
	topicRecoverySnapshot = "recovery/snapshot"
	topicRecoveryLive     = "recovery/live"
)

type recoveryRoutes struct {
	broker *tavern.SSEBroker
	lab    *demo.RecoveryLab
}

func (ar *appRoutes) initRecoveryRoutes(broker *tavern.SSEBroker) {
	r := &recoveryRoutes{
		broker: broker,
		lab:    demo.NewRecoveryLab(),
	}

	// Replay topic gets a small replay buffer so reconnecting clients can
	// recover via Last-Event-ID. Snapshot topic doesn't use replay; it relies
	// on per-subscriber snapshot delivery. Live topic gets neither.
	broker.SetReplayPolicy(topicRecoveryReplay, 50)

	ar.e.GET("/realtime/tavern/recovery", r.handlePage)
	ar.e.GET("/sse/tavern/recovery", r.handleSSE)
	ar.e.POST("/realtime/tavern/recovery/snapshot", r.handleSetSnapshot)
	ar.e.POST("/realtime/tavern/recovery/reset", r.handleReset)

	broker.RunPublisher(ar.ctx, r.startPublisher)
}

func (r *recoveryRoutes) handlePage(c echo.Context) error {
	return handler.RenderBaseLayout(c, views.TavernRecoveryPage(r.lab.Snapshot()))
}

func (r *recoveryRoutes) handleSetSnapshot(c echo.Context) error {
	value := strings.TrimSpace(c.FormValue("value"))
	if value == "" {
		return c.String(http.StatusBadRequest, "value is required")
	}
	r.lab.SetSnapshot(value)
	// Publish the new snapshot so currently-connected clients see the change
	// in their snapshot region (in addition to fresh subscribers receiving it
	// on connect via SubscribeWithSnapshot).
	html := renderRecoverySnapshotEvent(value, time.Now().Format("15:04:05.000"), false)
	r.broker.Publish(topicRecoverySnapshot, tavern.NewSSEMessage("recovery-snapshot", html).String())
	return c.NoContent(http.StatusNoContent)
}

func (r *recoveryRoutes) handleReset(c echo.Context) error {
	r.lab.Reset()
	r.broker.ClearReplay(topicRecoveryReplay)
	return c.NoContent(http.StatusNoContent)
}

// handleSSE multiplexes all three recovery topics into one SSE response.
// On reconnect, the replay region uses Last-Event-ID, the snapshot region
// gets the current snapshot value, and the live region just resumes streaming.
func (r *recoveryRoutes) handleSSE(c echo.Context) error {
	flusher, err := startSSEResponse(c)
	if err != nil {
		return err
	}

	lastEventID := c.Request().Header.Get("Last-Event-ID")
	reconnect := lastEventID != ""

	// Replay subscription: SubscribeFromID for reconnects, SubscribeMulti
	// otherwise. SubscribeFromID delivers any events newer than lastEventID.
	var replayCh <-chan string
	var replayUnsub func()
	if reconnect {
		replayCh, replayUnsub = r.broker.SubscribeFromID(topicRecoveryReplay, lastEventID)
	} else {
		replayCh, replayUnsub = r.broker.Subscribe(topicRecoveryReplay)
	}
	defer replayUnsub()

	// Snapshot subscription: deliver the current snapshot atomically before
	// any live events. The snapshot HTML is tagged so the client can show
	// "SNAPSHOT" rather than "LIVE" on the first delivery.
	snapCh, snapUnsub := r.broker.SubscribeWithSnapshot(topicRecoverySnapshot, func() string {
		html := renderRecoverySnapshotEvent(r.lab.Snapshot(), time.Now().Format("15:04:05.000"), true)
		return tavern.NewSSEMessage("recovery-snapshot", html).String()
	})
	defer snapUnsub()

	// Live subscription: no replay, no snapshot.
	liveCh, liveUnsub := r.broker.Subscribe(topicRecoveryLive)
	defer liveUnsub()

	// Tell the client whether this was a fresh connect or a reconnect, so
	// the regions can show the right "first delivery" badges.
	mode := "fresh"
	if reconnect {
		mode = "reconnect"
	}
	if _, werr := fmt.Fprintf(c.Response(), "event: recovery-mode\ndata: %s\n\n", mode); werr != nil {
		return nil
	}
	flusher.Flush()

	ctx := c.Request().Context()
	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-replayCh:
			if !ok {
				return nil
			}
			_, _ = fmt.Fprint(c.Response(), msg)
			flusher.Flush()
		case msg, ok := <-snapCh:
			if !ok {
				return nil
			}
			_, _ = fmt.Fprint(c.Response(), msg)
			flusher.Flush()
		case msg, ok := <-liveCh:
			if !ok {
				return nil
			}
			_, _ = fmt.Fprint(c.Response(), msg)
			flusher.Flush()
		case <-heartbeat.C:
			_, _ = fmt.Fprintf(c.Response(), ": heartbeat\n\n")
			flusher.Flush()
		}
	}
}

// startPublisher emits replay and live events on a slow timer so the demo
// always has fresh data without overwhelming the page.
func (r *recoveryRoutes) startPublisher(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Replay event with stable ID for Last-Event-ID recovery.
			id, seq, ts := r.lab.NextReplayEvent()
			html := renderRecoveryReplayEvent(seq, id, ts.Format("15:04:05.000"), false)
			msg := tavern.NewSSEMessage("recovery-replay", html).WithID(id).String()
			r.broker.PublishWithID(topicRecoveryReplay, id, msg)

			// Live event with no replay.
			liveSeq, liveTS := r.lab.NextLiveEvent()
			liveHTML := renderRecoveryLiveEvent(liveSeq, liveTS.Format("15:04:05.000"))
			r.broker.Publish(topicRecoveryLive, tavern.NewSSEMessage("recovery-live", liveHTML).String())
		}
	}
}

func renderRecoveryReplayEvent(seq int64, id, timestamp string, replayed bool) string {
	return renderToString("render recovery replay", views.RecoveryReplayEvent(seq, id, timestamp, replayed))
}

func renderRecoverySnapshotEvent(value, timestamp string, snapshot bool) string {
	return renderToString("render recovery snapshot", views.RecoverySnapshotEvent(value, timestamp, snapshot))
}

func renderRecoveryLiveEvent(seq int64, timestamp string) string {
	return renderToString("render recovery live", views.RecoveryLiveEvent(seq, timestamp))
}
