package entity

import (
	"encoding/json"
	"time"
)

type ReportSnapshot struct {
	ID           string             `json:"id" gorm:"primaryKey"`
	OwnerID      string             `json:"-" gorm:"index;not null"`
	Scope        MonthlyReportScope `json:"scope" gorm:"-"`
	Timezone     string             `json:"timezone"`
	AsOfDay      string             `json:"as_of_day"`
	DataRevision int64              `json:"data_revision"`
	RuleVersion  string             `json:"rule_version"`
	AsOf         time.Time          `json:"as_of"`
	ExpiresAt    time.Time          `json:"expires_at"`
	CreatedAt    time.Time          `json:"created_at"`
	Payload      json.RawMessage    `json:"-"`
	Report       MonthlyReport      `json:"-"`
}

func (ReportSnapshot) TableName() string { return "report_snapshots" }
