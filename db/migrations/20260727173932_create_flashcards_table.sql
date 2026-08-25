-- +goose Up
-- +goose StatementBegin
CREATE TABLE flashcards
(
    id         UUID PRIMARY KEY     DEFAULT uuidv7(),
    user_id    UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    front      TEXT        NOT NULL,
    back       TEXT        NOT NULL,
    example    TEXT        NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NULL
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX flashcards_user_id_idx ON flashcards (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS flashcards;
-- +goose StatementEnd
