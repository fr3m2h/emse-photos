-- name: AttemptCreatingUser :exec
INSERT INTO users (email, full_name, business_category, department_number)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE user_id = LAST_INSERT_ID(user_id);

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

-- name: GetUserWithSession :one
SELECT u.*
FROM users u
JOIN sessions s
ON s.user_id = u.user_id
WHERE s.session_token = ?;




-- name: CreateSession :exec
INSERT INTO sessions (user_id, session_token)
VALUES (?, ?);

-- name: GetSessionWithToken :one
SELECT *
FROM sessions
WHERE session_token = ?;

-- name: DeleteSessionWithToken :exec
DELETE FROM sessions WHERE session_token = ?;




-- name: CreateEvent :exec
INSERT INTO events (name, description, event_date, parent_event_id)
VALUES (?, ?, ?, ?);

-- name: GetEvents :many
SELECT name, description, event_date, creation_date, parent_event_id
FROM events;

-- name: UpdateEvent :exec
UPDATE events
SET name = ?, description = ?, event_date = ?, parent_event_id = ?
WHERE event_id = ?;

-- name: DeleteEvent :exec
DELETE FROM events WHERE event_id = ?;




-- name: CreatePhoto :exec
INSERT INTO photos (path_to_photo, event_id)
VALUES (?, ?);

-- name: GetPhoto :one
SELECT * FROM photos WHERE photo_id = ?;

-- name: GetPhotosByEventID :many
SELECT * FROM photos WHERE event_id = ?;

-- name: GetPhotosSortedByDate :many
SELECT * FROM photos ORDER BY creation_date DESC;

-- name: UpdatePhotoPath :exec
UPDATE photos
SET path_to_photo = ?
WHERE photo_id = ?;

-- name: DeletePhoto :exec
DELETE FROM photos WHERE photo_id = ?;
