// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package associations

import (
	"context"
)

type Querier interface {
	// RolePermissions returns all RolePermission record(s).
	RolePermissions(ctx context.Context) ([]*RolePermission, error)
	// UserRoles returns all UserRole record(s).
	UserRoles(ctx context.Context) ([]*RolePermission, error)
}

var _ Querier = (*Queries)(nil)
