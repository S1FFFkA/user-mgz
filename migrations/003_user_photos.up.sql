CREATE TABLE user_photos (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    object_key TEXT NOT NULL,
    url TEXT NOT NULL,
    position SMALLINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
