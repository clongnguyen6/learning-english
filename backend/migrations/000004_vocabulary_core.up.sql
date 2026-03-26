CREATE TABLE vocab_books (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    title varchar NOT NULL,
    slug varchar NOT NULL,
    description text,
    cover_image_url text,
    level varchar,
    language_from varchar NOT NULL DEFAULT 'en',
    language_to varchar NOT NULL DEFAULT 'vi',
    status varchar NOT NULL DEFAULT 'draft',
    latest_import_id uuid REFERENCES imports (id) ON DELETE SET NULL,
    draft_revision_id uuid REFERENCES content_revisions (id) ON DELETE RESTRICT,
    active_publication_id uuid REFERENCES publications (id) ON DELETE RESTRICT,
    word_count integer NOT NULL DEFAULT 0,
    sort_order integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    version integer NOT NULL DEFAULT 1,
    CONSTRAINT vocab_books_title_not_blank CHECK (btrim(title) <> ''),
    CONSTRAINT vocab_books_slug_not_blank CHECK (btrim(slug) <> ''),
    CONSTRAINT vocab_books_description_not_blank CHECK (description IS NULL OR btrim(description) <> ''),
    CONSTRAINT vocab_books_cover_image_url_not_blank CHECK (
        cover_image_url IS NULL OR btrim(cover_image_url) <> ''
    ),
    CONSTRAINT vocab_books_level_not_blank CHECK (level IS NULL OR btrim(level) <> ''),
    CONSTRAINT vocab_books_language_from_not_blank CHECK (btrim(language_from) <> ''),
    CONSTRAINT vocab_books_language_to_not_blank CHECK (btrim(language_to) <> ''),
    CONSTRAINT vocab_books_status_valid CHECK (status IN ('draft', 'published', 'archived')),
    CONSTRAINT vocab_books_word_count_non_negative CHECK (word_count >= 0),
    CONSTRAINT vocab_books_sort_order_non_negative CHECK (sort_order >= 0),
    CONSTRAINT vocab_books_version_positive CHECK (version > 0),
    CONSTRAINT vocab_books_active_publication_invariant CHECK (
        (status = 'published' AND active_publication_id IS NOT NULL)
        OR (status <> 'published' AND active_publication_id IS NULL)
    )
);

CREATE UNIQUE INDEX vocab_books_slug_key ON vocab_books (slug);
CREATE UNIQUE INDEX idx_vocab_books_active_publication_id
    ON vocab_books (active_publication_id)
    WHERE active_publication_id IS NOT NULL;
CREATE INDEX idx_vocab_books_status_sort_order ON vocab_books (status, sort_order, title);
CREATE INDEX idx_vocab_books_latest_import_id ON vocab_books (latest_import_id) WHERE latest_import_id IS NOT NULL;

CREATE TABLE vocab_words (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    book_id uuid NOT NULL REFERENCES vocab_books (id) ON DELETE CASCADE,
    word varchar NOT NULL,
    normalized_word varchar NOT NULL,
    ipa varchar,
    part_of_speech varchar,
    meaning_vi text,
    meaning_en text,
    context text,
    tips text,
    image_url text,
    status_default varchar NOT NULL DEFAULT 'new',
    sort_order integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT vocab_words_word_not_blank CHECK (btrim(word) <> ''),
    CONSTRAINT vocab_words_normalized_word_not_blank CHECK (btrim(normalized_word) <> ''),
    CONSTRAINT vocab_words_ipa_not_blank CHECK (ipa IS NULL OR btrim(ipa) <> ''),
    CONSTRAINT vocab_words_part_of_speech_not_blank CHECK (part_of_speech IS NULL OR btrim(part_of_speech) <> ''),
    CONSTRAINT vocab_words_context_not_blank CHECK (context IS NULL OR btrim(context) <> ''),
    CONSTRAINT vocab_words_tips_not_blank CHECK (tips IS NULL OR btrim(tips) <> ''),
    CONSTRAINT vocab_words_image_url_not_blank CHECK (image_url IS NULL OR btrim(image_url) <> ''),
    CONSTRAINT vocab_words_status_default_valid CHECK (status_default IN ('new', 'learning', 'mastered')),
    CONSTRAINT vocab_words_sort_order_non_negative CHECK (sort_order >= 0),
    CONSTRAINT vocab_words_has_meaning CHECK (
        COALESCE(NULLIF(btrim(meaning_vi), ''), NULLIF(btrim(meaning_en), '')) IS NOT NULL
    )
);

