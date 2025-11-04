-- name: GetUserFromEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: GetUserFromUserID :one
SELECT * FROM users
WHERE id = $1;
