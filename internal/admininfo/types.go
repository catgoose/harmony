// setup:feature:demo

// Package admininfo provides types for admin dashboard pages.
package admininfo

// SystemInfo holds runtime stats for the admin system page.
type SystemInfo struct {
	TotalAllocMB    string
	SysMB           string
	OS              string
	Arch            string
	NextGCMB        string
	HeapAllocMB     string
	Version         string
	Uptime          string
	GoVersion       string
	StackInUseMB    string
	HeapSysMB       string
	Goroutines      int
	NumThread       int
	LastPauseMicros uint64
	NumCPU          int
	HeapObjects     uint64
	LiveObjects     uint64
	GCCycles        uint32
}

// ConfigEntry is a single key/value pair for display, with optional masking.
type ConfigEntry struct {
	Key   string
	Value string
}

// UserPreferences holds per-session user preferences.
// These are stored in-memory keyed by session cookie; applications should
// persist them to their own database when needed.
type UserPreferences struct {
	DateFormat           string
	PageSize             int
	CompactTables        bool
	EmailOnError         bool
	DesktopNotifications bool
	ReduceMotion         bool
	HighContrast         bool
}

// DefaultUserPreferences returns sensible defaults.
func DefaultUserPreferences() UserPreferences {
	return UserPreferences{
		PageSize:   20,
		DateFormat: "relative",
	}
}
