-- name: GetDelivery :one
SELECT * FROM deliveries
WHERE id = ? LIMIT 1;

-- name: ListDeliveries :many
SELECT * FROM deliveries
WHERE "date" > ?
ORDER BY "date" DESC;

-- name: SearchDeliveriesByName :many
SELECT d.*
FROM deliveries d
JOIN deliveries_fts fts ON d.id = fts.id
WHERE d.date > ?
  AND fts.location_name MATCH ?
GROUP BY d.id
ORDER BY d.date DESC;

-- name: CreateDelivery :one
INSERT INTO deliveries (
  date, schedule, location_type, location_name, created_at
) VALUES (
  ?, ?, ?, ?, unixepoch()
)
RETURNING *;

-- name: DeleteDelivery :exec
DELETE FROM deliveries
WHERE id = ?;

-- name: GetPendingImports :many
SELECT * FROM imports
WHERE completed_at IS NULL
AND runs < ?
ORDER BY created_at DESC;

-- name: CountImportsByHash :one
SELECT COUNT(*) FROM imports
WHERE file_hash = ?;

-- name: CreateImport :one
INSERT INTO imports (
  file_path, file_hash, completed_at, runs, created_at
) VALUES (
  ?, ?, NULL, 0, unixepoch()
)
RETURNING *;

-- name: CompleteImport :exec
UPDATE imports
SET completed_at = unixepoch(),
    runs = runs + 1
WHERE id = ?;

-- name: FailImport :exec
UPDATE imports
SET failed_at = unixepoch(),
    runs = runs + 1
WHERE id = ?;

-- name: GetLatestImport :one
SELECT * FROM imports
WHERE completed_at IS NOT NULL
ORDER BY created_at DESC
LIMIT 1;
