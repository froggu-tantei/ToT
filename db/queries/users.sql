-- name: CreateUser :one
INSERT INTO users (email, password_hash, username, profile_picture, bio)
VALUES (
  $1,
  $2,
  $3,
  $4,
  $5
)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = $1;

-- name: UpdateUser :one
UPDATE users
SET email = $2,
    password_hash = $3,
    updated_at = NOW(),
    username = $4,
    bio = $5,
    profile_picture = $6
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: GetLeaderBoard :many
SELECT id, username, last_place_count, profile_picture, bio
FROM users
ORDER BY last_place_count DESC
LIMIT $1 OFFSET $2;

-- name: IncrementLastPlaceCount :one
UPDATE users
SET last_place_count = last_place_count + 1, updated_at = NOW()
WHERE id = $1
RETURNING *;