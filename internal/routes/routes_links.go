// setup:feature:demo

package routes

import "github.com/catgoose/linkwell"

// initLinkRelations registers all link relation declarations for the app.
// Hubs define parent→child discovery pages. Rings define peer groups.
// This is the single source of truth for the app's navigation topology.
func (ar *appRoutes) initLinkRelations() {
	// ── Hubs (discovery / index pages) ──────────────────────────────

	linkwell.Hub("/demo", "Demo",
		linkwell.Rel("/demo/inventory", "Inventory"),
		linkwell.Rel("/demo/catalog", "Catalog"),
		linkwell.Rel("/demo/people", "People"),
		linkwell.Rel("/demo/kanban", "Kanban"),
		linkwell.Rel("/demo/approvals", "Approvals"),
		linkwell.Rel("/demo/vendors", "Vendors"),
		linkwell.Rel("/demo/feed", "Feed"),
		linkwell.Rel("/demo/canvas", "Canvas"),
		linkwell.Rel("/demo/settings", "Settings"),
		linkwell.Rel("/demo/bulk", "Bulk"),
		linkwell.Rel("/demo/logging", "Logging"),
		linkwell.Rel("/demo/repository", "Repository"),
		linkwell.Rel("/dashboard", "Dashboard"),
		linkwell.Rel("/pwa", "PWA Offline"),
	)

	linkwell.Hub("/hypermedia", "Hypermedia",
		linkwell.Rel("/hypermedia/controls", "Controls"),
		linkwell.Rel("/hypermedia/crud", "CRUD"),
		linkwell.Rel("/hypermedia/lists", "Lists"),
		linkwell.Rel("/hypermedia/interactions", "Interactions"),
		linkwell.Rel("/hypermedia/state", "State"),
		linkwell.Rel("/hypermedia/errors", "Errors"),
		linkwell.Rel("/hypermedia/components", "Components"),
		linkwell.Rel("/hypermedia/components2", "Components 2"),
		linkwell.Rel("/hypermedia/components3", "Components 3"),
		linkwell.Rel("/hypermedia/realtime", "Realtime"),
		linkwell.Rel("/hypermedia/links", "Links"),
		linkwell.Rel("/hypermedia/hal", "HAL"),
		linkwell.Rel("/hypermedia/standards", "Standards"),
	)

	linkwell.Hub("/admin", "Admin",
		linkwell.Rel("/admin/health", "Health"),
		linkwell.Rel("/admin/sessions", "Sessions"),
		linkwell.Rel("/admin/settings", "Control Panel"),
		linkwell.Rel("/admin/error-traces", "Error Traces"),
		linkwell.Rel("/admin/error-reports", "Error Reports"),
		linkwell.Rel("/admin/system", "System"),
		linkwell.Rel("/admin/config", "Config"),
	)

	linkwell.Hub("/dashboard", "Dashboard",
		linkwell.Rel("/demo/inventory", "Inventory"),
		linkwell.Rel("/demo/people", "People"),
		linkwell.Rel("/demo/kanban", "Kanban"),
		linkwell.Rel("/demo/approvals", "Approvals"),
		linkwell.Rel("/demo/vendors", "Vendors"),
		linkwell.Rel("/demo/feed", "Feed"),
	)

	// ── Rings (peer groups) ─────────────────────────────────────────

	// Demo: data management pages
	linkwell.Ring("Data",
		linkwell.Rel("/demo/inventory", "Inventory"),
		linkwell.Rel("/demo/catalog", "Catalog"),
		linkwell.Rel("/demo/bulk", "Bulk Ops"),
		linkwell.Rel("/demo/people", "People"),
		linkwell.Rel("/demo/vendors", "Vendors"),
	)

	// Demo: workflow and process pages
	linkwell.Ring("Workflow",
		linkwell.Rel("/demo/kanban", "Kanban"),
		linkwell.Rel("/demo/approvals", "Approvals"),
		linkwell.Rel("/demo/feed", "Feed"),
	)

	// Demo: tools and utilities
	linkwell.Ring("Utility",
		linkwell.Rel("/demo/logging", "Logging"),
		linkwell.Rel("/demo/canvas", "Canvas"),
		linkwell.Rel("/demo/settings", "Settings"),
		linkwell.Rel("/demo/repository", "Repository"),
	)

	// Demo: dashboard children are also peers
	linkwell.Ring("Dashboard",
		linkwell.Rel("/demo/inventory", "Inventory"),
		linkwell.Rel("/demo/people", "People"),
		linkwell.Rel("/demo/kanban", "Kanban"),
		linkwell.Rel("/demo/approvals", "Approvals"),
		linkwell.Rel("/demo/vendors", "Vendors"),
		linkwell.Rel("/demo/feed", "Feed"),
	)

	// Hypermedia: pattern pages
	linkwell.Ring("Patterns",
		linkwell.Rel("/hypermedia/controls", "Controls"),
		linkwell.Rel("/hypermedia/crud", "CRUD"),
		linkwell.Rel("/hypermedia/lists", "Lists"),
		linkwell.Rel("/hypermedia/interactions", "Interactions"),
		linkwell.Rel("/hypermedia/state", "State"),
		linkwell.Rel("/hypermedia/errors", "Errors"),
		linkwell.Rel("/hypermedia/realtime", "Realtime"),
		linkwell.Rel("/hypermedia/links", "Links"),
		linkwell.Rel("/hypermedia/hal", "HAL"),
		linkwell.Rel("/hypermedia/standards", "Standards"),
	)

	// Hypermedia: component gallery pages
	linkwell.Ring("Components",
		linkwell.Rel("/hypermedia/components", "Components"),
		linkwell.Rel("/hypermedia/components2", "Components 2"),
		linkwell.Rel("/hypermedia/components3", "Components 3"),
	)

	// Admin: operational pages
	linkwell.Ring("Admin Ops",
		linkwell.Rel("/admin/health", "Health"),
		linkwell.Rel("/admin/error-traces", "Error Traces"),
		linkwell.Rel("/admin/error-reports", "Error Reports"),
		linkwell.Rel("/admin/sessions", "Sessions"),
		linkwell.Rel("/admin/settings", "Control Panel"),
	)

	// Admin: system introspection
	linkwell.Ring("System",
		linkwell.Rel("/admin/system", "System"),
		linkwell.Rel("/admin/config", "Config"),
		linkwell.Rel("/admin/health", "Health"),
		linkwell.Rel("/admin/error-traces", "Error Traces"),
	)

	// ── Settings ────────────────────────────────────────────────────

	linkwell.Link("/settings", "related", "/user/settings", "Preferences")
	linkwell.Link("/settings", "related", "/admin/config", "Admin Config")
	linkwell.Link("/settings", "related", "/admin/settings", "Control Panel")

	// ── Action relations ────────────────────────────────────────────

	// List pages with create forms
	linkwell.Link("/demo/inventory", "create-form", "/demo/inventory/items/new", "New Item")
	linkwell.Link("/demo/repository", "create-form", "/demo/repository/tasks", "New Task")
}
