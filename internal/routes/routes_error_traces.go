package routes

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"catgoose/dothog/internal/requestlog"
	"catgoose/dothog/internal/routes/handler"
	"catgoose/dothog/internal/routes/hypermedia"
	"catgoose/dothog/web/views"

	hx "catgoose/dothog/internal/routes/htmx"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

const errorTracesBase = "/admin/error-traces"

func (ar *appRoutes) initErrorTracesRoutes() {
	if ar.reqLogStore == nil {
		return
	}
	ar.e.GET(errorTracesBase, ar.handleErrorTracesPage)
	ar.e.GET(errorTracesBase+"/list", ar.handleErrorTracesList)
	ar.e.GET(errorTracesBase+"/:requestID", ar.handleErrorTraceDetail)
	ar.e.DELETE(errorTracesBase+"/:requestID", ar.handleErrorTraceDelete)
}

func (ar *appRoutes) handleErrorTracesPage(c echo.Context) error {
	bar, container, err := ar.buildErrorTracesContent(c)
	if err != nil {
		return handler.HandleHypermediaError(c, 500, "Failed to load error traces", err)
	}
	return handler.RenderBaseLayout(c, views.ErrorTracesPage(bar, container))
}

func (ar *appRoutes) handleErrorTracesList(c echo.Context) error {
	_, container, err := ar.buildErrorTracesContent(c)
	if err != nil {
		return handler.HandleHypermediaError(c, 500, "Failed to load error traces", err)
	}
	if hx.IsHTMX(c) {
		pushURL := errorTracesBase
		if q := c.Request().URL.RawQuery; q != "" {
			pushURL += "?" + q
		}
		hx.ReplaceURL(c, pushURL)
	}
	return handler.RenderComponent(c, container)
}

func (ar *appRoutes) handleErrorTraceDetail(c echo.Context) error {
	requestID := c.Param("requestID")
	trace := ar.reqLogStore.Get(requestID)
	if trace == nil {
		return handler.HandleHypermediaError(c, 404, "Error trace not found", nil)
	}
	return handler.RenderComponent(c, views.ErrorTraceDetailContent(trace))
}

func (ar *appRoutes) handleErrorTraceDelete(c echo.Context) error {
	requestID := c.Param("requestID")
	if err := ar.reqLogStore.DeleteTrace(requestID); err != nil {
		return handler.HandleHypermediaError(c, 500, "Failed to delete trace", err)
	}
	// Re-apply current filters from HX-Current-URL
	if raw := c.Request().Header.Get("HX-Current-URL"); raw != "" {
		if u, err := url.Parse(raw); err == nil && u.RawQuery != "" {
			c.Request().URL.RawQuery = u.RawQuery
		}
	}
	_, container, err := ar.buildErrorTracesContent(c)
	if err != nil {
		return handler.HandleHypermediaError(c, 500, "Failed to reload traces", err)
	}
	return handler.RenderComponent(c, container)
}

