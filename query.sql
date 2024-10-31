-- name: GetUser :one
SELECT *
FROM users
WHERE user_id = ?;

-- name: GetUserWithEmail :one
SELECT *
FROM users
WHERE email = ?;

-- name: GetUserLastInsertID :one
SELECT * FROM users WHERE user_id = LAST_INSERT_ID();

-- name: AttemptCreatingUser :exec
INSERT INTO users (email, full_name, business_category, department_number)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE user_id = LAST_INSERT_ID(user_id);

-- name: CreateSession :exec
INSERT INTO sessions (user_id, session_token)
VALUES (?, ?);

-- name: DeleteSessionWithToken :exec
DELETE FROM sessions WHERE session_token = ?;

-- name: GetSessionWithToken :one
SELECT *
FROM sessions
WHERE session_token = ?;

-- name: GetUserWithSession :one
SELECT u.*
FROM users u
JOIN sessions s
ON s.user_id = u.user_id
WHERE s.session_token = ?;
