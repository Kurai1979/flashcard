-- +goose Up
-- +goose StatementBegin
-- ListFlashcards pages with ORDER BY id DESC scoped to one user. The existing
-- flashcards_user_id_idx can serve the filter but leaves the ordering to a sort;
-- carrying id in the index lets the keyset scan walk straight down it.
CREATE INDEX flashcards_user_id_id_idx ON flashcards (user_id, id DESC);
-- +goose StatementEnd

-- +goose StatementBegin
-- Redundant now: (user_id, id DESC) covers every lookup the single-column index did.
DROP INDEX IF EXISTS flashcards_user_id_idx;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE INDEX flashcards_user_id_idx ON flashcards (user_id);
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX IF EXISTS flashcards_user_id_id_idx;
-- +goose StatementEnd
