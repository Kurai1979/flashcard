-- +goose Up
-- +goose StatementBegin
CREATE TABLE deck_flashcards
(
    deck_id      UUID        NOT NULL REFERENCES decks (id) ON DELETE CASCADE,
    flashcard_id UUID        NOT NULL REFERENCES flashcards (id) ON DELETE CASCADE,
    position     INT         NULL,
    added_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (deck_id, flashcard_id)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX deck_flashcards_flashcard_id_idx ON deck_flashcards (flashcard_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS deck_flashcards;
-- +goose StatementEnd
