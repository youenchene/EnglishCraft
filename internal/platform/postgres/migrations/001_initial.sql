CREATE TABLE irregular_verbs (
    id TEXT PRIMARY KEY,
    base_form TEXT NOT NULL,
    past_simple TEXT NOT NULL,
    past_participle TEXT NOT NULL,
    translation_fr TEXT NOT NULL,
    skill_id TEXT NOT NULL CHECK (skill_id LIKE 'english.%')
);

-- statement
CREATE TABLE practice_attempts (
    id UUID PRIMARY KEY,
    student_id TEXT NOT NULL,
    verb_id TEXT NOT NULL REFERENCES irregular_verbs(id),
    skill_id TEXT NOT NULL,
    past_simple_answer TEXT NOT NULL,
    past_participle_answer TEXT NOT NULL,
    correct BOOLEAN NOT NULL,
    answered_at TIMESTAMPTZ NOT NULL
);

-- statement
CREATE INDEX practice_attempts_student_verb_idx
    ON practice_attempts (student_id, verb_id, answered_at DESC);

-- statement
CREATE TABLE skill_progress (
    student_id TEXT NOT NULL,
    skill_id TEXT NOT NULL,
    correct_answers INTEGER NOT NULL CHECK (correct_answers >= 0),
    total_answers INTEGER NOT NULL CHECK (total_answers >= correct_answers),
    validation_level TEXT NOT NULL CHECK (validation_level IN ('not_started', 'recognition', 'recall', 'context')),
    next_review_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (student_id, skill_id)
);

-- statement
INSERT INTO irregular_verbs (id, base_form, past_simple, past_participle, translation_fr, skill_id) VALUES
    ('verb-be', 'be', 'was/were', 'been', 'être', 'english.grammar.past_simple.irregular_verbs'),
    ('verb-build', 'build', 'built', 'built', 'construire', 'english.grammar.past_simple.irregular_verbs'),
    ('verb-come', 'come', 'came', 'come', 'venir', 'english.grammar.past_simple.irregular_verbs'),
    ('verb-do', 'do', 'did', 'done', 'faire', 'english.grammar.past_simple.irregular_verbs'),
    ('verb-find', 'find', 'found', 'found', 'trouver', 'english.grammar.past_simple.irregular_verbs'),
    ('verb-get', 'get', 'got', 'got', 'obtenir', 'english.grammar.past_simple.irregular_verbs'),
    ('verb-go', 'go', 'went', 'gone', 'aller', 'english.grammar.past_simple.irregular_verbs'),
    ('verb-have', 'have', 'had', 'had', 'avoir', 'english.grammar.past_simple.irregular_verbs'),
    ('verb-make', 'make', 'made', 'made', 'fabriquer', 'english.grammar.past_simple.irregular_verbs'),
    ('verb-win', 'win', 'won', 'won', 'gagner', 'english.grammar.past_simple.irregular_verbs');
