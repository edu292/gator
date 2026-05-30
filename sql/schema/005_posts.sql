-- +goose up
CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    title VARCHAR(100) NOT NULL,
    url VARCHAR(255) UNIQUE NOT NULL,
    description TEXT DEFAULT '' NOT NULL,
    published_at TIMESTAMPTZ,
    feed_id INT REFERENCES feeds (id) ON DELETE CASCADE NOT NULL
);

-- +goose down
DROP TABLE posts;
