// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: queries.sql

package overrides

import (
	"context"
)

const userPermissions = `-- name: UserPermissions :many
SELECT id, "user-id", "permission-id" FROM "User-Permission"
`

// UserPermissions returns all UserPermission record(s).
func (q *Queries) UserPermissions(ctx context.Context) ([]*UserPermission, error) {
	rows, err := q.db.Query(ctx, userPermissions)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*UserPermission{}
	for rows.Next() {
		var i UserPermission
		if err := rows.Scan(&i.ID, &i.UserID, &i.PermissionID); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}