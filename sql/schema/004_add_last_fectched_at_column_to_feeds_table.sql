-- +goose up
ALTER TABLE
  feeds
ADD
  last_fetched_at TIMESTAMPTZ;

-- +goose down
ALTER TABLE
  feeds
DROP COLUMN
  last_fetched_at;
