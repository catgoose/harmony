package schema

import "fmt"

// NewTagsTable creates a lookup-style tags table with ID, Type, and Label columns.
// The Type column allows a single table to serve multiple tag categories
// (e.g., "priority", "status", "category").
func NewTagsTable(name string) *TableDef {
	return NewTable(name).
		Columns(
			AutoIncrCol("ID"),
			Col("Type", TypeVarchar(100)).NotNull(),
			Col("Label", TypeVarchar(255)).NotNull(),
		).
		Indexes(
			Index(fmt.Sprintf("idx_%s_type", name), "Type"),
			Index(fmt.Sprintf("idx_%s_type_label", name), "Type, Label"),
		)
}

// NewTagJoinTable creates a many-to-many join table linking an owner table to a tags table.
// It creates a composite primary key on (OwnerID, TagID).
func NewTagJoinTable(name, ownerTable, tagsTable string) *TableDef {
	return NewTable(name).
		Columns(
			Col("OwnerID", TypeInt()).NotNull().Immutable(),
			Col("TagID", TypeInt()).NotNull().Immutable(),
		).
		Indexes(
			Index(fmt.Sprintf("idx_%s_ownerid", name), "OwnerID"),
			Index(fmt.Sprintf("idx_%s_tagid", name), "TagID"),
		)
}
