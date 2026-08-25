-- name: CreateDeck :one
INSERT INTO decks(user_id, name, description)
VALUES ($1, $2, $3)
RETURNING id, user_id, name, description, created_at, updated_at;

-- name: UpdateDeck :one
UPDATE decks
SET name        = $3,
    description = $4,
    updated_at  = NOW()
WHERE id = $1
  AND user_id = $2
RETURNING id, user_id, name, description, created_at, updated_at;

-- name: DeleteDeck :execrows
DELETE
FROM decks
WHERE id = $1
  AND user_id = $2;

-- name: AddFlashcardToDeck :execrows
INSERT INTO deck_flashcards (deck_id, flashcard_id)
SELECT d.id, f.id
FROM decks d
         INNER JOIN flashcards f ON f.user_id = d.user_id
WHERE d.id = $1
  AND f.id = $2
  AND d.user_id = $3;

-- name: RemoveFlashcardFromDeck :execrows
DELETE
FROM deck_flashcards AS df
    USING decks d
WHERE df.deck_id = d.id
  AND df.deck_id = $1
  AND df.flashcard_id = $2
  AND d.user_id = $3;

-- name: GetDeck :one
SELECT d.id, d.user_id, d.name, d.description, d.created_at, d.updated_at
FROM decks d
WHERE d.id = $1
  AND d.user_id = $2;

-- name: ListDecks :many
SELECT d.id, d.user_id, d.name, d.description, d.created_at, d.updated_at
FROM decks d
WHERE d.user_id = $1
ORDER BY d.created_at DESC;
