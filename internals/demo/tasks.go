// setup:feature:demo

package demo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"catgoose/harmony/internals/database/dialect"
	dbrepo "catgoose/harmony/internals/database/repository"
	"catgoose/harmony/internals/database/schema"
)

// TaskStatuses is the list of valid task statuses.
var TaskStatuses = []string{"draft", "active", "done"}

// TasksTable defines the tasks table schema using the full repository pattern.
var TasksTable = schema.NewTable("Tasks").
	Columns(
		schema.AutoIncrCol("ID"),
		schema.Col("Title", schema.TypeString(255)).NotNull(),
		schema.Col("Description", schema.TypeText()),
	).
	WithStatus("draft").
	WithSortOrder().
	WithVersion().
	WithNotes().
	WithArchive().
	WithReplacement().
	WithTimestamps().
	WithSoftDelete().
	WithSeedRows(
		schema.SeedRow{"Title": "'Design schema builder'", "Description": "'Build composable DDL API with traits'", "Status": "'done'", "SortOrder": "1", "Version": "3"},
		schema.SeedRow{"Title": "'Implement CRUD handlers'", "Description": "'Create REST endpoints with HTMX responses'", "Status": "'active'", "SortOrder": "2", "Version": "2"},
		schema.SeedRow{"Title": "'Add filtering and search'", "Description": "'WhereBuilder with composable filters'", "Status": "'active'", "SortOrder": "3", "Version": "1"},
		schema.SeedRow{"Title": "'Write unit tests'", "Description": "'Test all repository helpers and schema builder'", "Status": "'draft'", "SortOrder": "4", "Version": "1"},
		schema.SeedRow{"Title": "'Setup pagination'", "Description": "'SelectBuilder with LIMIT/OFFSET support'", "Status": "'draft'", "SortOrder": "5", "Version": "1"},
		schema.SeedRow{"Title": "'Configure CI pipeline'", "Description": "'GitHub Actions with lint, test, build'", "Status": "'done'", "SortOrder": "6", "Version": "4"},
		schema.SeedRow{"Title": "'Add soft delete support'", "Description": "'DeletedAt timestamp with NotDeleted filter'", "Status": "'active'", "SortOrder": "7", "Version": "1"},
		schema.SeedRow{"Title": "'Implement archive flow'", "Description": "'ArchivedAt for preserving historical snapshots'", "Status": "'draft'", "SortOrder": "8", "Version": "1"},
	).
	Indexes(
		schema.Index("idx_tasks_status", "Status"),
		schema.Index("idx_tasks_sortorder", "SortOrder"),
	)

// Task represents a single task row using repository domain patterns.
type Task struct {
	ID          int            `db:"ID"`
	Title       string         `db:"Title"`
	Description sql.NullString `db:"Description"`
	Status      string         `db:"Status"`
	SortOrder   int            `db:"SortOrder"`
	Version     int            `db:"Version"`
	Notes       sql.NullString `db:"Notes"`
	ArchivedAt  sql.NullTime   `db:"ArchivedAt"`
	ReplacedBy  sql.NullInt64  `db:"ReplacedByID"`
	CreatedAt   time.Time      `db:"CreatedAt"`
	UpdatedAt   time.Time      `db:"UpdatedAt"`
	DeletedAt   sql.NullTime   `db:"DeletedAt"`
}

// TaskStore provides CRUD operations on the Tasks table using the repository pattern.
type TaskStore struct {
	db      *sql.DB
	dialect dialect.Dialect
}

// NewTaskStore creates a TaskStore.
func NewTaskStore(db *sql.DB) *TaskStore {
	return &TaskStore{db: db, dialect: dialect.SQLiteDialect{}}
}

// ListTasks queries tasks with optional filters, sort, and pagination.
func (s *TaskStore) ListTasks(ctx context.Context, search, status, showArchived, showDeleted, sortBy, sortDir string, page, perPage int) ([]Task, int, error) {
	w := dbrepo.NewWhere()

	if showDeleted != "true" {
		w.NotDeleted()
	}
	if showArchived != "true" {
		w.NotArchived()
	}
	if status != "" {
		w.HasStatus(status)
	}
	if search != "" {
		w.Search(search, "Title", "Description")
	}

	cols := dbrepo.Columns(TasksTable.SelectColumns()...)

	// Sort
	colMap := map[string]string{
		"title":     "Title",
		"status":    "Status",
		"sortorder": "SortOrder",
		"version":   "Version",
		"created":   "CreatedAt",
		"updated":   "UpdatedAt",
	}

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * perPage

	sb := dbrepo.NewSelect(TasksTable.Name, cols).
		Where(w).
		OrderByMap(sortBy+":"+sortDir, colMap, "SortOrder ASC").
		Paginate(perPage, offset).
		WithDialect(s.dialect)

	query, args := sb.Build()
	countQuery, countArgs := sb.CountQuery()

	// Count
	var total int
	if err := s.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count tasks: %w", err)
	}

	// Query
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(
			&t.ID, &t.Title, &t.Description,
			&t.Status, &t.SortOrder, &t.Version, &t.Notes,
			&t.ArchivedAt, &t.ReplacedBy,
			&t.CreatedAt, &t.UpdatedAt, &t.DeletedAt,
		); err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, t)
	}
	return tasks, total, rows.Err()
}

// GetTask returns a single task by ID.
func (s *TaskStore) GetTask(ctx context.Context, id int) (Task, error) {
	cols := dbrepo.Columns(TasksTable.SelectColumns()...)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE ID = ?", cols, TasksTable.Name)

	var t Task
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID, &t.Title, &t.Description,
		&t.Status, &t.SortOrder, &t.Version, &t.Notes,
		&t.ArchivedAt, &t.ReplacedBy,
		&t.CreatedAt, &t.UpdatedAt, &t.DeletedAt,
	)
	if err != nil {
		return Task{}, fmt.Errorf("get task %d: %w", id, err)
	}
	return t, nil
}

