-- name: GetUserByEmail :one
SELECT id,
       email,
       password_hash,
       is_active,
       is_verified,
       created_at,
       updated_at,
       last_login_at
FROM users
WHERE email = $1;


-- name: GetUserById :one
SELECT id,
       email,
       password_hash,
       is_active,
       is_verified,
       created_at,
       updated_at,
       last_login_at
FROM users
WHERE id = $1;

-- name: CreateUser :one
INSERT INTO users (email, password_hash, is_active, is_verified)
VALUES ($1, $2, $3, $4)
RETURNING id, email, password_hash, is_active, is_verified, created_at, updated_at, last_login_at;

-- name: UpdateLastLogin :exec
UPDATE users
SET last_login_at = NOW(),
    updated_at    = NOW()
WHERE id = $1;
