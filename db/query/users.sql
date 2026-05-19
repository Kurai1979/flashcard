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