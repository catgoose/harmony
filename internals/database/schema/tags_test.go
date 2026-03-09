package schema

import (
	"testing"

	"catgoose/go-htmx-demo/internals/database/dialect"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTagsTable(t *testing.T) {
	td := NewTagsTable("Tags")

	cols := td.SelectColumns()
	assert.Equal(t, []string{"ID", "Type", "Label"}, cols)

	// ID excluded from insert
	insert := td.InsertColumns()
	assert.Equal(t, []string{"Type", "Label"}, insert)

	d := dialect.SQLiteDialect{}
	stmts := td.CreateSQL(d)
	require.Len(t, stmts, 3) // CREATE TABLE + 2 indexes

	assert.Contains(t, stmts[0], "CREATE TABLE Tags")
	assert.Contains(t, stmts[0], "Type TEXT NOT NULL")
	assert.Contains(t, stmts[0], "Label TEXT NOT NULL")
	assert.Contains(t, stmts[1], "idx_Tags_type")
	assert.Contains(t, stmts[2], "idx_Tags_type_label")
}

func TestNewTagsTable_MSSQL(t *testing.T) {
	td := NewTagsTable("Lookups")

	d := dialect.MSSQLDialect{}
	stmts := td.CreateSQL(d)
	require.Len(t, stmts, 3)

	assert.Contains(t, stmts[0], "Type VARCHAR(100) NOT NULL")
	assert.Contains(t, stmts[0], "Label VARCHAR(255) NOT NULL")
}

func TestNewTagJoinTable(t *testing.T) {
	td := NewTagJoinTable("ItemTags", "Items", "Tags")

	cols := td.SelectColumns()
	assert.Equal(t, []string{"OwnerID", "TagID"}, cols)

	// Both columns are immutable
	assert.Empty(t, td.UpdateColumns())

	// Both included in insert (no auto-increment)
	assert.Equal(t, []string{"OwnerID", "TagID"}, td.InsertColumns())

	d := dialect.SQLiteDialect{}
	stmts := td.CreateSQL(d)
	require.Len(t, stmts, 3) // CREATE TABLE + 2 indexes

	assert.Contains(t, stmts[0], "OwnerID INTEGER NOT NULL")
	assert.Contains(t, stmts[0], "TagID INTEGER NOT NULL")
	assert.Contains(t, stmts[1], "idx_ItemTags_ownerid")
	assert.Contains(t, stmts[2], "idx_ItemTags_tagid")
}