// CreateTask inserts a new task using repository helpers.
func (s *TaskStore) CreateTask(ctx context.Context, t *Task) error {
	dbrepo.SetCreateTimestamps(&t.CreatedAt, &t.UpdatedAt)
	dbrepo.InitVersion(&t.Version)
	if t.Status == "" {
		dbrepo.SetStatus(&t.Status, "draft")
	}

	insertCols := TasksTable.InsertColumns()
	query := dbrepo.InsertInto(TasksTable.Name, insertCols...)

	res, err := s.db.ExecContext(ctx, query,
		sql.Named("Title", t.Title),
		sql.Named("Description", t.Description),
		sql.Named("Status", t.Status),
		sql.Named("SortOrder", t.SortOrder),
		sql.Named("Version", t.Version),
		sql.Named("Notes", t.Notes),
		sql.Named("ArchivedAt", t.ArchivedAt),
		sql.Named("ReplacedByID", t.ReplacedBy),
		sql.Named("CreatedAt", t.CreatedAt),
		sql.Named("UpdatedAt", t.UpdatedAt),
		sql.Named("DeletedAt", t.DeletedAt),
	)
	if err != nil {
		return fmt.Errorf("create task: %w", err)
	}
	id, _ := res.LastInsertId()
	t.ID = int(id)
	return nil
}

// UpdateTask updates a task using repository helpers.
func (s *TaskStore) UpdateTask(ctx context.Context, t *Task) error {
	dbrepo.SetUpdateTimestamp(&t.UpdatedAt)
	dbrepo.IncrementVersion(&t.Version)

	updateCols := TasksTable.UpdateColumns()
	query := fmt.Sprintf("UPDATE %s SET %s WHERE ID = ? AND Version = ?",
		TasksTable.Name,
		dbrepo.SetClause(updateCols...),
	)

	res, err := s.db.ExecContext(ctx, query,
		sql.Named("Title", t.Title),
		sql.Named("Description", t.Description),
		sql.Named("Status", t.Status),
		sql.Named("SortOrder", t.SortOrder),
		sql.Named("Version", t.Version),
		sql.Named("Notes", t.Notes),
		sql.Named("ArchivedAt", t.ArchivedAt),
		sql.Named("ReplacedByID", t.ReplacedBy),
		sql.Named("UpdatedAt", t.UpdatedAt),
		sql.Named("DeletedAt", t.DeletedAt),
		t.ID, t.Version-1, // optimistic lock: WHERE Version = previous version
	)
	if err != nil {
		return fmt.Errorf("update task %d: %w", t.ID, err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("task %d: version conflict or not found", t.ID)
	}
	return nil
}

// SoftDeleteTask marks a task as deleted using SetSoftDelete.
func (s *TaskStore) SoftDeleteTask(ctx context.Context, id int) error {
	t, err := s.GetTask(ctx, id)
	if err != nil {
		return err
	}
	deletedAt := t.DeletedAt.Time
	dbrepo.SetSoftDelete(&deletedAt)
	_, err = s.db.ExecContext(ctx,
		fmt.Sprintf("UPDATE %s SET DeletedAt = ?, UpdatedAt = ? WHERE ID = ?", TasksTable.Name),
		deletedAt, dbrepo.GetNow(), id,
	)
	return err
}

// RestoreTask clears the DeletedAt timestamp.
func (s *TaskStore) RestoreTask(ctx context.Context, id int) error {
	_, err := s.db.ExecContext(ctx,
		fmt.Sprintf("UPDATE %s SET DeletedAt = NULL, UpdatedAt = ? WHERE ID = ?", TasksTable.Name),
		dbrepo.GetNow(), id,
	)
	return err
}

// ArchiveTask sets ArchivedAt using the archive helper.
func (s *TaskStore) ArchiveTask(ctx context.Context, id int) error {
	now := dbrepo.GetNow()
	archivedAt := now
	dbrepo.SetArchive(&archivedAt)
	_, err := s.db.ExecContext(ctx,
		fmt.Sprintf("UPDATE %s SET ArchivedAt = ?, UpdatedAt = ? WHERE ID = ?", TasksTable.Name),
		archivedAt, now, id,
	)
	return err
}

// UnarchiveTask clears ArchivedAt using ClearArchive.
func (s *TaskStore) UnarchiveTask(ctx context.Context, id int) error {
	var archivedAt time.Time
	dbrepo.ClearArchive(&archivedAt)
	_, err := s.db.ExecContext(ctx,
		fmt.Sprintf("UPDATE %s SET ArchivedAt = NULL, UpdatedAt = ? WHERE ID = ?", TasksTable.Name),
		dbrepo.GetNow(), id,
	)
	return err
}

// initTasks creates the tasks table and seeds it if empty.
func (d *DB) initTasks() error {
	sqliteDialect := dialect.SQLiteDialect{}
	for _, stmt := range TasksTable.CreateIfNotExistsSQL(sqliteDialect) {
		if _, err := d.db.Exec(stmt); err != nil {
			return fmt.Errorf("create tasks table: %w", err)
		}
	}

	var count int
	if err := d.db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", TasksTable.Name)).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		for _, stmt := range TasksTable.SeedSQL() {
			if _, err := d.db.Exec(stmt); err != nil {
				return fmt.Errorf("seed tasks: %w", err)
			}
		}
	}
	return nil
}
