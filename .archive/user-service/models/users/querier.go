// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package users

import (
	"context"
)

type Querier interface {
	// All returns the total number of User records, including deleted record(s).
	All(ctx context.Context, db DBTX) (int64, error)
	// Attributes will use the user's [User.ID] to hydrate all available User attribute(s). Note that the following call is more taxing on the database.
	Attributes(ctx context.Context, db DBTX, id int64) (User, error)
	// Count returns 0 or 1 depending on if a User record matching the provided email exists.
	Count(ctx context.Context, db DBTX, email string) (int64, error)
	Create(ctx context.Context, db DBTX, arg *CreateParams) (User, error)
	// List returns all active User record(s).
	List(ctx context.Context, db DBTX) ([]User, error)
	// Total returns the total number of User records, excluding deleted record(s).
	Total(ctx context.Context, db DBTX) (int64, error)
	UpdateUserAvatar(ctx context.Context, db DBTX, arg *UpdateUserAvatarParams) error
	// Users returns all User record(s).
	Users(ctx context.Context, db DBTX) ([]User, error)
}

var _ Querier = (*Queries)(nil)
