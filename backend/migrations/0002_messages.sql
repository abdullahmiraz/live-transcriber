-- 0002_messages: in-meeting realtime text chat (PostgreSQL is the source of truth).
CREATE TABLE IF NOT EXISTS messages (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    meeting_id  uuid NOT NULL REFERENCES meetings (id) ON DELETE CASCADE,
    sender_id   text NOT NULL,
    sender_name text NOT NULL DEFAULT '',
    content     text NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now()
);

-- Optimized for chronological reads + keyset pagination within a meeting.
CREATE INDEX IF NOT EXISTS idx_messages_meeting_created
    ON messages (meeting_id, created_at DESC, id DESC);
