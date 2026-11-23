ALTER TABLE skill_progress
    ADD COLUMN version BIGINT NOT NULL DEFAULT 1 CHECK (version > 0);
