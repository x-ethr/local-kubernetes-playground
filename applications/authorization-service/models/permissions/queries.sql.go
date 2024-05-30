// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: queries.sql

package permissions

import (
	"context"
)

const permissions = `-- name: Permissions :many
SELECT id, name, description, creation, modification, deletion FROM "Permission"
`

// Permissions returns all Permission record(s).
func (q *Queries) Permissions(ctx context.Context) ([]*Permission, error) {
	rows, err := q.db.Query(ctx, permissions)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*Permission{}
	for rows.Next() {
		var i Permission
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.Creation,
			&i.Modification,
			&i.Deletion,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}