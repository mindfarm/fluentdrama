-- +goose Up
-- SQL in this section is executed when the migration is applied.
CREATE TABLE IF NOT EXISTS logs (
    channel TEXT NOT NULL,
    nick TEXT NOT NULL,
    stamp TIMESTAMP DEFAULT NOW(),
    said TEXT
);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP TABLE IF EXISTS logs;
