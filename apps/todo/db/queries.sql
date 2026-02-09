-- LISTS

-- name: CreateList :one
INSERT INTO todo.lists (user_id, title)
VALUES ($1, $2)
RETURNING *;

-- name: GetListsByUser :many
SELECT
  l.id, l.user_id, l.title, l.created_at, l.updated_at,
  COUNT(i.id)::int AS total_items,
  COUNT(i.id) FILTER (WHERE i.completed)::int AS completed_items,
  COALESCE(owner.name, '') AS owner_name,
  (l.user_id = $1) AS is_owner
FROM todo.lists l
LEFT JOIN todo.items i ON i.list_id = l.id
LEFT JOIN id.users owner ON owner.id = l.user_id
WHERE l.user_id = $1
   OR EXISTS (SELECT 1 FROM todo.list_members lm WHERE lm.list_id = l.id AND lm.user_id = $1)
GROUP BY l.id, owner.name
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
FROM todo.lists l
WHERE l.id = $1
  AND (l.user_id = $3 OR EXISTS (SELECT 1 FROM todo.list_members lm WHERE lm.list_id = l.id AND lm.user_id = $3))
RETURNING *;

-- name: GetAllItemsFromList :many
SELECT i.* FROM todo.items i
INNER JOIN todo.lists l ON i.list_id = l.id
WHERE i.list_id = $1
  AND (l.user_id = $2 OR EXISTS (SELECT 1 FROM todo.list_members lm WHERE lm.list_id = l.id AND lm.user_id = $2))
ORDER BY i.created_at;

-- name: ToggleItemCompleted :one
UPDATE todo.items i
SET
    completed = NOT completed,
    updated_at = NOW()
FROM todo.lists l
WHERE i.id = $1 AND i.list_id = l.id
  AND (l.user_id = $2 OR EXISTS (SELECT 1 FROM todo.list_members lm WHERE lm.list_id = l.id AND lm.user_id = $2))
RETURNING i.*;

-- name: DeleteItem :one
DELETE FROM todo.items i
USING todo.lists l
WHERE i.id = $1 AND i.list_id = l.id
  AND (l.user_id = $2 OR EXISTS (SELECT 1 FROM todo.list_members lm WHERE lm.list_id = l.id AND lm.user_id = $2))
RETURNING i.id;

-- MEMBERS

-- name: AddListMember :one
INSERT INTO todo.list_members (list_id, user_id, name, email)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: RemoveListMember :exec
DELETE FROM todo.list_members
WHERE list_id = $1 AND user_id = $2;

-- name: GetListMembers :many
SELECT * FROM todo.list_members
WHERE list_id = $1
ORDER BY created_at;

-- name: IsListMember :one
SELECT EXISTS(
    SELECT 1 FROM todo.list_members
    WHERE list_id = $1 AND user_id = $2
) AS is_member;

-- name: GetListOwner :one
SELECT user_id FROM todo.lists
WHERE id = $1;
