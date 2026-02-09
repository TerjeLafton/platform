-- LISTS

-- name: CreateList :one
INSERT INTO todo.lists (user_id, title)
VALUES ($1, $2)
RETURNING *;

-- name: GetListsByUser :many
SELECT
  l.id, l.user_id, l.title, l.created_at, l.updated_at,
  COUNT(i.id)::int AS total_items,
  COUNT(i.id) FILTER (WHERE i.completed)::int AS completed_items
FROM todo.lists l
LEFT JOIN todo.items i ON i.list_id = l.id
WHERE l.user_id = $1
GROUP BY l.id
ORDER BY l.title;

-- name: UpdateListTitle :one
UPDATE todo.lists
SET
    title = $3,
    updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteList :one
DELETE FROM todo.lists
WHERE id = $1 AND user_id = $2
RETURNING id;

-- ITEMS

-- name: CreateItem :one
INSERT INTO todo.items (list_id, title)
SELECT $1, $2
FROM todo.lists
WHERE id = $1 AND user_id = $3
RETURNING *;

-- name: GetAllItemsFromList :many
SELECT i.* FROM todo.items i
INNER JOIN todo.lists l ON i.list_id = l.id
WHERE i.list_id = $1 AND l.user_id = $2
ORDER BY i.created_at;

-- name: ToggleItemCompleted :one
UPDATE todo.items i
SET
    completed = NOT completed,
    updated_at = NOW()
FROM todo.lists l
WHERE i.id = $1 AND i.list_id = l.id AND l.user_id = $2
RETURNING i.*;

-- name: DeleteItem :one
DELETE FROM todo.items i
USING todo.lists l
WHERE i.id = $1 AND i.list_id = l.id AND l.user_id = $2
RETURNING i.id;
