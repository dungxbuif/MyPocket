package entity

import "time"

type MonthNote struct {
	OwnerID   string       `json:"-" gorm:"primaryKey"`
	Month     CalendarDate `json:"month" gorm:"primaryKey;type:date"`
	Note      string       `json:"note"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

func (MonthNote) TableName() string { return "month_notes" }

type MonthCategoryTotal struct {
	CategoryID *string `json:"category_id,omitempty"`
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	Amount     int64   `json:"amount"`
	Count      int     `json:"count"`
}

type MonthSummary struct {
	Month            string               `json:"month"`
	Timezone         string               `json:"timezone"`
	StartAt          time.Time            `json:"start_at"`
	NextStartAt      time.Time            `json:"next_start_at"`
	IsCurrent        bool                 `json:"is_current"`
	IsComplete       bool                 `json:"is_complete"`
	Income           int64                `json:"income"`
	Expense          int64                `json:"expense"`
	Net              int64                `json:"net"`
	TransactionCount int                  `json:"transaction_count"`
	Categories       []MonthCategoryTotal `json:"categories"`
	Note             string               `json:"note"`
	JarSummary       *JarMonthSummary     `json:"jar_summary,omitempty"`
	CalculatedAt     time.Time            `json:"calculated_at"`
}
