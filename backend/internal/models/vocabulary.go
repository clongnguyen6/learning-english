package models

import "time"

const (
	VocabBookStatusDraft     = "draft"
	VocabBookStatusPublished = "published"
	VocabBookStatusArchived  = "archived"
)

const (
	VocabLearningStatusNew      = "new"
	VocabLearningStatusLearning = "learning"
	VocabLearningStatusMastered = "mastered"
)

const (
	VocabSessionModeGuessMeaning = "guess_meaning"
	VocabSessionModeLearnWord    = "learn_word"
)

const (
	VocabStudyModeFlip  = "flip"
	VocabStudyModeWrite = "write"
)

const (
	VocabDirectionENToVI = "en_vi"
	VocabDirectionVIToEN = "vi_en"
)

const (
	VocabOrderingRandom     = "random"
	VocabOrderingSequential = "sequential"
	VocabOrderingRepeat     = "repeat"
)

const (
	VocabQuestionTypeMCQ   = "mcq"
	VocabQuestionTypeWrite = "write"
	VocabQuestionTypeFlip  = "flip"
)

const (
	VocabSessionItemStatusPending  = "pending"
	VocabSessionItemStatusAnswered = "answered"
	VocabSessionItemStatusSkipped  = "skipped"
)

type VocabBook struct {
	ID                  string
	Title               string
	Slug                string
	Description         string
	CoverImageURL       string
	Level               string
	LanguageFrom        string
	LanguageTo          string
	Status              string
	LatestImportID      *string
	DraftRevisionID     *string
	ActivePublicationID *string
	WordCount           int
	SortOrder           int
	CreatedAt           time.Time
	UpdatedAt           time.Time
	Version             int
}

type VocabWord struct {
	ID             string
	BookID         string
	Word           string
	NormalizedWord string
	IPA            string
	PartOfSpeech   string
	MeaningVI      string
	MeaningEN      string
	Context        string
	Tips           string
	ImageURL       string
	StatusDefault  string
	SortOrder      int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type VocabWordExample struct {
	ID        string
	WordID    string
	ExampleEN string
	ExampleVI string
	SortOrder int
	CreatedAt time.Time
}

type VocabWordMedia struct {
	ID        string
	WordID    string
	MediaType string
	MediaURL  string
	SortOrder int
	CreatedAt time.Time
}

type UserVocabProgress struct {
	ID             string
	UserID         string
	WordID         string
	LearningStatus string
	LastStudiedAt  *time.Time
	CorrectCount   int
	WrongCount     int
	StreakCount    int
	NextReviewAt   *time.Time
	IsFavorite     bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Version        int
}

type VocabStudySession struct {
	ID                 string
	UserID             string
	BookID             string
	Mode               string
	StudyMode          string
	Direction          string
	Ordering           string
	ShowIPA            bool
	StatusFilter       string
	AutoNext           bool
	CurrentQuestionSeq int
	StartedAt          time.Time
	EndedAt            *time.Time
	TotalItems         int
	CompletedItems     int
	CreatedAt          time.Time
	UpdatedAt          time.Time
	Version            int
}

type VocabSessionItem struct {
	ID             string
	SessionID      string
	WordID         string
	QuestionSeq    int
	QuestionType   string
	PromptLanguage string
	PromptText     string
	Options        []string
	CorrectAnswer  string
	Status         string
	CreatedAt      time.Time
}

type VocabQuizAttempt struct {
	ID                    string
	SessionID             string
	SessionItemID         string
	UserID                string
	WordID                string
	QuestionType          string
	PromptLanguage        string
	AnswerText            string
	SelectedOptionIndex   *int
	IsCorrect             *bool
	AnsweredAt            time.Time
	RequestIDempotencyKey string
}
