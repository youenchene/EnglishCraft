CREATE TABLE ai_usage (
 id TEXT PRIMARY KEY,
 student_id TEXT NOT NULL REFERENCES family_accounts(id) ON DELETE CASCADE,
 provider TEXT NOT NULL,
 day DATE NOT NULL,
 reserved_micros BIGINT NOT NULL CHECK(reserved_micros > 0),
 voice_seconds INTEGER NOT NULL DEFAULT 0 CHECK(voice_seconds >= 0),
 input_tokens INTEGER NOT NULL DEFAULT 0,
 output_tokens INTEGER NOT NULL DEFAULT 0,
 status TEXT NOT NULL DEFAULT 'reserved'
);
-- statement
CREATE INDEX ai_usage_day_idx ON ai_usage(day,provider,student_id);
-- statement
CREATE TABLE voice_sessions (
 id TEXT PRIMARY KEY,
 student_id TEXT NOT NULL REFERENCES family_accounts(id) ON DELETE CASCADE,
 mission_id TEXT NOT NULL,
 expires_at TIMESTAMPTZ NOT NULL
);
-- statement
CREATE TABLE school_imports (
 id TEXT PRIMARY KEY,
 student_id TEXT NOT NULL REFERENCES family_accounts(id) ON DELETE CASCADE,
 document JSONB NOT NULL CHECK(jsonb_typeof(document)='object'),
 approved BOOLEAN NOT NULL DEFAULT false
);
-- statement
CREATE TABLE ai_preferences (
 student_id TEXT PRIMARY KEY REFERENCES family_accounts(id) ON DELETE CASCADE,
 enabled BOOLEAN NOT NULL DEFAULT false
);
