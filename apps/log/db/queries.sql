-- name: InsertLog :exec
INSERT INTO log.entries (timestamp, level, service, module, correlation_id, message, attrs)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: QueryLogs :many
SELECT * FROM log.entries
WHERE
    ($1::text = '' OR service = $1) AND
    ($2::text = '' OR level = $2) AND
    ($3::text = '' OR correlation_id = $3)
ORDER BY timestamp DESC
LIMIT $4 OFFSET $5;

-- name: CountLogs :one
SELECT COUNT(*) FROM log.entries
WHERE
    ($1::text = '' OR service = $1) AND
    ($2::text = '' OR level = $2) AND
    ($3::text = '' OR correlation_id = $3);

-- name: DeleteOldLogs :exec
DELETE FROM log.entries WHERE timestamp < $1;
