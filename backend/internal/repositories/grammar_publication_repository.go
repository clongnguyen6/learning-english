package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"learning-english/backend/internal/database"
	"learning-english/backend/internal/models"
)

type GrammarImportStateRepository interface {
	MarkCommitted(ctx context.Context, importID string) (models.Import, error)
}

type GrammarContentRevisionRepository interface {
	FindByID(ctx context.Context, revisionID string) (models.ContentRevision, error)
	NextVersionNo(ctx context.Context, entityType, entityID string) (int, error)
	Create(ctx context.Context, params CreateGrammarContentRevisionParams) (models.ContentRevision, error)
}

type GrammarTopicRepository interface {
	FindBySlug(ctx context.Context, slug string) (models.GrammarTopic, error)
	Create(ctx context.Context, params CreateGrammarTopicParams) (models.GrammarTopic, error)
}

type GrammarLessonRepository interface {
	FindBySlug(ctx context.Context, slug string) (models.GrammarLesson, error)
	Create(ctx context.Context, params CreateGrammarLessonParams) (models.GrammarLesson, error)
	SaveDraft(ctx context.Context, params SaveGrammarLessonDraftParams) (models.GrammarLesson, error)
	ApplyPublication(ctx context.Context, params ApplyGrammarLessonPublicationParams) (models.GrammarLesson, error)
}

type CreateGrammarTopicParams struct {
	Title       string
	Slug        string
	Description string
	Level       string
	SortOrder   int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CreateGrammarLessonParams struct {
	TopicID        string
	Title          string
	Slug           string
	Summary        string
	Status         string
	LatestImportID *string
	SortOrder      int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type SaveGrammarLessonDraftParams struct {
	LessonID            string
	TopicID             string
	Title               string
	Slug                string
	Summary             string
	Status              string
	LatestImportID      *string
	DraftRevisionID     *string
	ActivePublicationID *string
	UpdatedAt           time.Time
}

type ApplyGrammarLessonPublicationParams struct {
	LessonID            string
	TopicID             string
	Title               string
	Slug                string
	Summary             string
	Status              string
	ActivePublicationID *string
	UpdatedAt           time.Time
}

type CreateGrammarContentRevisionParams struct {
	EntityType        string
	EntityID          string
	VersionNo         int
	ImportID          *string
	SourceType        string
	SourcePayload     string
	ParsedPayloadJSON map[string]any
	KeywordsJSON      []string
	RenderArtifactURL string
	ChangeSummary     string
	CreatedBy         string
	CreatedAt         time.Time
}

type sqlGrammarImportStateRepository struct {
	exec database.DBTX
}

type sqlGrammarContentRevisionRepository struct {
	exec database.DBTX
}

type sqlGrammarTopicRepository struct {
	exec database.DBTX
}

type sqlGrammarLessonRepository struct {
	exec database.DBTX
}

func NewGrammarImportStateRepository(exec database.DBTX) GrammarImportStateRepository {
	return &sqlGrammarImportStateRepository{exec: exec}
}

func NewGrammarContentRevisionRepository(exec database.DBTX) GrammarContentRevisionRepository {
	return &sqlGrammarContentRevisionRepository{exec: exec}
}

func NewGrammarTopicRepository(exec database.DBTX) GrammarTopicRepository {
	return &sqlGrammarTopicRepository{exec: exec}
}

func NewGrammarLessonRepository(exec database.DBTX) GrammarLessonRepository {
	return &sqlGrammarLessonRepository{exec: exec}
}

func (r *sqlGrammarImportStateRepository) MarkCommitted(
	ctx context.Context,
	importID string,
) (models.Import, error) {
	if r == nil || r.exec == nil {
		return models.Import{}, errors.New("grammar import state repository is not configured")
	}

	row := r.exec.QueryRowContext(ctx, `
		UPDATE imports
		SET status = $2
		WHERE id = $1
		  AND status IN ($3, $4, $5)
		RETURNING `+importColumnsSQL,
		strings.TrimSpace(importID),
		models.ImportStatusCommitted,
		models.ImportStatusValidated,
		models.ImportStatusCommitQueued,
		models.ImportStatusCommitted,
	)

	record, err := scanImport(row)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return models.Import{}, ErrConflict
		}
		return models.Import{}, err
	}

	return record, nil
}

