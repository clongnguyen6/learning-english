package models

import "time"

const (
	ReadingDocumentStatusDraft     = "draft"
	ReadingDocumentStatusPublished = "published"
	ReadingDocumentStatusArchived  = "archived"
)

const (
	ReadingDisplayModeDocs = "docs"
)

const (
	ReadingHighlightColorYellow = "yellow"
)

type ReadingDocument struct {
	ID                  string
	Title               string
	Slug                string
	Description         string
	IsBilingual         bool
	DefaultDisplayMode  string
	Level               string
	Status              string
	LatestImportID      *string
	DraftRevisionID     *string
	ActivePublicationID *string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	Version             int
}

type ReadingSection struct {
	ID            string
	PublicationID string
	SectionOrder  int
	Heading       string
	ContentEN     string
	ContentVI     string
	KeywordsJSON  []string
	BlocksJSON    []any
	ContentHash   string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type ReadingHighlight struct {
	ID            string
	PublicationID string
	SectionOrder  int
	Keyword       string
	Color         string
	StartIndex    *int
	EndIndex      *int
	CreatedAt     time.Time
}

type ReadingProgress struct {
	ID                string
	UserID            string
	DocumentID        string
	LastSectionID     *string
	LastSectionOrder  int
	PublicationID     *string
	CompletionPercent float64
	LastReadAt        *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
	Version           int
}
