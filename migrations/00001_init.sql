-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS keys (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    host_pattern TEXT NOT NULL,
    user_pattern TEXT NOT NULL,
    encrypted_pem BLOB NOT NULL,
    comment TEXT
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS keys;
-- +goose StatementEnd
