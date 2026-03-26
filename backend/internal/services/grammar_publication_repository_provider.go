package services

import (
	"context"
	"database/sql"
	"errors"

	"learning-english/backend/internal/database"
	"learning-english/backend/internal/models"
	"learning-english/backend/internal/repositories"
)

type sqlGrammarPublicationRepositoryProvider struct {
	handle *database.Handle
}

type grammarImportStateRepositoryAdapter struct {
	repo repositories.GrammarImportStateRepository
}

type grammarContentRevisionRepositoryAdapter struct {
	repo repositories.GrammarContentRevisionRepository
}

type grammarTopicRepositoryAdapter struct {
	repo repositories.GrammarTopicRepository
}

type grammarLessonRepositoryAdapter struct {
	repo repositories.GrammarLessonRepository
}

func NewSQLGrammarPublicationRepositoryProvider(
	handle *database.Handle,
) GrammarPublicationRepositoryProvider {
	return &sqlGrammarPublicationRepositoryProvider{handle: handle}
}

func (p *sqlGrammarPublicationRepositoryProvider) ReadOnly() GrammarPublicationRepositorySet {
	if p == nil || p.handle == nil {
		return GrammarPublicationRepositorySet{}
	}

	return newGrammarPublicationRepositorySet(p.handle.DB())
}

func (p *sqlGrammarPublicationRepositoryProvider) WithinTx(
	ctx context.Context,
	fn func(GrammarPublicationRepositorySet) error,
) error {
	if p == nil || p.handle == nil {
		return errors.New("grammar publication repository provider is not configured")
	}
	if fn == nil {
		return errors.New("grammar publication repository callback is required")
	}

	return p.handle.WithTx(ctx, func(tx *sql.Tx) error {
		return fn(newGrammarPublicationRepositorySet(tx))
	})
}

func newGrammarPublicationRepositorySet(exec database.DBTX) GrammarPublicationRepositorySet {
	return GrammarPublicationRepositorySet{
		Imports:      repositories.NewImportRepository(exec),
		ImportStates: grammarImportStateRepositoryAdapter{repo: repositories.NewGrammarImportStateRepository(exec)},
		Revisions:    grammarContentRevisionRepositoryAdapter{repo: repositories.NewGrammarContentRevisionRepository(exec)},
		Publications: repositories.NewPublicationRepository(exec),
		AuditLogs:    repositories.NewContentAuditLogRepository(exec),
		Entities:     repositories.NewContentEntityRepository(exec),
		Topics:       grammarTopicRepositoryAdapter{repo: repositories.NewGrammarTopicRepository(exec)},
		Lessons:      grammarLessonRepositoryAdapter{repo: repositories.NewGrammarLessonRepository(exec)},
	}
}

func (a grammarImportStateRepositoryAdapter) MarkCommitted(
	ctx context.Context,
	importID string,
) (models.Import, error) {
	return a.repo.MarkCommitted(ctx, importID)
}

func (a grammarContentRevisionRepositoryAdapter) FindByID(
	ctx context.Context,
	revisionID string,
) (models.ContentRevision, error) {
	return a.repo.FindByID(ctx, revisionID)
}

func (a grammarContentRevisionRepositoryAdapter) NextVersionNo(
	ctx context.Context,
	entityType, entityID string,
) (int, error) {
	return a.repo.NextVersionNo(ctx, entityType, entityID)
}

func (a grammarContentRevisionRepositoryAdapter) Create(
	ctx context.Context,
	params CreateGrammarContentRevisionParams,
) (models.ContentRevision, error) {
	return a.repo.Create(ctx, repositories.CreateGrammarContentRevisionParams{
		EntityType:        params.EntityType,
		EntityID:          params.EntityID,
		VersionNo:         params.VersionNo,
		ImportID:          params.ImportID,
		SourceType:        params.SourceType,
		SourcePayload:     params.SourcePayload,
		ParsedPayloadJSON: cloneMap(params.ParsedPayloadJSON),
		KeywordsJSON:      append([]string(nil), params.KeywordsJSON...),
		RenderArtifactURL: params.RenderArtifactURL,
		ChangeSummary:     params.ChangeSummary,
		CreatedBy:         params.CreatedBy,
		CreatedAt:         params.CreatedAt,
	})
}

func (a grammarTopicRepositoryAdapter) FindBySlug(
	ctx context.Context,
	slug string,
) (models.GrammarTopic, error) {
	return a.repo.FindBySlug(ctx, slug)
}

func (a grammarTopicRepositoryAdapter) Create(
	ctx context.Context,
	params CreateGrammarTopicParams,
) (models.GrammarTopic, error) {
	return a.repo.Create(ctx, repositories.CreateGrammarTopicParams{
		Title:       params.Title,
		Slug:        params.Slug,
		Description: params.Description,
		Level:       params.Level,
		SortOrder:   params.SortOrder,
		CreatedAt:   params.CreatedAt,
		UpdatedAt:   params.UpdatedAt,
	})
}

func (a grammarLessonRepositoryAdapter) FindBySlug(
	ctx context.Context,
	slug string,
) (models.GrammarLesson, error) {
	return a.repo.FindBySlug(ctx, slug)
}

func (a grammarLessonRepositoryAdapter) Create(
	ctx context.Context,
	params CreateGrammarLessonParams,
) (models.GrammarLesson, error) {
	return a.repo.Create(ctx, repositories.CreateGrammarLessonParams{
		TopicID:        params.TopicID,
		Title:          params.Title,
		Slug:           params.Slug,
		Summary:        params.Summary,
		Status:         params.Status,
		LatestImportID: params.LatestImportID,
		SortOrder:      params.SortOrder,
		CreatedAt:      params.CreatedAt,
		UpdatedAt:      params.UpdatedAt,
	})
}

func (a grammarLessonRepositoryAdapter) SaveDraft(
	ctx context.Context,
	params SaveGrammarLessonDraftParams,
) (models.GrammarLesson, error) {
	return a.repo.SaveDraft(ctx, repositories.SaveGrammarLessonDraftParams{
		LessonID:            params.LessonID,
		TopicID:             params.TopicID,
		Title:               params.Title,
		Slug:                params.Slug,
		Summary:             params.Summary,
		Status:              params.Status,
		LatestImportID:      params.LatestImportID,
		DraftRevisionID:     params.DraftRevisionID,
		ActivePublicationID: params.ActivePublicationID,
		UpdatedAt:           params.UpdatedAt,
	})
}

func (a grammarLessonRepositoryAdapter) ApplyPublication(
	ctx context.Context,
	params ApplyGrammarLessonPublicationParams,
) (models.GrammarLesson, error) {
	return a.repo.ApplyPublication(ctx, repositories.ApplyGrammarLessonPublicationParams{
		LessonID:            params.LessonID,
		TopicID:             params.TopicID,
		Title:               params.Title,
		Slug:                params.Slug,
		Summary:             params.Summary,
		Status:              params.Status,
		ActivePublicationID: params.ActivePublicationID,
		UpdatedAt:           params.UpdatedAt,
	})
}
