package entity

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	FeedbackTypeBug         = "bug"
	FeedbackTypeFeature     = "feature"
	FeedbackTypeImprovement = "improvement"

	FeedbackStatusOpen       = "open"
	FeedbackStatusTriaged    = "triaged"
	FeedbackStatusInProgress = "in_progress"
	FeedbackStatusFixed      = "fixed"
	FeedbackStatusRejected   = "rejected"
)

var (
	ErrFeedbackTypeInvalid   = errors.New("feedback type is invalid")
	ErrFeedbackStatusInvalid = errors.New("feedback status is invalid")
	ErrFeedbackTitleRequired = errors.New("feedback title is required")
	ErrFeedbackDescription   = errors.New("feedback description is invalid")
)

type Feedback struct {
	ID                  string     `json:"id" gorm:"primaryKey"`
	UserID              string     `json:"-" gorm:"column:user_id;index;not null"`
	Type                string     `json:"type" gorm:"not null"`
	Title               string     `json:"title" gorm:"not null"`
	Description         string     `json:"description" gorm:"not null"`
	Status              string     `json:"status" gorm:"not null;default:open"`
	FixedAt             *time.Time `json:"fixed_at,omitempty"`
	ChangelogID         *string    `json:"changelog_id,omitempty" gorm:"column:changelog_id;index"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	ScreenshotObjectKey string     `json:"-" gorm:"column:screenshot_object_key"`
	ScreenshotMIMEType  string     `json:"-" gorm:"column:screenshot_mime_type"`
	ScreenshotSizeBytes int64      `json:"-" gorm:"column:screenshot_size_bytes"`
	ScreenshotCreatedAt *time.Time `json:"-" gorm:"column:screenshot_created_at"`
	ScreenshotAvailable bool       `json:"screenshot_available" gorm:"-"`
}

func (Feedback) TableName() string { return "feedback" }

func (f Feedback) Validate() error {
	if f.Type != FeedbackTypeBug && f.Type != FeedbackTypeFeature && f.Type != FeedbackTypeImprovement {
		return ErrFeedbackTypeInvalid
	}
	if strings.TrimSpace(f.Title) == "" || utf8.RuneCountInString(strings.TrimSpace(f.Title)) > 200 {
		return ErrFeedbackTitleRequired
	}
	if strings.TrimSpace(f.Description) == "" || utf8.RuneCountInString(strings.TrimSpace(f.Description)) > 10000 {
		return ErrFeedbackDescription
	}
	status := f.Status
	if status == "" {
		status = FeedbackStatusOpen
	}
	if !validFeedbackStatus(status) {
		return ErrFeedbackStatusInvalid
	}
	return nil
}

func validFeedbackStatus(status string) bool {
	switch status {
	case FeedbackStatusOpen, FeedbackStatusTriaged, FeedbackStatusInProgress, FeedbackStatusFixed, FeedbackStatusRejected:
		return true
	default:
		return false
	}
}

func CanTransitionFeedback(from, to string) bool {
	switch from {
	case FeedbackStatusOpen:
		return to == FeedbackStatusTriaged
	case FeedbackStatusTriaged:
		return to == FeedbackStatusInProgress
	case FeedbackStatusInProgress:
		return to == FeedbackStatusFixed || to == FeedbackStatusRejected
	default:
		return false
	}
}
