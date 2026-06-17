-- 0001_init: core meeting platform schema
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS meetings (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slug       text UNIQUE NOT NULL,
    title      text NOT NULL DEFAULT '',
    host_name  text NOT NULL DEFAULT '',
    status     text NOT NULL DEFAULT 'active',
    created_at timestamptz NOT NULL DEFAULT now(),
    ended_at   timestamptz
);
CREATE INDEX IF NOT EXISTS idx_meetings_status ON meetings (status);

CREATE TABLE IF NOT EXISTS participants (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    meeting_id   uuid NOT NULL REFERENCES meetings (id) ON DELETE CASCADE,
    display_name text NOT NULL,
    role         text NOT NULL DEFAULT 'guest',
    joined_at    timestamptz NOT NULL DEFAULT now(),
    left_at      timestamptz
);
CREATE INDEX IF NOT EXISTS idx_participants_meeting ON participants (meeting_id);

CREATE TABLE IF NOT EXISTS transcript_segments (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    meeting_id      uuid NOT NULL REFERENCES meetings (id) ON DELETE CASCADE,
    participant_id  uuid REFERENCES participants (id) ON DELETE SET NULL,
    source_lang     text,
    text_original   text NOT NULL,
    target_lang     text,
    text_translated text,
    is_final        boolean NOT NULL DEFAULT false,
    created_at      timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_segments_meeting_time ON transcript_segments (meeting_id, created_at);
