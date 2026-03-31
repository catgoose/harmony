// setup:feature:graph

package domain

import (
	"database/sql"
	"time"
)

// Timestamps provides CreatedAt and UpdatedAt fields for embedding in domain models.
type Timestamps struct {
	CreatedAt time.Time `db:"CreatedAt" json:"createdAt"`
	UpdatedAt time.Time `db:"UpdatedAt" json:"updatedAt"`
}

// SoftDelete provides a nullable DeletedAt field for embedding in domain models.
type SoftDelete struct {
	DeletedAt sql.NullTime `db:"DeletedAt" json:"deletedAt,omitzero"`
}

// Version provides an optimistic concurrency control field for embedding in domain models.
type Version struct {
	Version int `db:"Version" json:"version"`
}

// SortOrder provides a manual ordering field for embedding in domain models.
type SortOrder struct {
	SortOrder int `db:"SortOrder" json:"sortOrder"`
}

// Status provides a status field for embedding in domain models.
type Status struct {
	Status string `db:"Status" json:"status"`
}

// Notes provides a nullable notes field for embedding in domain models.
type Notes struct {
	Notes sql.NullString `db:"Notes" json:"notes,omitzero"`
}

// Archive provides a nullable ArchivedAt timestamp for embedding in domain models.
// Semantically softer than SoftDelete — archived records are hidden from default views
// but remain accessible and restorable.
type Archive struct {
	ArchivedAt sql.NullTime `db:"ArchivedAt" json:"archivedAt,omitzero"`
}

// Replacement provides a nullable reference to the entity that replaced this one.
type Replacement struct {
	ReplacedByID sql.NullInt64 `db:"ReplacedByID" json:"replacedById,omitzero"`
}

// ToNullString converts a string to sql.NullString. Empty strings are treated as null.
func ToNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
