-- name: CreateFlashcard :one
INSERT INTO flashcards(user_id, front, back, example)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, front, back, example, created_at, updated_at;

-- name: UpdateFlashcard :one
UPDATE flashcards
SET front      = $1,
    back       = $2,
    example    = $3,
    updated_at = NOW()
WHERE id = $4
  AND user_id = $5
RETURNING id, user_id, front, back, example, created_at, updated_at;

-- name: DeleteFlashcard :execrows
DELETE
FROM flashcards
WHERE id = $1
  AND user_id = $2;

-- name: ListFlashcards :many
-- Keyset pagination, newest first. Flashcard ids are uuidv7, whose leading bits
-- are a timestamp, so ordering by id is chronological and the last id on a page
-- is a stable cursor: unlike OFFSET it can't skip or repeat rows when cards are
-- inserted mid-listing, and it never scans the rows it skips. Pass a NULL cursor
-- for the first page.
SELECT f.id, f.front, f.back, f.example, f.created_at, f.updated_at
FROM flashcards AS f
WHERE f.user_id = @user_id
  AND (sqlc.narg(cursor)::uuid IS NULL OR f.id < sqlc.narg(cursor))
ORDER BY f.id DESC
LIMIT @page_size;