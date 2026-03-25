package database

import (
	"context"
	"errors"
	"strings"
)

// Handle is the first backend-shell seam for database access.
// The real SQL/GORM connection will land in later beads without changing callers.
type Handle struct {
	dsn string
}

func New(dsn string) *Handle {
	return &Handle{dsn: strings.TrimSpace(dsn)}
}

func (h *Handle) Ping(_ context.Context) error {
	if h == nil || h.dsn == "" {
		return errors.New("database is not configured")
	}

	return nil
}
