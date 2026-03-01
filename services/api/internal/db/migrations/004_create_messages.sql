CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sender_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    recipient_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_messages_sender_recipient_created_at
    ON messages (sender_user_id, recipient_user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_messages_recipient_sender_created_at
    ON messages (recipient_user_id, sender_user_id, created_at DESC);
