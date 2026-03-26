CREATE TABLE reading_documents (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    title varchar NOT NULL,
    slug varchar NOT NULL,
    description text,
    is_bilingual boolean NOT NULL DEFAULT true,
    default_display_mode varchar NOT NULL DEFAULT 'docs',
    level varchar,
    status varchar NOT NULL DEFAULT 'draft',
    latest_import_id uuid REFERENCES imports (id) ON DELETE SET NULL,
    draft_revision_id uuid REFERENCES content_revisions (id) ON DELETE RESTRICT,
    active_publication_id uuid REFERENCES publications (id) ON DELETE RESTRICT,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    version integer NOT NULL DEFAULT 1,
    CONSTRAINT reading_documents_title_not_blank CHECK (btrim(title) <> ''),
    CONSTRAINT reading_documents_slug_not_blank CHECK (btrim(slug) <> ''),
    CONSTRAINT reading_documents_description_not_blank CHECK (description IS NULL OR btrim(description) <> ''),
    CONSTRAINT reading_documents_default_display_mode_valid CHECK (default_display_mode IN ('docs')),
    CONSTRAINT reading_documents_level_not_blank CHECK (level IS NULL OR btrim(level) <> ''),
    CONSTRAINT reading_documents_status_valid CHECK (status IN ('draft', 'published', 'archived')),
    CONSTRAINT reading_documents_version_positive CHECK (version > 0),
    CONSTRAINT reading_documents_active_publication_invariant CHECK (
        (status = 'published' AND active_publication_id IS NOT NULL)
        OR (status <> 'published' AND active_publication_id IS NULL)
    )
);

CREATE UNIQUE INDEX reading_documents_slug_key ON reading_documents (slug);
CREATE UNIQUE INDEX idx_reading_documents_active_publication_id
    ON reading_documents (active_publication_id)
    WHERE active_publication_id IS NOT NULL;
CREATE INDEX idx_reading_documents_status_title ON reading_documents (status, title);
CREATE INDEX idx_reading_documents_latest_import_id
    ON reading_documents (latest_import_id)
    WHERE latest_import_id IS NOT NULL;

CREATE TABLE reading_sections (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    publication_id uuid NOT NULL REFERENCES publications (id) ON DELETE CASCADE,
    section_order integer NOT NULL DEFAULT 0,
    heading varchar,
    content_en text,
    content_vi text,
    keywords_json jsonb,
    blocks_json jsonb NOT NULL,
    content_hash varchar,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT reading_sections_section_order_non_negative CHECK (section_order >= 0),
    CONSTRAINT reading_sections_heading_not_blank CHECK (heading IS NULL OR btrim(heading) <> ''),
    CONSTRAINT reading_sections_content_en_not_blank CHECK (content_en IS NULL OR btrim(content_en) <> ''),
    CONSTRAINT reading_sections_content_vi_not_blank CHECK (content_vi IS NULL OR btrim(content_vi) <> ''),
    CONSTRAINT reading_sections_has_content CHECK (
        COALESCE(NULLIF(btrim(content_en), ''), NULLIF(btrim(content_vi), '')) IS NOT NULL
    ),
    CONSTRAINT reading_sections_keywords_json_array CHECK (
        keywords_json IS NULL OR jsonb_typeof(keywords_json) = 'array'
    ),
    CONSTRAINT reading_sections_blocks_json_array CHECK (jsonb_typeof(blocks_json) = 'array'),
    CONSTRAINT reading_sections_content_hash_not_blank CHECK (content_hash IS NULL OR btrim(content_hash) <> ''),
    CONSTRAINT reading_sections_publication_order_unique UNIQUE (publication_id, section_order),
    CONSTRAINT reading_sections_id_publication_unique UNIQUE (id, publication_id)
);

CREATE INDEX idx_reading_sections_publication_order
    ON reading_sections (publication_id, section_order);
CREATE INDEX idx_reading_sections_publication_hash
    ON reading_sections (publication_id, content_hash)
    WHERE content_hash IS NOT NULL;

CREATE TABLE reading_highlights (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    publication_id uuid NOT NULL,
    section_order integer NOT NULL,
    keyword varchar NOT NULL,
    color varchar NOT NULL DEFAULT 'yellow',
    start_index integer,
    end_index integer,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT reading_highlights_keyword_not_blank CHECK (btrim(keyword) <> ''),
    CONSTRAINT reading_highlights_color_not_blank CHECK (btrim(color) <> ''),
    CONSTRAINT reading_highlights_section_order_non_negative CHECK (section_order >= 0),
    CONSTRAINT reading_highlights_start_index_non_negative CHECK (
        start_index IS NULL OR start_index >= 0
    ),
    CONSTRAINT reading_highlights_end_index_non_negative CHECK (
        end_index IS NULL OR end_index >= 0
    ),
    CONSTRAINT reading_highlights_range_consistent CHECK (
        (start_index IS NULL AND end_index IS NULL)
        OR (
            start_index IS NOT NULL
            AND end_index IS NOT NULL
            AND end_index >= start_index
        )
    ),
    CONSTRAINT reading_highlights_section_fk FOREIGN KEY (publication_id, section_order)
        REFERENCES reading_sections (publication_id, section_order) ON DELETE CASCADE
);

CREATE INDEX idx_reading_highlights_publication_section
    ON reading_highlights (publication_id, section_order, created_at);
CREATE INDEX idx_reading_highlights_publication_keyword
    ON reading_highlights (publication_id, keyword);

CREATE TABLE reading_progress (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    document_id uuid NOT NULL REFERENCES reading_documents (id) ON DELETE CASCADE,
    last_section_id uuid,
    last_section_order integer NOT NULL DEFAULT 0,
    publication_id uuid REFERENCES publications (id) ON DELETE RESTRICT,
    completion_percent numeric(5,2) NOT NULL DEFAULT 0,
    last_read_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    version integer NOT NULL DEFAULT 1,
    CONSTRAINT reading_progress_last_section_order_non_negative CHECK (last_section_order >= 0),
    CONSTRAINT reading_progress_completion_percent_range CHECK (
        completion_percent >= 0 AND completion_percent <= 100
    ),
    CONSTRAINT reading_progress_version_positive CHECK (version > 0),
    CONSTRAINT reading_progress_last_read_after_create CHECK (
        last_read_at IS NULL OR last_read_at >= created_at
    ),
    CONSTRAINT reading_progress_last_section_requires_publication CHECK (
        last_section_id IS NULL OR publication_id IS NOT NULL
    ),
    CONSTRAINT reading_progress_last_section_order_requires_section CHECK (
        last_section_id IS NOT NULL OR last_section_order = 0
    ),
    CONSTRAINT reading_progress_user_document_unique UNIQUE (user_id, document_id),
    CONSTRAINT reading_progress_section_publication_fk FOREIGN KEY (last_section_id, publication_id)
        REFERENCES reading_sections (id, publication_id)
);

CREATE INDEX idx_reading_progress_user_last_read
    ON reading_progress (user_id, last_read_at DESC);
CREATE INDEX idx_reading_progress_document_publication
    ON reading_progress (document_id, publication_id);