func (r *sqlGrammarContentRevisionRepository) FindByID(
	ctx context.Context,
	revisionID string,
) (models.ContentRevision, error) {
	if r == nil || r.exec == nil {
		return models.ContentRevision{}, errors.New("grammar content revision repository is not configured")
	}

	row := r.exec.QueryRowContext(
		ctx,
		selectGrammarContentRevisionsSQL+` WHERE id = $1 LIMIT 1`,
		strings.TrimSpace(revisionID),
	)
	return scanGrammarContentRevision(row)
}

func (r *sqlGrammarContentRevisionRepository) NextVersionNo(
	ctx context.Context,
	entityType, entityID string,
) (int, error) {
	if r == nil || r.exec == nil {
		return 0, errors.New("grammar content revision repository is not configured")
	}

	var version sql.NullInt64
	err := r.exec.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(version_no), 0) + 1
		FROM content_revisions
		WHERE entity_type = $1
		  AND entity_id = $2
	`,
		strings.TrimSpace(entityType),
		strings.TrimSpace(entityID),
	).Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("query next grammar revision version: %w", err)
	}
	if !version.Valid || version.Int64 < 1 {
		return 1, nil
	}

	return int(version.Int64), nil
}

func (r *sqlGrammarContentRevisionRepository) Create(
	ctx context.Context,
	params CreateGrammarContentRevisionParams,
) (models.ContentRevision, error) {
	if r == nil || r.exec == nil {
		return models.ContentRevision{}, errors.New("grammar content revision repository is not configured")
	}

	parsedPayloadJSON, err := marshalOptionalJSONObject(params.ParsedPayloadJSON)
	if err != nil {
		return models.ContentRevision{}, fmt.Errorf("marshal grammar parsed payload json: %w", err)
	}

	var keywordsJSON any
	if len(params.KeywordsJSON) > 0 {
		raw, err := json.Marshal(params.KeywordsJSON)
		if err != nil {
			return models.ContentRevision{}, fmt.Errorf("marshal grammar keywords json: %w", err)
		}
		keywordsJSON = string(raw)
	}

	row := r.exec.QueryRowContext(ctx, `
		INSERT INTO content_revisions (
			entity_type,
			entity_id,
			version_no,
			import_id,
			source_type,
			source_payload,
			parsed_payload_json,
			keywords_json,
			render_artifact_url,
			change_summary,
			created_by,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8::jsonb, $9, $10, $11, $12)
		RETURNING `+grammarContentRevisionColumnsSQL,
		strings.TrimSpace(params.EntityType),
		strings.TrimSpace(params.EntityID),
		params.VersionNo,
		nullableStringPointer(params.ImportID),
		strings.TrimSpace(params.SourceType),
		strings.TrimSpace(params.SourcePayload),
		parsedPayloadJSON,
		keywordsJSON,
		nullableTrimmedString(params.RenderArtifactURL),
		nullableTrimmedString(params.ChangeSummary),
		strings.TrimSpace(params.CreatedBy),
		params.CreatedAt.UTC(),
	)

	record, err := scanGrammarContentRevision(row)
	if err != nil {
		if isUniqueViolation(err) {
			return models.ContentRevision{}, ErrConflict
		}
		return models.ContentRevision{}, err
	}

	return record, nil
}

func (r *sqlGrammarTopicRepository) FindBySlug(
	ctx context.Context,
	slug string,
) (models.GrammarTopic, error) {
	if r == nil || r.exec == nil {
		return models.GrammarTopic{}, errors.New("grammar topic repository is not configured")
	}

	row := r.exec.QueryRowContext(
		ctx,
		selectGrammarTopicsSQL+` WHERE slug = $1 LIMIT 1`,
		strings.TrimSpace(slug),
	)
	return scanGrammarTopic(row)
}

func (r *sqlGrammarTopicRepository) Create(
	ctx context.Context,
	params CreateGrammarTopicParams,
) (models.GrammarTopic, error) {
	if r == nil || r.exec == nil {
		return models.GrammarTopic{}, errors.New("grammar topic repository is not configured")
	}

	row := r.exec.QueryRowContext(ctx, `
		INSERT INTO grammar_topics (
			title,
			slug,
			description,
			level,
			sort_order,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING `+grammarTopicColumnsSQL,
		strings.TrimSpace(params.Title),
		strings.TrimSpace(params.Slug),
		nullableTrimmedString(params.Description),
		nullableTrimmedString(params.Level),
		params.SortOrder,
		params.CreatedAt.UTC(),
		params.UpdatedAt.UTC(),
	)

	record, err := scanGrammarTopic(row)
	if err != nil {
		if isUniqueViolation(err) {
			return models.GrammarTopic{}, ErrConflict
		}
		return models.GrammarTopic{}, err
	}

	return record, nil
}

func (r *sqlGrammarLessonRepository) FindBySlug(
	ctx context.Context,
	slug string,
) (models.GrammarLesson, error) {
	if r == nil || r.exec == nil {
		return models.GrammarLesson{}, errors.New("grammar lesson repository is not configured")
	}

	row := r.exec.QueryRowContext(
		ctx,
		selectGrammarLessonsSQL+` WHERE slug = $1 LIMIT 1`,
		strings.TrimSpace(slug),
	)
	return scanGrammarLesson(row)
}

func (r *sqlGrammarLessonRepository) Create(
	ctx context.Context,
	params CreateGrammarLessonParams,
) (models.GrammarLesson, error) {
	if r == nil || r.exec == nil {
		return models.GrammarLesson{}, errors.New("grammar lesson repository is not configured")
	}

	row := r.exec.QueryRowContext(ctx, `
		INSERT INTO grammar_lessons (
			topic_id,
			title,
			slug,
			summary,
			status,
			latest_import_id,
			sort_order,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING `+grammarLessonColumnsSQL,
		strings.TrimSpace(params.TopicID),
		strings.TrimSpace(params.Title),
		strings.TrimSpace(params.Slug),
		nullableTrimmedString(params.Summary),
		strings.TrimSpace(params.Status),
		nullableStringPointer(params.LatestImportID),
		params.SortOrder,
		params.CreatedAt.UTC(),
		params.UpdatedAt.UTC(),
	)

	record, err := scanGrammarLesson(row)
	if err != nil {
		if isUniqueViolation(err) {
			return models.GrammarLesson{}, ErrConflict
		}
		return models.GrammarLesson{}, err
	}

	return record, nil
}

func (r *sqlGrammarLessonRepository) SaveDraft(
	ctx context.Context,
	params SaveGrammarLessonDraftParams,
) (models.GrammarLesson, error) {
	if r == nil || r.exec == nil {
		return models.GrammarLesson{}, errors.New("grammar lesson repository is not configured")
	}

	row := r.exec.QueryRowContext(ctx, `
		UPDATE grammar_lessons
		SET topic_id = $2,
			title = $3,
			slug = $4,
			summary = $5,
			status = $6,
			latest_import_id = $7,
			draft_revision_id = $8,
			active_publication_id = $9,
			updated_at = $10,
			version = version + 1
		WHERE id = $1
		RETURNING `+grammarLessonColumnsSQL,
		strings.TrimSpace(params.LessonID),
		strings.TrimSpace(params.TopicID),
		strings.TrimSpace(params.Title),
		strings.TrimSpace(params.Slug),
		nullableTrimmedString(params.Summary),
		strings.TrimSpace(params.Status),
		nullableStringPointer(params.LatestImportID),
		nullableStringPointer(params.DraftRevisionID),
		nullableStringPointer(params.ActivePublicationID),
		params.UpdatedAt.UTC(),
	)

	record, err := scanGrammarLesson(row)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return models.GrammarLesson{}, ErrConflict
		}
		return models.GrammarLesson{}, err
	}

	return record, nil
}

func (r *sqlGrammarLessonRepository) ApplyPublication(
	ctx context.Context,
	params ApplyGrammarLessonPublicationParams,
) (models.GrammarLesson, error) {
	if r == nil || r.exec == nil {
		return models.GrammarLesson{}, errors.New("grammar lesson repository is not configured")
	}

	row := r.exec.QueryRowContext(ctx, `
		UPDATE grammar_lessons
		SET topic_id = $2,
			title = $3,
			slug = $4,
			summary = $5,
			status = $6,
			active_publication_id = $7,
			updated_at = $8,
			version = version + 1
		WHERE id = $1
		RETURNING `+grammarLessonColumnsSQL,
		strings.TrimSpace(params.LessonID),
		strings.TrimSpace(params.TopicID),
		strings.TrimSpace(params.Title),
		strings.TrimSpace(params.Slug),
		nullableTrimmedString(params.Summary),
		strings.TrimSpace(params.Status),
		nullableStringPointer(params.ActivePublicationID),
		params.UpdatedAt.UTC(),
	)

	record, err := scanGrammarLesson(row)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return models.GrammarLesson{}, ErrConflict
		}
		return models.GrammarLesson{}, err
	}

	return record, nil
}

const grammarTopicColumnsSQL = `
	id,
	title,
	slug,
	COALESCE(description, ''),
	COALESCE(level, ''),
	sort_order,
	created_at,
	updated_at
`

const selectGrammarTopicsSQL = `
	SELECT ` + grammarTopicColumnsSQL + `
	FROM grammar_topics
`

const grammarLessonColumnsSQL = `
	id,
	topic_id,
	title,
	slug,
	COALESCE(summary, ''),
	status,
	latest_import_id,
	draft_revision_id,
	active_publication_id,
	sort_order,
	created_at,
	updated_at,
	version
`

const selectGrammarLessonsSQL = `
	SELECT ` + grammarLessonColumnsSQL + `
	FROM grammar_lessons
`

const grammarContentRevisionColumnsSQL = `
	id,
	entity_type,
	entity_id,
	version_no,
	import_id,
	source_type,
	source_payload,
	parsed_payload_json,
	keywords_json,
	COALESCE(render_artifact_url, ''),
	COALESCE(change_summary, ''),
	created_by,
	created_at
`

const selectGrammarContentRevisionsSQL = `
	SELECT ` + grammarContentRevisionColumnsSQL + `
	FROM content_revisions
`

func scanGrammarTopic(scanner rowScanner) (models.GrammarTopic, error) {
	var record models.GrammarTopic

	err := scanner.Scan(
		&record.ID,
		&record.Title,
		&record.Slug,
		&record.Description,
		&record.Level,
		&record.SortOrder,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.GrammarTopic{}, ErrNotFound
		}
		return models.GrammarTopic{}, fmt.Errorf("scan grammar topic: %w", err)
	}

	return record, nil
}

func scanGrammarLesson(scanner rowScanner) (models.GrammarLesson, error) {
	var (
		record              models.GrammarLesson
		latestImportID      sql.NullString
		draftRevisionID     sql.NullString
		activePublicationID sql.NullString
	)

	err := scanner.Scan(
		&record.ID,
		&record.TopicID,
		&record.Title,
		&record.Slug,
		&record.Summary,
		&record.Status,
		&latestImportID,
		&draftRevisionID,
		&activePublicationID,
		&record.SortOrder,
		&record.CreatedAt,
		&record.UpdatedAt,
		&record.Version,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.GrammarLesson{}, ErrNotFound
		}
		return models.GrammarLesson{}, fmt.Errorf("scan grammar lesson: %w", err)
	}

	if latestImportID.Valid {
		value := strings.TrimSpace(latestImportID.String)
		record.LatestImportID = &value
	}
	if draftRevisionID.Valid {
		value := strings.TrimSpace(draftRevisionID.String)
		record.DraftRevisionID = &value
	}
	if activePublicationID.Valid {
		value := strings.TrimSpace(activePublicationID.String)
		record.ActivePublicationID = &value
	}

	return record, nil
}

func scanGrammarContentRevision(scanner rowScanner) (models.ContentRevision, error) {
	var (
		record             models.ContentRevision
		importID           sql.NullString
		parsedPayloadRaw   []byte
		keywordsPayloadRaw []byte
	)

	err := scanner.Scan(
		&record.ID,
		&record.EntityType,
		&record.EntityID,
		&record.VersionNo,
		&importID,
		&record.SourceType,
		&record.SourcePayload,
		&parsedPayloadRaw,
		&keywordsPayloadRaw,
		&record.RenderArtifactURL,
		&record.ChangeSummary,
		&record.CreatedBy,
		&record.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.ContentRevision{}, ErrNotFound
		}
		return models.ContentRevision{}, fmt.Errorf("scan grammar content revision: %w", err)
	}

	if importID.Valid {
		value := strings.TrimSpace(importID.String)
		record.ImportID = &value
	}
	if len(parsedPayloadRaw) > 0 {
		record.ParsedPayloadJSON = map[string]any{}
		if err := json.Unmarshal(parsedPayloadRaw, &record.ParsedPayloadJSON); err != nil {
			return models.ContentRevision{}, fmt.Errorf("unmarshal grammar parsed payload json: %w", err)
		}
	}
	if len(keywordsPayloadRaw) > 0 {
		if err := json.Unmarshal(keywordsPayloadRaw, &record.KeywordsJSON); err != nil {
			return models.ContentRevision{}, fmt.Errorf("unmarshal grammar keywords json: %w", err)
		}
	}

	return record, nil
}
