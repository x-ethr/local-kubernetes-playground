-- name: Attributes :one
-- Attributes will use the user's [User.ID] to hydrate all available User attribute(s). Note that the following call is more taxing on the database.
SELECT *
FROM "User"
WHERE (id) = sqlc.arg(id)::bigserial
  AND (deletion) IS NULL
LIMIT 1;

-- name: Create :one
INSERT INTO "User" (email, password) VALUES ($1, $2) RETURNING id, email;

-- name: Get :one
SELECT email, password FROM "User" WHERE email = $1;

-- name: Count :one
-- Count returns 0 or 1 depending on if a User record matching the provided email exists.
SELECT count(*) FROM "User" WHERE (email) = sqlc.arg(email)::text AND (deletion) IS NULL;

-- name: Total :one
-- Total returns the total number of User records, excluding deleted record(s).
SELECT count(*) FROM "User" WHERE (deletion) IS NULL;

-- name: All :one
-- All returns the total number of User records, including deleted record(s).
SELECT count(*) FROM "User";

-- name: List :many
-- List returns all active User record(s).
SELECT * FROM "User" WHERE (deletion) IS NULL;

-- name: Users :many
-- Users returns all User record(s).
SELECT * FROM "User";