CREATE UNIQUE INDEX idx_vocab_words_book_normalized_word ON vocab_words (book_id, normalized_word);
CREATE INDEX idx_vocab_words_book_sort_order ON vocab_words (book_id, sort_order, normalized_word);

CREATE TABLE vocab_word_examples (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    word_id uuid NOT NULL REFERENCES vocab_words (id) ON DELETE CASCADE,
    example_en text NOT NULL,
    example_vi text,
    sort_order integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT vocab_word_examples_example_en_not_blank CHECK (btrim(example_en) <> ''),
    CONSTRAINT vocab_word_examples_example_vi_not_blank CHECK (example_vi IS NULL OR btrim(example_vi) <> ''),
    CONSTRAINT vocab_word_examples_sort_order_non_negative CHECK (sort_order >= 0)
);

CREATE INDEX idx_vocab_word_examples_word_sort_order ON vocab_word_examples (word_id, sort_order, created_at);

CREATE TABLE vocab_word_media (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    word_id uuid NOT NULL REFERENCES vocab_words (id) ON DELETE CASCADE,
    media_type varchar NOT NULL,
    media_url text NOT NULL,
    sort_order integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT vocab_word_media_type_valid CHECK (media_type IN ('image', 'audio')),
    CONSTRAINT vocab_word_media_url_not_blank CHECK (btrim(media_url) <> ''),
    CONSTRAINT vocab_word_media_sort_order_non_negative CHECK (sort_order >= 0)
);

CREATE INDEX idx_vocab_word_media_word_sort_order ON vocab_word_media (word_id, sort_order, created_at);

CREATE TABLE user_vocab_progress (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    word_id uuid NOT NULL REFERENCES vocab_words (id) ON DELETE CASCADE,
    learning_status varchar NOT NULL DEFAULT 'new',
    last_studied_at timestamptz,
    correct_count integer NOT NULL DEFAULT 0,
    wrong_count integer NOT NULL DEFAULT 0,
    streak_count integer NOT NULL DEFAULT 0,
    next_review_at timestamptz,
    is_favorite boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    version integer NOT NULL DEFAULT 1,
    CONSTRAINT user_vocab_progress_learning_status_valid CHECK (
        learning_status IN ('new', 'learning', 'mastered')
    ),
    CONSTRAINT user_vocab_progress_correct_count_non_negative CHECK (correct_count >= 0),
    CONSTRAINT user_vocab_progress_wrong_count_non_negative CHECK (wrong_count >= 0),
    CONSTRAINT user_vocab_progress_streak_count_non_negative CHECK (streak_count >= 0),
    CONSTRAINT user_vocab_progress_version_positive CHECK (version > 0),
    CONSTRAINT user_vocab_progress_user_word_unique UNIQUE (user_id, word_id)
);

CREATE INDEX idx_user_vocab_progress_user_status_review
    ON user_vocab_progress (user_id, learning_status, next_review_at);
CREATE INDEX idx_user_vocab_progress_word_id ON user_vocab_progress (word_id);

CREATE TABLE vocab_study_sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    book_id uuid NOT NULL REFERENCES vocab_books (id) ON DELETE CASCADE,
    mode varchar NOT NULL,
    study_mode varchar NOT NULL,
    direction varchar NOT NULL,
    ordering varchar NOT NULL,
    show_ipa boolean NOT NULL DEFAULT true,
    status_filter varchar,
    auto_next boolean NOT NULL DEFAULT true,
    current_question_seq integer NOT NULL DEFAULT 0,
    started_at timestamptz NOT NULL DEFAULT now(),
    ended_at timestamptz,
    total_items integer NOT NULL DEFAULT 0,
    completed_items integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    version integer NOT NULL DEFAULT 1,
    CONSTRAINT vocab_study_sessions_mode_valid CHECK (mode IN ('guess_meaning', 'learn_word')),
    CONSTRAINT vocab_study_sessions_study_mode_valid CHECK (study_mode IN ('flip', 'write')),
    CONSTRAINT vocab_study_sessions_direction_valid CHECK (direction IN ('en_vi', 'vi_en')),
    CONSTRAINT vocab_study_sessions_ordering_valid CHECK (ordering IN ('random', 'sequential', 'repeat')),
    CONSTRAINT vocab_study_sessions_status_filter_valid CHECK (
        status_filter IS NULL OR status_filter IN ('all', 'new', 'learning', 'mastered')
    ),
    CONSTRAINT vocab_study_sessions_current_question_seq_non_negative CHECK (current_question_seq >= 0),
    CONSTRAINT vocab_study_sessions_current_question_seq_within_total CHECK (current_question_seq <= total_items),
    CONSTRAINT vocab_study_sessions_total_items_non_negative CHECK (total_items >= 0),
    CONSTRAINT vocab_study_sessions_completed_items_non_negative CHECK (completed_items >= 0),
    CONSTRAINT vocab_study_sessions_completed_items_within_total CHECK (completed_items <= total_items),
    CONSTRAINT vocab_study_sessions_ended_after_start CHECK (ended_at IS NULL OR ended_at >= started_at),
    CONSTRAINT vocab_study_sessions_version_positive CHECK (version > 0)
);

