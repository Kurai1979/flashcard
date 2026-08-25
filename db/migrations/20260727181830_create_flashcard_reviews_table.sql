-- +goose Up
-- +goose StatementBegin
CREATE TABLE flashcard_reviews
(
    id           UUID PRIMARY KEY     DEFAULT uuidv7(),
    flashcard_id UUID        NOT NULL REFERENCES flashcards (id) ON DELETE CASCADE,
    user_id      UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    rating       SMALLINT    NOT NULL CHECK (rating BETWEEN 1 AND 4),
    duration_ms  INT         NULL CHECK (duration_ms IS NULL OR duration_ms >= 0),
    reviewed_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose StatementBegin
COMMENT ON COLUMN flashcard_reviews.rating IS '1=again, 2=hard, 3=good, 4=easy';
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX flashcard_reviews_card_idx ON flashcard_reviews (flashcard_id, reviewed_at DESC);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX flashcard_reviews_user_idx ON flashcard_reviews (user_id, reviewed_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS flashcard_reviews;
-- +goose StatementEnd
