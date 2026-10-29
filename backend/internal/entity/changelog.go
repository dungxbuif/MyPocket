package entity

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	ErrChangelogVersionRequired = errors.New("changelog version is required")
	ErrChangelogTitleRequired   = errors.New("changelog title is required")
	ErrChangelogDescription     = errors.New("changelog description is invalid")
)

type Changelog struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Version     string    `json:"version" gorm:"uniqueIndex;not null"`
	Title       string    `json:"title" gorm:"not null"`
	Description string    `json:"description" gorm:"not null"`
	PublishedAt time.Time `json:"published_at"`
	CreatedAt   time.Time `json:"created_at"`
}

func (Changelog) TableName() string { return "changelogs" }

func (c Changelog) Validate() error {
	if strings.TrimSpace(c.Version) == "" || utf8.RuneCountInString(strings.TrimSpace(c.Version)) > 50 {
		return ErrChangelogVersionRequired
	}
	if strings.TrimSpace(c.Title) == "" || utf8.RuneCountInString(strings.TrimSpace(c.Title)) > 200 {
		return ErrChangelogTitleRequired
	}
	if strings.TrimSpace(c.Description) == "" || utf8.RuneCountInString(strings.TrimSpace(c.Description)) > 10000 {
		return ErrChangelogDescription
	}
	return nil
}
