package entity

import "time"

const (
	JarAllocationNone    = "none"
	JarAllocationFixed   = "fixed"
	JarAllocationPercent = "percent"
)

type Jar struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	OwnerID   string    `json:"owner_id" gorm:"not null;uniqueIndex:jar_owner_id_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (Jar) TableName() string { return "jars" }

type JarMonth struct {
	OwnerID string       `json:"-" gorm:"primaryKey"`
	Month   CalendarDate `json:"month" gorm:"primaryKey;type:date"`
}

func (JarMonth) TableName() string { return "jar_months" }

type JarMonthConfig struct {
	OwnerID              string       `json:"-" gorm:"primaryKey"`
	Month                CalendarDate `json:"month" gorm:"primaryKey;type:date"`
	JarID                string       `json:"jar_id" gorm:"primaryKey"`
	Name                 string       `json:"name"`
	AllocationMode       string       `json:"allocation_mode"`
	AllocationAmount     *int64       `json:"allocation_amount,omitempty"`
	AllocationPercentBPS *int         `json:"-" gorm:"column:allocation_percent_bps"`
	Active               bool         `json:"active" gorm:"not null;default:true"`
	CreatedAt            time.Time    `json:"created_at"`
	UpdatedAt            time.Time    `json:"updated_at"`
}

func (JarMonthConfig) TableName() string { return "jar_month_configs" }

type JarMonthItem struct {
	JarID                string   `json:"jar_id"`
	Name                 string   `json:"name"`
	AllocationMode       string   `json:"allocation_mode"`
	AllocationAmount     *int64   `json:"allocation_amount,omitempty"`
	AllocationPercent    *float64 `json:"allocation_percent,omitempty"`
	CalculatedAllocation *int64   `json:"calculated_allocation,omitempty"`
	Spent                int64    `json:"spent"`
	Active               bool     `json:"active"`
}

type JarIdentity struct {
	JarID string `json:"jar_id"`
	Name  string `json:"name"`
}

type JarMonthSummary struct {
	Month                  string         `json:"month"`
	Timezone               string         `json:"timezone"`
	ActualIncome           int64          `json:"actual_income"`
	TotalSpent             int64          `json:"total_spent"`
	UnassignedSpent        int64          `json:"unassigned_spent"`
	TotalAllocated         int64          `json:"total_allocated"`
	TotalAllocationPercent float64        `json:"total_allocation_percent"`
	OverIncome             bool           `json:"over_income"`
	OverOneHundredPercent  bool           `json:"over_one_hundred_percent"`
	Jars                   []JarIdentity  `json:"jars"`
	Items                  []JarMonthItem `json:"items"`
}

type JarCumulativeMonth struct {
	Month                string `json:"month"`
	Name                 string `json:"name,omitempty"`
	Spent                int64  `json:"spent"`
	CalculatedAllocation *int64 `json:"calculated_allocation,omitempty"`
	AllocationCovered    bool   `json:"allocation_covered"`
}

type JarCumulativeSummary struct {
	JarID              string               `json:"jar_id"`
	Name               string               `json:"name"`
	FromMonth          string               `json:"from_month"`
	ToMonth            string               `json:"to_month"`
	TotalSpent         int64                `json:"total_spent"`
	TotalAllocated     int64                `json:"total_allocated"`
	AllocationMonths   int                  `json:"allocation_months"`
	MonthsInRange      int                  `json:"months_in_range"`
	AllocationVariance *int64               `json:"allocation_variance,omitempty"`
	Months             []JarCumulativeMonth `json:"months"`
}
