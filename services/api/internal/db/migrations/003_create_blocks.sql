CREATE TABLE IF NOT EXISTS blocks (
    blocker_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    blocked_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (blocker_user_id, blocked_user_id)
);

CREATE INDEX IF NOT EXISTS idx_blocks_blocker_blocked ON blocks (blocker_user_id, blocked_user_id);
CREATE INDEX IF NOT EXISTS idx_blocks_blocked_blocker ON blocks (blocked_user_id, blocker_user_id);
