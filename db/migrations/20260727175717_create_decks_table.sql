-- +goose Up
-- +goose StatementBegin
CREATE TABLE decks
(
    id          UUID PRIMARY KEY     DEFAULT uuidv7(),
    user_id     UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name        TEXT        NOT NULL,
    description TEXT        NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NULL
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX decks_user_id_idx ON decks (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS decks;
-- +goose StatementEnd
