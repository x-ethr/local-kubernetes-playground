// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: queries.sql

package users

import (
	"context"
)

const all = `-- name: All :one
SELECT count(*) FROM "User"
`

// All returns the total number of User records, including deleted record(s).
func (q *Queries) All(ctx context.Context, db DBTX) (int64, error) {
	row := db.QueryRow(ctx, all)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const attributes = `-- name: Attributes :one
SELECT id, name, "display-name", "account-type", email, username, avatar, "verification-status", marketing, creation, modification, deletion
FROM "User"
WHERE (id) = $1::bigserial
  AND (deletion) IS NULL
LIMIT 1
`

// Attributes will use the user's [User.ID] to hydrate all available User attribute(s). Note that the following call is more taxing on the database.
func (q *Queries) Attributes(ctx context.Context, db DBTX, id int64) (User, error) {
	row := db.QueryRow(ctx, attributes, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.DisplayName,
		&i.AccountType,
		&i.Email,
		&i.Username,
		&i.Avatar,
		&i.VerificationStatus,
		&i.Marketing,
		&i.Creation,
		&i.Modification,
		&i.Deletion,
	)
	return i, err
}

const count = `-- name: Count :one
SELECT count(*) FROM "User" WHERE (email) = $1::text AND (deletion) IS NULL
`

// Count returns 0 or 1 depending on if a User record matching the provided email exists.
func (q *Queries) Count(ctx context.Context, db DBTX, email string) (int64, error) {
	row := db.QueryRow(ctx, count, email)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const create = `-- name: Create :one
INSERT INTO "User" (email, avatar) VALUES ($1, $2) RETURNING id, name, "display-name", "account-type", email, username, avatar, "verification-status", marketing, creation, modification, deletion
`

type CreateParams struct {
	Email  string  `db:"email" json:"email"`
	Avatar *string `db:"avatar" json:"avatar"`
}

func (q *Queries) Create(ctx context.Context, db DBTX, arg *CreateParams) (User, error) {
	row := db.QueryRow(ctx, create, arg.Email, arg.Avatar)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.DisplayName,
		&i.AccountType,
		&i.Email,
		&i.Username,
		&i.Avatar,
		&i.VerificationStatus,
		&i.Marketing,
		&i.Creation,
		&i.Modification,
		&i.Deletion,
	)
	return i, err
}

const list = `-- name: List :many
SELECT id, name, "display-name", "account-type", email, username, avatar, "verification-status", marketing, creation, modification, deletion FROM "User" WHERE (deletion) IS NULL
`

// List returns all active User record(s).
func (q *Queries) List(ctx context.Context, db DBTX) ([]User, error) {
	rows, err := db.Query(ctx, list)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []User{}
	for rows.Next() {
		var i User
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.DisplayName,
			&i.AccountType,
			&i.Email,
			&i.Username,
			&i.Avatar,
			&i.VerificationStatus,
			&i.Marketing,
			&i.Creation,
			&i.Modification,
			&i.Deletion,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const total = `-- name: Total :one
SELECT count(*) FROM "User" WHERE (deletion) IS NULL
`

// Total returns the total number of User records, excluding deleted record(s).
func (q *Queries) Total(ctx context.Context, db DBTX) (int64, error) {
	row := db.QueryRow(ctx, total)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const updateUserAvatar = `-- name: UpdateUserAvatar :exec
UPDATE "User" SET avatar = $2 WHERE (email) = $1 AND (deletion) IS NULL
`

type UpdateUserAvatarParams struct {
	Email  string  `db:"email" json:"email"`
	Avatar *string `db:"avatar" json:"avatar"`
}

func (q *Queries) UpdateUserAvatar(ctx context.Context, db DBTX, arg *UpdateUserAvatarParams) error {
	_, err := db.Exec(ctx, updateUserAvatar, arg.Email, arg.Avatar)
	return err
}

const users = `-- name: Users :many
SELECT id, name, "display-name", "account-type", email, username, avatar, "verification-status", marketing, creation, modification, deletion FROM "User"
`

// Users returns all User record(s).
func (q *Queries) Users(ctx context.Context, db DBTX) ([]User, error) {
	rows, err := db.Query(ctx, users)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []User{}
	for rows.Next() {
		var i User
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.DisplayName,
			&i.AccountType,
			&i.Email,
			&i.Username,
			&i.Avatar,
			&i.VerificationStatus,
			&i.Marketing,
			&i.Creation,
			&i.Modification,
			&i.Deletion,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
