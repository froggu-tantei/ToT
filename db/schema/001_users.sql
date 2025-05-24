-- +goose Up
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  username TEXT NOT NULL UNIQUE,
  last_place_count INTEGER NOT NULL DEFAULT 0,
  profile_picture TEXT,
  bio TEXT
);

-- +goose Down
DROP TABLE users;