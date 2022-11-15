// Filename: internal/data/entries.go
package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"water.biling.system.driane.perez.net/internal/validator"
)

type Todo_list struct {
	ID          int64     `json:"id"`
	CreatedAt   time.Time `json:"-"`
	Waterbill   string    `json:"waterbill"`
	Description string    `json:"description"`
	Notes       string    `json:"notes"`
	Category    string    `json:"category"`
	Priority    string    `json:"priority"`
	Status      []string  `json:"status"`
	Version     int32     `json:"version"`
}

func ValidateEntires(v *validator.Validator, entries *Todo_list) {
	//use the check method to execute our validation checks
	v.Check(entries.Waterbill != "", "waterbill", "must be provided")
	v.Check(len(entries.Waterbill) <= 200, "waterbill", "must not be more than 200 bytes long")

	v.Check(entries.Description != "", "description", "must be provided")
	v.Check(len(entries.Description) <= 200, "description", "must not be more than 200 bytes long")

	v.Check(entries.Notes != "", "notes", "must be provided")
	v.Check(len(entries.Notes) <= 200, "notes", "must not be more than 200 bytes long")

	v.Check(entries.Category != "", "category", "must be provided")
	v.Check(len(entries.Category) <= 200, "category", "must not be more than 200 bytes long")

	v.Check(entries.Priority != "", "priority", "must be provided")
	v.Check(len(entries.Priority) <= 200, "priority", "must not be more than 200 bytes long")

	v.Check(entries.Status != nil, "status", "must be provided")
	v.Check(len(entries.Status) >= 1, "status", "must contain one Status")
	v.Check(len(entries.Status) <= 5, "status", "must contain at least five Status")
	v.Check(validator.Unique(entries.Status), "status", "must not contain duplicate Status")

}

// define a todo_list model which wraps a sql.DB connection pool
type Todo_listModel struct {
	DB *sql.DB
}

// Insert() allows us to create a new todo_list
func (m Todo_listModel) Insert(Todo_list *Todo_list) error {
	query := `
	INSERT INTO water_system (waterbill, description, notes, category, priority, status)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id, created_at, version
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()

	// Collect the data fields into a slice
	args := []interface{}{
		Todo_list.Waterbill,
		Todo_list.Description,
		Todo_list.Notes,
		Todo_list.Category,
		Todo_list.Priority,
		pq.Array(Todo_list.Status),
	}
	return m.DB.QueryRowContext(ctx, query, args...).Scan(&Todo_list.ID, &Todo_list.CreatedAt, &Todo_list.Version)
}

// GET () allow us to retrieve a specific todo_list
func (m Todo_listModel) Get(id int64) (*Todo_list, error) {
	//ensure that there is a valid id
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// Create query
	query := `
		SELECT id, created_at, waterbill, description, notes, category, priority, status, version
		FROM water_system
		WHERE id = $1
	`
	// Declare a Todo_list variable to hold the return data
	var todo_list Todo_list
	//create a context
	//time starts after context is created
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()
	// Execute Query using the QueryRowContext()
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&todo_list.ID,
		&todo_list.CreatedAt,
		&todo_list.Waterbill,
		&todo_list.Description,
		&todo_list.Notes,
		&todo_list.Category,
		&todo_list.Priority,
		pq.Array(&todo_list.Status),
		&todo_list.Version,
	)
	// Handle any errors
	if err != nil {
		// Check the type of error
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	// Success
	return &todo_list, nil
}

// Update() allows us to edit/alter a specific Todolist
//optimistic locking (version number)
func (m Todo_listModel) Update(Todo_list *Todo_list) error {
	//created a query
	query := `
	UPDATE water_system 
	set waterbill = $1,
	description = $2, 
	notes = $3,
	category = $4, 
	priority = $5,
	status = $6, 
	version = version + 1
	WHERE id = $7
	AND version = $8
	RETURNING version
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()

	args := []interface{}{
		Todo_list.Waterbill,
		Todo_list.Description,
		Todo_list.Notes,
		Todo_list.Category,
		Todo_list.Priority,
		pq.Array(Todo_list.Status),
		Todo_list.ID,
		Todo_list.Version,
	}
	// Check for edit conflicts
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&Todo_list.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

// deletes() removes a specific todolist
func (m Todo_listModel) Delete(id int64) error {
	// Ensure that there is a valid id
	if id < 1 {
		return ErrRecordNotFound
	}
	// Create the delete query
	query := `
		DELETE FROM water_system
		WHERE id = $1
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()
	// Execute the query
	results, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	// Check how many rows were affected by the delete operations. We
	// call the RowsAffected() method on the result variable
	rowsAffected, err := results.RowsAffected()
	if err != nil {
		return err
	}
	// Check if no rows were affected
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

// The GetAll() returns a list of all the todo items sorted by ID
func (m Todo_listModel) GetAll(waterbill string, priority string, status []string, filters Filters) ([]*Todo_list, Metadata, error) {
	// Construct the query
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), 
		id, created_at, 
		waterbill, description, 
		notes, category, 
		priority, 
		status, 
		version
		FROM water_system
		WHERE (to_tsvector('simple',waterbill) @@ plainto_tsquery('simple', $1) OR $1 = '')
		AND (to_tsvector('simple',priority) @@ plainto_tsquery('simple', $2) OR $2 = '')
		AND (status @> $3 OR $3 = '{}')
		ORDER BY %s %s, id ASC
		LIMIT $4 OFFSET $5`, filters.sortColumn(), filters.sortOrder())

	// Create a 3-second-timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	args := []interface{}{waterbill, priority, pq.Array(status), filters.limit(), filters.offset()}
	// Execute query
	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	// Close the result set
	defer rows.Close()
	totalRecords := 0
	// Initialize an empty slice to hold the todo_list data
	todo_listD := []*Todo_list{}
	// Iterate over the rows in the results set
	for rows.Next() {
		var todo_List Todo_list
		// Scan the values from the row in to the Todo_list struct
		err := rows.Scan(
			&totalRecords,
			&todo_List.ID,
			&todo_List.CreatedAt,
			&todo_List.Waterbill,
			&todo_List.Description,
			&todo_List.Notes,
			&todo_List.Category,
			&todo_List.Priority,
			pq.Array(&todo_List.Status),
			&todo_List.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		// Add the Todo to our slice
		todo_listD = append(todo_listD, &todo_List)
	}
	// Check for errors after looping through the results set
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	// Return the slice of Todo_list
	return todo_listD, metadata, nil
}