CREATE INDEX idx_vocab_study_sessions_user_book_created_at
    ON vocab_study_sessions (user_id, book_id, created_at DESC);
CREATE INDEX idx_vocab_study_sessions_user_open
    ON vocab_study_sessions (user_id, book_id, updated_at DESC)
    WHERE ended_at IS NULL;

CREATE TABLE vocab_session_items (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id uuid NOT NULL REFERENCES vocab_study_sessions (id) ON DELETE CASCADE,
    word_id uuid NOT NULL REFERENCES vocab_words (id) ON DELETE CASCADE,
    question_seq integer NOT NULL,
    question_type varchar NOT NULL,
    prompt_language varchar NOT NULL,
    prompt_text text NOT NULL,
    options_json jsonb,
    correct_answer text NOT NULL,
    status varchar NOT NULL DEFAULT 'pending',
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT vocab_session_items_question_seq_positive CHECK (question_seq > 0),
    CONSTRAINT vocab_session_items_question_type_valid CHECK (question_type IN ('mcq', 'write', 'flip')),
    CONSTRAINT vocab_session_items_prompt_language_valid CHECK (prompt_language IN ('en', 'vi')),
    CONSTRAINT vocab_session_items_prompt_text_not_blank CHECK (btrim(prompt_text) <> ''),
    CONSTRAINT vocab_session_items_correct_answer_not_blank CHECK (btrim(correct_answer) <> ''),
    CONSTRAINT vocab_session_items_status_valid CHECK (status IN ('pending', 'answered', 'skipped')),
    CONSTRAINT vocab_session_items_options_json_array CHECK (
        options_json IS NULL OR jsonb_typeof(options_json) = 'array'
    ),
    CONSTRAINT vocab_session_items_session_question_unique UNIQUE (session_id, question_seq),
    CONSTRAINT vocab_session_items_id_session_unique UNIQUE (id, session_id)
);

CREATE INDEX idx_vocab_session_items_session_status_seq
    ON vocab_session_items (session_id, status, question_seq);
CREATE INDEX idx_vocab_session_items_word_id ON vocab_session_items (word_id);

CREATE TABLE vocab_quiz_attempts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id uuid NOT NULL REFERENCES vocab_study_sessions (id) ON DELETE CASCADE,
    session_item_id uuid NOT NULL REFERENCES vocab_session_items (id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    word_id uuid NOT NULL REFERENCES vocab_words (id) ON DELETE CASCADE,
    question_type varchar NOT NULL,
    prompt_language varchar NOT NULL,
    answer_text text,
    selected_option_index integer,
    is_correct boolean,
    answered_at timestamptz NOT NULL DEFAULT now(),
    request_idempotency_key varchar,
    CONSTRAINT vocab_quiz_attempts_question_type_valid CHECK (question_type IN ('mcq', 'write', 'flip')),
    CONSTRAINT vocab_quiz_attempts_prompt_language_valid CHECK (prompt_language IN ('en', 'vi')),
    CONSTRAINT vocab_quiz_attempts_answer_text_not_blank CHECK (answer_text IS NULL OR btrim(answer_text) <> ''),
    CONSTRAINT vocab_quiz_attempts_selected_option_index_non_negative CHECK (
        selected_option_index IS NULL OR selected_option_index >= 0
    ),
    CONSTRAINT vocab_quiz_attempts_request_idempotency_key_not_blank CHECK (
        request_idempotency_key IS NULL OR btrim(request_idempotency_key) <> ''
    ),
    CONSTRAINT vocab_quiz_attempts_session_item_unique UNIQUE (session_id, session_item_id),
    CONSTRAINT vocab_quiz_attempts_session_item_fk FOREIGN KEY (session_item_id, session_id)
        REFERENCES vocab_session_items (id, session_id) ON DELETE CASCADE
);

CREATE INDEX idx_vocab_quiz_attempts_session_answered_at
    ON vocab_quiz_attempts (session_id, answered_at DESC);
CREATE INDEX idx_vocab_quiz_attempts_user_answered_at
    ON vocab_quiz_attempts (user_id, answered_at DESC);
