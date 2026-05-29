-- +goose up
CREATE TABLE feeds (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    name VARCHAR(255),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    url VARCHAR(255) UNIQUE NOT NULL
);

-- +goose down
DROP TABLE feeds;
