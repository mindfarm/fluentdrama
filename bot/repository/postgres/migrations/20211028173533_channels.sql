-- +goose Up
-- NOTE that a key value db is a better solution here, but the lack of data used
-- and RAM on my VPS mean it it more efficient to hold it in the same RDBMS as
-- the logs
CREATE TABLE [IF NOT EXISTS] channels (
    name TEXT NOT NULL,
    UNIQUE(name)
);


-- +goose Down
DROP TABLE [IF EXISTS] channels;
