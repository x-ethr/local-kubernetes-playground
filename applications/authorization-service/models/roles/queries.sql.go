// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: queries.sql

package roles

import (
	"context"
)

const roles = `-- name: Roles :many
SELECT id, name, description, creation, modification, deletion FROM "Role"
`

// Roles returns all Role record(s).
func (q *Queries) Roles(ctx context.Context) ([]*Role, error) {
	rows, err := q.db.Query(ctx, roles)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*Role{}
	for rows.Next() {
		var i Role
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