func (ar *appRoutes) buildErrorTracesContent(c echo.Context) (hypermedia.FilterBar, templ.Component, error) {
	const perPage = 20
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	f := requestlog.TraceFilter{
		Q:       c.QueryParam("q"),
		Status:  c.QueryParam("status"),
		Method:  c.QueryParam("method"),
		Sort:    c.QueryParam("sort"),
		Dir:     c.QueryParam("dir"),
		Page:    page,
		PerPage: perPage,
	}

	traces, total, err := ar.reqLogStore.ListTraces(f)
	if err != nil {
		return hypermedia.FilterBar{}, nil, err
	}

	target := "#error-traces-table-container"
	listURL := errorTracesBase + "/list"

	bar := hypermedia.NewFilterBar(listURL, target,
		hypermedia.SearchField("q", "Search routes, errors, request IDs, users\u2026", f.Q),
		hypermedia.SelectField("status", "Status", f.Status,
			hypermedia.SelectOptions(f.Status,
				"", "All",
				"4xx", "4xx Client",
				"5xx", "5xx Server",
				"400", "400",
				"401", "401",
				"403", "403",
				"404", "404",
				"500", "500",
				"502", "502",
				"504", "504",
			)),
		hypermedia.SelectField("method", "Method", f.Method,
			hypermedia.SelectOptions(f.Method,
				"", "All",
				"GET", "GET",
				"POST", "POST",
				"PUT", "PUT",
				"DELETE", "DELETE",
			)),
	)

	sortBase := traceStripParams(c.Request().URL, "sort", "dir")
	cols := []hypermedia.TableCol{
		hypermedia.SortableCol("CreatedAt", "Time", f.Sort, f.Dir, sortBase, target, "#filter-form"),
		hypermedia.SortableCol("StatusCode", "Status", f.Sort, f.Dir, sortBase, target, "#filter-form"),
		hypermedia.SortableCol("Method", "Method", f.Sort, f.Dir, sortBase, target, "#filter-form"),
		hypermedia.SortableCol("Route", "Route", f.Sort, f.Dir, sortBase, target, "#filter-form"),
		{Label: "Error"},
		{Label: "IP"},
		{Label: ""},
	}

	pageBase := traceStripParams(c.Request().URL, "page")
	info := hypermedia.PageInfo{
		Page:       page,
		PerPage:    perPage,
		TotalItems: total,
		TotalPages: hypermedia.ComputeTotalPages(total, perPage),
		BaseURL:    pageBase,
		Target:     target,
		Include:    "#filter-form",
	}

	body := views.ErrorTracesBody(traces)
	container := views.ErrorTracesTableContainer(cols, body, info)
	return bar, container, nil
}

// traceStripParams returns a copy of u with the named query params removed.
func traceStripParams(u *url.URL, params ...string) string {
	cp := *u
	q := cp.Query()
	for _, p := range params {
		q.Del(p)
	}
	cp.RawQuery = q.Encode()
	return cp.String()
}

