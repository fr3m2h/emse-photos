-- name: GetUser :one
SELECT *
FROM users
WHERE user_id = ?;

-- name: GetUserWithEmail :one
SELECT *
FROM users
WHERE email = ?;


-- name: GetUserWithSession :one
SELECT u.*, s.*
FROM users u
JOIN sessions s
ON s.user_id = u.user_id
WHERE s.session_token = ?;
