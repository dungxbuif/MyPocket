package entity

import "time"

type RecurringSchedule struct {
	ID          string     `json:"id" gorm:"primaryKey"`
	OwnerID     string     `json:"owner_id" gorm:"index;not null"`
	Name        string     `json:"name" gorm:"not null"`
	WalletID    string     `json:"wallet_id" gorm:"index;not null"`
	CategoryID  *string    `json:"category_id,omitempty" gorm:"index"`
	Type        string     `json:"type" gorm:"not null"`
	Amount      int64      `json:"amount" gorm:"not null"`
	Note        *string    `json:"note,omitempty"`
	Frequency   string     `json:"frequency" gorm:"not null"`
	Interval    int        `json:"interval" gorm:"not null;default:1"`
	NextRunAt   time.Time  `json:"next_run_at" gorm:"index;not null"`
	EndsAt      *time.Time `json:"ends_at,omitempty"`
	AnchorDay   int        `json:"anchor_day" gorm:"not null"`
	AnchorMonth int        `json:"anchor_month" gorm:"not null"`
	Active      bool       `json:"active" gorm:"not null;default:true"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type RecurringOccurrence struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	ScheduleID    string    `json:"schedule_id" gorm:"uniqueIndex:recurring_occurrence_due"`
	DueAt         time.Time `json:"due_at" gorm:"uniqueIndex:recurring_occurrence_due"`
	TransactionID string    `json:"transaction_id" gorm:"index;not null"`
	CreatedAt     time.Time `json:"created_at"`
}
