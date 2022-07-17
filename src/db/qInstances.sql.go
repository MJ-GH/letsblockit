// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.14.0
// source: qInstances.sql

package db

import (
	"context"

	"github.com/jackc/pgtype"
)

const countInstanceForUserAndFilter = `-- name: CountInstanceForUserAndFilter :one
SELECT COUNT(*)
FROM filter_instances
WHERE (user_id = $1 AND filter_name = $2)
`

type CountInstanceForUserAndFilterParams struct {
	UserID     string
	FilterName string
}

func (q *Queries) CountInstanceForUserAndFilter(ctx context.Context, arg CountInstanceForUserAndFilterParams) (int64, error) {
	row := q.db.QueryRow(ctx, countInstanceForUserAndFilter, arg.UserID, arg.FilterName)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const createInstanceForUserAndFilter = `-- name: CreateInstanceForUserAndFilter :exec
INSERT INTO filter_instances (filter_list_id, user_id, filter_name, params, test_mode)
VALUES ((SELECT id FROM filter_lists WHERE user_id = $1), $1, $2, $3, $4)
`

type CreateInstanceForUserAndFilterParams struct {
	UserID     string
	FilterName string
	Params     pgtype.JSONB
	TestMode   bool
}

func (q *Queries) CreateInstanceForUserAndFilter(ctx context.Context, arg CreateInstanceForUserAndFilterParams) error {
	_, err := q.db.Exec(ctx, createInstanceForUserAndFilter,
		arg.UserID,
		arg.FilterName,
		arg.Params,
		arg.TestMode,
	)
	return err
}

const deleteInstanceForUserAndFilter = `-- name: DeleteInstanceForUserAndFilter :exec
DELETE
FROM filter_instances
WHERE (user_id = $1 AND filter_name = $2)
`

type DeleteInstanceForUserAndFilterParams struct {
	UserID     string
	FilterName string
}

func (q *Queries) DeleteInstanceForUserAndFilter(ctx context.Context, arg DeleteInstanceForUserAndFilterParams) error {
	_, err := q.db.Exec(ctx, deleteInstanceForUserAndFilter, arg.UserID, arg.FilterName)
	return err
}

const getActiveFiltersForUser = `-- name: GetActiveFiltersForUser :many
SELECT filter_name, params, test_mode
FROM filter_instances
WHERE user_id = $1
`

type GetActiveFiltersForUserRow struct {
	FilterName string
	Params     pgtype.JSONB
	TestMode   bool
}

func (q *Queries) GetActiveFiltersForUser(ctx context.Context, userID string) ([]GetActiveFiltersForUserRow, error) {
	rows, err := q.db.Query(ctx, getActiveFiltersForUser, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetActiveFiltersForUserRow
	for rows.Next() {
		var i GetActiveFiltersForUserRow
		if err := rows.Scan(&i.FilterName, &i.Params, &i.TestMode); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getInstanceForUserAndFilter = `-- name: GetInstanceForUserAndFilter :one
SELECT params, test_mode
FROM filter_instances
WHERE (user_id = $1 AND filter_name = $2)
`

type GetInstanceForUserAndFilterParams struct {
	UserID     string
	FilterName string
}

type GetInstanceForUserAndFilterRow struct {
	Params   pgtype.JSONB
	TestMode bool
}

func (q *Queries) GetInstanceForUserAndFilter(ctx context.Context, arg GetInstanceForUserAndFilterParams) (GetInstanceForUserAndFilterRow, error) {
	row := q.db.QueryRow(ctx, getInstanceForUserAndFilter, arg.UserID, arg.FilterName)
	var i GetInstanceForUserAndFilterRow
	err := row.Scan(&i.Params, &i.TestMode)
	return i, err
}

const getInstancesForList = `-- name: GetInstancesForList :many
SELECT filter_name, params, test_mode
FROM filter_instances
WHERE filter_list_id = $1
ORDER BY filter_name ASC
`

type GetInstancesForListRow struct {
	FilterName string
	Params     pgtype.JSONB
	TestMode   bool
}

func (q *Queries) GetInstancesForList(ctx context.Context, filterListID int32) ([]GetInstancesForListRow, error) {
	rows, err := q.db.Query(ctx, getInstancesForList, filterListID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetInstancesForListRow
	for rows.Next() {
		var i GetInstancesForListRow
		if err := rows.Scan(&i.FilterName, &i.Params, &i.TestMode); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateInstanceForUserAndFilter = `-- name: UpdateInstanceForUserAndFilter :exec
UPDATE filter_instances
SET params     = $3,
    test_mode  = $4,
    updated_at = NOW()
WHERE (user_id = $1 AND filter_name = $2)
`

type UpdateInstanceForUserAndFilterParams struct {
	UserID     string
	FilterName string
	Params     pgtype.JSONB
	TestMode   bool
}

func (q *Queries) UpdateInstanceForUserAndFilter(ctx context.Context, arg UpdateInstanceForUserAndFilterParams) error {
	_, err := q.db.Exec(ctx, updateInstanceForUserAndFilter,
		arg.UserID,
		arg.FilterName,
		arg.Params,
		arg.TestMode,
	)
	return err
}