// SeedErrorTraces inserts 1000 demo error traces spread over the past 3 years.
func SeedErrorTraces(store *requestlog.Store) {
	// Check if already seeded
	existing, _, _ := store.ListTraces(requestlog.TraceFilter{Page: 1, PerPage: 1})
	if len(existing) > 0 {
		return
	}

	type errorTemplate struct {
		ErrorChain string
		StatusCode int
		Route      string
		Method     string
		Entries    []requestlog.Entry
	}

	templates := []errorTemplate{
		{
			ErrorChain: "get user %d: sql: no rows in result set",
			StatusCode: 404, Route: "/api/users/%d", Method: "GET",
			Entries: []requestlog.Entry{
				{Level: "INFO", Message: "Querying user by ID"},
				{Level: "ERROR", Message: "User not found"},
			},
		},
		{
			ErrorChain: "process order: validate inventory: insufficient stock for SKU-%04d",
			StatusCode: 422, Route: "/api/orders", Method: "POST",
			Entries: []requestlog.Entry{
				{Level: "INFO", Message: "Parsing order payload"},
				{Level: "INFO", Message: "Validating inventory"},
				{Level: "WARN", Message: "Low stock detected"},
				{Level: "ERROR", Message: "Insufficient stock"},
			},
		},
		{
			ErrorChain: "render dashboard: query metrics: context deadline exceeded",
			StatusCode: 504, Route: "/dashboard", Method: "GET",
			Entries: []requestlog.Entry{
				{Level: "INFO", Message: "Fetching metrics"},
				{Level: "WARN", Message: "Query slow"},
				{Level: "ERROR", Message: "Context deadline exceeded"},
			},
		},
		{
			ErrorChain: "upload file: multipart: NextPart: unexpected EOF",
			StatusCode: 400, Route: "/api/files/upload", Method: "POST",
			Entries: []requestlog.Entry{
				{Level: "INFO", Message: "Parsing multipart form"},
				{Level: "ERROR", Message: "Multipart parse failed"},
			},
		},
		{
			ErrorChain: "save settings: database is locked",
			StatusCode: 500, Route: "/settings/theme", Method: "POST",
			Entries: []requestlog.Entry{
				{Level: "INFO", Message: "Updating theme"},
				{Level: "ERROR", Message: "Database write failed"},
			},
		},
		{
			ErrorChain: "fetch report: connect: connection refused",
			StatusCode: 502, Route: "/api/reports/monthly", Method: "GET",
			Entries: []requestlog.Entry{
				{Level: "INFO", Message: "Calling reporting service"},
				{Level: "ERROR", Message: "Upstream connection refused"},
			},
		},
		{
			ErrorChain: "authenticate: token expired",
			StatusCode: 401, Route: "/api/protected/data", Method: "GET",
			Entries: []requestlog.Entry{
				{Level: "WARN", Message: "Token validation failed"},
				{Level: "ERROR", Message: "Authentication failed"},
			},
		},
		{
			ErrorChain: "create item: UNIQUE constraint failed: items.name",
			StatusCode: 409, Route: "/demo/inventory/items", Method: "POST",
			Entries: []requestlog.Entry{
				{Level: "INFO", Message: "Creating item"},
				{Level: "ERROR", Message: "Insert failed"},
			},
		},
		{
			ErrorChain: "authorize /admin/settings: role viewer cannot access admin resource",
			StatusCode: 403, Route: "/admin/settings", Method: "GET",
			Entries: []requestlog.Entry{
				{Level: "INFO", Message: "Authenticating user"},
				{Level: "WARN", Message: "Authorization denied"},
			},
		},
		{
			ErrorChain: "update item %d: optimistic lock: version mismatch",
			StatusCode: 409, Route: "/api/items/%d", Method: "PUT",
			Entries: []requestlog.Entry{
				{Level: "INFO", Message: "Loading item for update"},
				{Level: "ERROR", Message: "Version conflict detected"},
			},
		},
		{
			ErrorChain: "delete user %d: foreign key constraint: user has active orders",
			StatusCode: 409, Route: "/api/users/%d", Method: "DELETE",
			Entries: []requestlog.Entry{
				{Level: "INFO", Message: "Checking user dependencies"},
				{Level: "ERROR", Message: "Cannot delete: active orders exist"},
			},
		},
		{
			ErrorChain: "parse JSON body: unexpected end of JSON input",
			StatusCode: 400, Route: "/api/webhooks/stripe", Method: "POST",
			Entries: []requestlog.Entry{
				{Level: "INFO", Message: "Receiving webhook"},
				{Level: "ERROR", Message: "Malformed JSON body"},
			},
		},
	}

	users := []string{
		"alice@contoso.com", "bob@contoso.com", "charlie@contoso.com",
		"dana@contoso.com", "eve@contoso.com", "frank@contoso.com",
		"grace@contoso.com", "hank@contoso.com", "",
	}
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
		"Mozilla/5.0 (X11; Linux x86_64; rv:109.0)",
		"curl/8.1.2",
		"PostmanRuntime/7.32.3",
	}
	ips := []string{
		"192.168.1.100", "10.0.0.50", "172.16.0.25", "192.168.1.55",
		"10.0.0.12", "172.16.0.30", "192.168.1.200", "10.0.0.5",
	}

	now := time.Now()
	threeYears := 3 * 365 * 24 * time.Hour
	const count = 1000

	for i := 0; i < count; i++ {
		// Spread evenly over 3 years
		offset := threeYears * time.Duration(i) / count
		createdAt := now.Add(-threeYears + offset)

		tmpl := templates[i%len(templates)]
		id := i + 1

		errorChain := fmt.Sprintf(tmpl.ErrorChain, id)
		route := fmt.Sprintf(tmpl.Route, id)

		store.PromoteAt(requestlog.ErrorTrace{
			RequestID:  fmt.Sprintf("seed-%08x", i),
			ErrorChain: errorChain,
			StatusCode: tmpl.StatusCode,
			Route:      route,
			Method:     tmpl.Method,
			UserAgent:  userAgents[i%len(userAgents)],
			RemoteIP:   ips[i%len(ips)],
			UserID:     users[i%len(users)],
			Entries:    tmpl.Entries,
		}, createdAt)
	}
}
