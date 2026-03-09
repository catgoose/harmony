package schema

import (
	"fmt"
	"strings"

	"catgoose/go-htmx-demo/internals/database/dialect"
)

// TableDef defines a table schema.
type TableDef struct {
	Name          string
	cols          []ColumnDef
	indexes       []IndexDef
	hasSoftDelete bool
	hasVersion    bool
	hasExpiry     bool
}

// NewTable creates a new table definition.
func NewTable(name string) *TableDef {
	return &TableDef{Name: name}
}

// Columns appends column definitions.
func (t *TableDef) Columns(cols ...ColumnDef) *TableDef {
	t.cols = append(t.cols, cols...)
	return t
}

// WithTimestamps appends CreatedAt and UpdatedAt columns.
func (t *TableDef) WithTimestamps() *TableDef {
	t.cols = append(t.cols, TimestampColumnDefs()...)
	return t
}

// WithSoftDelete appends a DeletedAt column and marks the table for soft-delete.
func (t *TableDef) WithSoftDelete() *TableDef {
	t.hasSoftDelete = true
	t.cols = append(t.cols, SoftDeleteColumnDefs()...)
	return t
}

// WithAuditTrail appends CreatedBy, UpdatedBy, and DeletedBy columns.
func (t *TableDef) WithAuditTrail() *TableDef {
	t.cols = append(t.cols, AuditColumnDefs()...)
	return t
}

// WithVersion appends a Version column for optimistic concurrency control.
func (t *TableDef) WithVersion() *TableDef {
	t.hasVersion = true
	t.cols = append(t.cols, VersionColumnDefs()...)
	return t
}

// WithSortOrder appends a SortOrder column for manual ordering.
func (t *TableDef) WithSortOrder() *TableDef {
	t.cols = append(t.cols, SortOrderColumnDefs()...)
	return t
}

// WithStatus appends a Status column with the given default value.
func (t *TableDef) WithStatus(defaultStatus string) *TableDef {
	t.cols = append(t.cols, StatusColumnDefs(defaultStatus)...)
	return t
}

// WithNotes appends a nullable Notes text column.
func (t *TableDef) WithNotes() *TableDef {
	t.cols = append(t.cols, NotesColumnDefs()...)
	return t
}

// WithUUID appends a UUID column (NOT NULL, UNIQUE, immutable).
func (t *TableDef) WithUUID() *TableDef {
	t.cols = append(t.cols, UUIDColumnDefs()...)
	return t
}

// WithParent appends a nullable ParentID column for tree/hierarchy structures.
func (t *TableDef) WithParent() *TableDef {
	t.cols = append(t.cols, ParentColumnDefs()...)
	return t
}

// WithExpiry appends a nullable ExpiresAt timestamp column.
func (t *TableDef) WithExpiry() *TableDef {
	t.hasExpiry = true
	t.cols = append(t.cols, ExpiryColumnDefs()...)
	return t
}

// Indexes appends index definitions.
func (t *TableDef) Indexes(indexes ...IndexDef) *TableDef {
	t.indexes = append(t.indexes, indexes...)
	return t
}

// SelectColumns returns all column names.
func (t *TableDef) SelectColumns() []string {
	names := make([]string, len(t.cols))
	for i, c := range t.cols {
		names[i] = c.name
	}
	return names
}

// InsertColumns returns column names excluding auto-increment columns.
func (t *TableDef) InsertColumns() []string {
	var names []string
	for _, c := range t.cols {
		if !c.autoIncr {
			names = append(names, c.name)
		}
	}
	return names
}

// UpdateColumns returns only mutable column names.
func (t *TableDef) UpdateColumns() []string {
	var names []string
	for _, c := range t.cols {
		if c.mutable {
			names = append(names, c.name)
		}
	}
	return names
}

// HasSoftDelete reports whether the table uses soft-delete.
func (t *TableDef) HasSoftDelete() bool {
	return t.hasSoftDelete
}

// HasVersion reports whether the table uses optimistic concurrency control.
func (t *TableDef) HasVersion() bool {
	return t.hasVersion
}

// HasExpiry reports whether the table uses expiry.
func (t *TableDef) HasExpiry() bool {
	return t.hasExpiry
}

// CreateSQL returns the CREATE TABLE statement followed by CREATE INDEX statements.
func (t *TableDef) CreateSQL(d dialect.Dialect) []string {
	var colLines []string
	for _, c := range t.cols {
		colLines = append(colLines, "\t\t\t"+c.ddl(d))
	}

	create := fmt.Sprintf("\n\t\tCREATE TABLE %s (\n%s\n\t\t)",
		t.Name, strings.Join(colLines, ",\n"))

	stmts := []string{create}
	for _, idx := range t.indexes {
		stmts = append(stmts, d.CreateIndexIfNotExists(idx.name, t.Name, idx.columns))
	}
	return stmts
}

// DropSQL returns the DROP TABLE statement for the given dialect.
func (t *TableDef) DropSQL(d dialect.Dialect) string {
	return d.DropTableIfExists(t.Name)
}
