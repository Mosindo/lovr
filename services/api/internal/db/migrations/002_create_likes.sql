CREATE TABLE IF NOT EXISTS likes (
    from_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    to_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (from_user_id, to_user_id)
);

CREATE INDEX IF NOT EXISTS idx_likes_from_to ON likes (from_user_id, to_user_id);
CREATE INDEX IF NOT EXISTS idx_likes_to_from ON likes (to_user_id, from_user_id);
