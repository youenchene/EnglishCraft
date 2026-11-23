CREATE TABLE family_accounts (
 id TEXT PRIMARY KEY,
 username TEXT NOT NULL UNIQUE CHECK (length(username) BETWEEN 3 AND 80),
 password_hash TEXT NOT NULL,
 role TEXT NOT NULL CHECK (role IN ('parent','child')),
 parent_id TEXT REFERENCES family_accounts(id),
 failed_logins INTEGER NOT NULL DEFAULT 0 CHECK (failed_logins >= 0),
 locked_until TIMESTAMPTZ,
 CHECK ((role='parent' AND parent_id IS NULL) OR (role='child' AND parent_id IS NOT NULL))
);
-- statement
CREATE TABLE auth_sessions (
 token_hash TEXT PRIMARY KEY,
 account_id TEXT NOT NULL REFERENCES family_accounts(id) ON DELETE CASCADE,
 expires_at TIMESTAMPTZ NOT NULL
);
-- statement
CREATE INDEX auth_sessions_expiry_idx ON auth_sessions(expires_at);
-- statement
CREATE TABLE student_states (
 student_id TEXT PRIMARY KEY REFERENCES family_accounts(id) ON DELETE CASCADE,
 state JSONB NOT NULL,
 updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 CHECK (jsonb_typeof(state)='object')
);
