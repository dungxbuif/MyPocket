package repository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	contract "github.com/mypocket/backend/internal/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type JarPostgresRepository struct{ db *gorm.DB }

func NewJarPostgresRepository(db *gorm.DB) contract.JarRepository {
	return &JarPostgresRepository{db: db}
}

func parseJarPeriod(month, timezone string) (entity.CalendarDate, *time.Location, error) {
	label, err := entity.ParseMonth(month)
	if err != nil {
		return entity.CalendarDate{}, nil, err
	}
	location, err := time.LoadLocation(timezone)
	if err != nil || timezone == "Local" {
		return entity.CalendarDate{}, nil, fmt.Errorf("invalid account timezone")
	}
	return entity.CalendarDate{Time: label}, location, nil
}

func (r *JarPostgresRepository) lockOwner(tx *gorm.DB, owner string) error {
	var user entity.User
	return tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", owner).First(&user).Error
}

func (r *JarPostgresRepository) ensureMonth(tx *gorm.DB, owner string, month entity.CalendarDate) error {
	result := tx.Exec("INSERT INTO jar_months(owner_id, month) VALUES (?, ?) ON CONFLICT DO NOTHING", owner, month)
	if result.Error != nil || result.RowsAffected == 0 {
		return result.Error
	}
	var previous entity.CalendarDate
	err := tx.Model(&entity.JarMonth{}).Select("month").Where("owner_id = ? AND month < ?", owner, month).Order("month DESC").Limit(1).Scan(&previous).Error
	if err != nil || previous.IsZero() {
		return err
	}
	return tx.Exec(`INSERT INTO jar_month_configs(owner_id, month, jar_id, name, allocation_mode, allocation_amount, allocation_percent_bps, active)
		SELECT owner_id, ?, jar_id, name, allocation_mode, allocation_amount, allocation_percent_bps, true
		FROM jar_month_configs WHERE owner_id = ? AND month = ? AND active = true`, month, owner, previous).Error
}

func (r *JarPostgresRepository) ListMonth(owner, monthLabel, timezone string) (*entity.JarMonthSummary, error) {
	month, location, err := parseJarPeriod(monthLabel, timezone)
	if err != nil {
		return nil, contract.ErrJarInvalid
	}
	start, next, _ := entity.MonthRangeUTC(month.Time, location)
	var summary entity.JarMonthSummary
	err = r.db.Transaction(func(tx *gorm.DB) error {
		if err := r.lockOwner(tx, owner); err != nil {
			return err
		}
		if err := r.ensureMonth(tx, owner, month); err != nil {
			return err
		}
		summary.Month, summary.Timezone = monthLabel, timezone
		if err := tx.Raw(`SELECT j.id AS jar_id,
			COALESCE((SELECT c.name FROM jar_month_configs c WHERE c.owner_id = j.owner_id AND c.jar_id = j.id ORDER BY c.month DESC LIMIT 1), 'Hũ đã gỡ') AS name
			FROM jars j WHERE j.owner_id = ? ORDER BY name, j.id`, owner).Scan(&summary.Jars).Error; err != nil {
			return err
		}
		if err := tx.Raw(`SELECT COALESCE(SUM(t.amount), 0) FROM transactions t
			LEFT JOIN categories c ON c.id = t.category_id
			WHERE t.owner_id = ? AND t.type = 'income' AND t.included_in_reports = true
			AND COALESCE(c.system_key, '') <> 'income_transfer_in' AND t.occurred_at >= ? AND t.occurred_at < ?`, owner, start, next).Scan(&summary.ActualIncome).Error; err != nil {
			return err
		}
		if err := tx.Raw(`SELECT COALESCE(SUM(t.amount), 0) FROM transactions t
			LEFT JOIN categories c ON c.id = t.category_id
			WHERE t.owner_id = ? AND t.type = 'expense' AND t.included_in_reports = true
			AND COALESCE(c.system_key, '') <> 'expense_transfer_out' AND t.occurred_at >= ? AND t.occurred_at < ?`, owner, start, next).Scan(&summary.TotalSpent).Error; err != nil {
			return err
		}
		if err := tx.Raw(`SELECT COALESCE(SUM(t.amount), 0) FROM transactions t
			LEFT JOIN categories c ON c.id = t.category_id
			WHERE t.owner_id = ? AND t.type = 'expense' AND t.included_in_reports = true
			AND t.jar_id IS NULL AND COALESCE(c.system_key, '') <> 'expense_transfer_out' AND t.occurred_at >= ? AND t.occurred_at < ?`, owner, start, next).Scan(&summary.UnassignedSpent).Error; err != nil {
			return err
		}
		var spentRows []struct {
			JarID string `gorm:"column:jar_id"`
			Spent int64  `gorm:"column:spent"`
		}
		if err := tx.Raw(`SELECT t.jar_id, COALESCE(SUM(t.amount), 0) AS spent FROM transactions t
			LEFT JOIN categories c ON c.id = t.category_id
			WHERE t.owner_id = ? AND t.type = 'expense' AND t.included_in_reports = true AND t.jar_id IS NOT NULL
			AND COALESCE(c.system_key, '') <> 'expense_transfer_out' AND t.occurred_at >= ? AND t.occurred_at < ?
			GROUP BY t.jar_id`, owner, start, next).Scan(&spentRows).Error; err != nil {
			return err
		}
		spent := make(map[string]int64, len(spentRows))
		for _, row := range spentRows {
			spent[row.JarID] = row.Spent
		}
		var configs []entity.JarMonthConfig
		if err := tx.Where("owner_id = ? AND month = ?", owner, month).Order("active DESC, created_at ASC").Find(&configs).Error; err != nil {
			return err
		}
		for _, config := range configs {
			amount := spent[config.JarID]
			if !config.Active && amount == 0 {
				continue
			}
			item := entity.JarMonthItem{JarID: config.JarID, Name: config.Name, AllocationMode: config.AllocationMode, AllocationAmount: config.AllocationAmount, Spent: amount, Active: config.Active}
			switch config.AllocationMode {
			case entity.JarAllocationFixed:
				value := *config.AllocationAmount
				item.CalculatedAllocation = &value
			case entity.JarAllocationPercent:
				basisPoints := *config.AllocationPercentBPS
				percent := float64(basisPoints) / 100
				item.AllocationPercent = &percent
				value := summary.ActualIncome/10000*int64(basisPoints) + summary.ActualIncome%10000*int64(basisPoints)/10000
				item.CalculatedAllocation = &value
				summary.TotalAllocationPercent += percent
			}
			if item.CalculatedAllocation != nil {
				summary.TotalAllocated += *item.CalculatedAllocation
			}
			summary.Items = append(summary.Items, item)
		}
		configured := make(map[string]struct{}, len(configs))
		for _, config := range configs {
			configured[config.JarID] = struct{}{}
		}
		for _, row := range spentRows {
			if _, ok := configured[row.JarID]; ok {
				continue
			}
			var history []entity.JarMonthConfig
			if err := tx.Where("owner_id = ? AND jar_id = ?", owner, row.JarID).Find(&history).Error; err != nil {
				return err
			}
			name := nearestJarConfigName(history, month.Time)
			if name == "" {
				name = "Hũ không có cấu hình tháng"
			}
			summary.Items = append(summary.Items, entity.JarMonthItem{JarID: row.JarID, Name: name, AllocationMode: entity.JarAllocationNone, Spent: row.Spent, Active: false})
		}
		summary.OverIncome = summary.TotalAllocated > summary.ActualIncome
		summary.OverOneHundredPercent = summary.TotalAllocationPercent > 100
		return nil
	})
	if err != nil {
		return nil, err
	}
	if summary.Jars == nil {
		summary.Jars = []entity.JarIdentity{}
	}
	if summary.Items == nil {
		summary.Items = []entity.JarMonthItem{}
	}
	return &summary, nil
}

func nearestJarConfigName(configs []entity.JarMonthConfig, target time.Time) string {
	bestDistance := int(^uint(0) >> 1)
	bestMonth := time.Time{}
	name := ""
	targetIndex := target.Year()*12 + int(target.Month())
	for _, config := range configs {
		month := config.Month.Time
		index := month.Year()*12 + int(month.Month())
		distance := index - targetIndex
		if distance < 0 {
			distance = -distance
		}
		if distance < bestDistance || (distance == bestDistance && (bestMonth.IsZero() || month.Before(bestMonth))) {
			bestDistance, bestMonth, name = distance, month, config.Name
		}
	}
	return name
}

func (r *JarPostgresRepository) Create(owner, monthLabel, timezone, name, mode string, amount *int64, percentBPS *int) (*entity.JarMonthConfig, error) {
	month, _, err := parseJarPeriod(monthLabel, timezone)
	if err != nil {
		return nil, contract.ErrJarInvalid
	}
	config := &entity.JarMonthConfig{OwnerID: owner, Month: month, JarID: uuid.NewString(), Name: name, AllocationMode: mode, AllocationAmount: amount, AllocationPercentBPS: percentBPS, Active: true}
	err = r.db.Transaction(func(tx *gorm.DB) error {
		if err := r.lockOwner(tx, owner); err != nil {
			return err
		}
		if err := r.ensureMonth(tx, owner, month); err != nil {
			return err
		}
		jar := entity.Jar{ID: config.JarID, OwnerID: owner}
		if err := tx.Create(&jar).Error; err != nil {
			return err
		}
		return tx.Create(config).Error
	})
	return config, err
}

func (r *JarPostgresRepository) UpdateMonthConfig(owner, jarID, monthLabel, name, mode string, amount *int64, percentBPS *int) (*entity.JarMonthConfig, error) {
	month, _, err := parseJarPeriod(monthLabel, "UTC")
	if err != nil {
		return nil, contract.ErrJarInvalid
	}
	var config entity.JarMonthConfig
	err = r.db.Transaction(func(tx *gorm.DB) error {
		if err := r.lockOwner(tx, owner); err != nil {
			return err
		}
		result := tx.Model(&entity.JarMonthConfig{}).Where("owner_id = ? AND jar_id = ? AND month = ? AND active = true", owner, jarID, month).Updates(map[string]any{
			"name": name, "allocation_mode": mode, "allocation_amount": amount, "allocation_percent_bps": percentBPS, "updated_at": time.Now().UTC(),
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return tx.Where("owner_id = ? AND jar_id = ? AND month = ?", owner, jarID, month).First(&config).Error
	})
	return &config, err
}

func (r *JarPostgresRepository) RemoveMonthConfig(owner, jarID, monthLabel string) error {
	month, _, err := parseJarPeriod(monthLabel, "UTC")
	if err != nil {
		return contract.ErrJarInvalid
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := r.lockOwner(tx, owner); err != nil {
			return err
		}
		result := tx.Model(&entity.JarMonthConfig{}).Where("owner_id = ? AND jar_id = ? AND month = ? AND active = true", owner, jarID, month).Updates(map[string]any{"active": false, "updated_at": time.Now().UTC()})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func (r *JarPostgresRepository) FindMonthConfig(owner, jarID, monthLabel string, includeInactive bool) (*entity.JarMonthConfig, error) {
	month, _, err := parseJarPeriod(monthLabel, "UTC")
	if err != nil {
		return nil, contract.ErrJarInvalid
	}
	query := r.db.Where("owner_id = ? AND jar_id = ? AND month = ?", owner, jarID, month)
	if !includeInactive {
		query = query.Where("active = true")
	}
	var config entity.JarMonthConfig
	if err := query.First(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *JarPostgresRepository) Cumulative(owner, jarID, fromLabel, toLabel, timezone string) (*entity.JarCumulativeSummary, error) {
	location, err := time.LoadLocation(timezone)
	if err != nil || timezone == "Local" {
		return nil, contract.ErrJarInvalid
	}
	var jar entity.Jar
	if err := r.db.Where("owner_id = ? AND id = ?", owner, jarID).First(&jar).Error; err != nil {
		return nil, err
	}
	var allConfigs []entity.JarMonthConfig
	if err := r.db.Where("owner_id = ? AND jar_id = ?", owner, jarID).Order("month ASC").Find(&allConfigs).Error; err != nil {
		return nil, err
	}
	if toLabel == "" {
		toLabel = time.Now().In(location).Format("2006-01")
	}
	to, _, err := parseJarPeriod(toLabel, timezone)
	if err != nil {
		return nil, contract.ErrJarInvalid
	}
	if fromLabel == "" {
		var earliest time.Time
		for _, config := range allConfigs {
			if earliest.IsZero() || config.Month.Time.Before(earliest) {
				earliest = config.Month.Time
			}
		}
		var earliestSpend entity.CalendarDate
		if err := r.db.Raw(`SELECT MIN(date_trunc('month', t.occurred_at AT TIME ZONE ?)::date)
			FROM transactions t LEFT JOIN categories c ON c.id = t.category_id
			WHERE t.owner_id = ? AND t.jar_id = ? AND t.type = 'expense' AND t.included_in_reports = true
			AND COALESCE(c.system_key, '') <> 'expense_transfer_out'`, timezone, owner, jarID).Scan(&earliestSpend).Error; err != nil {
			return nil, err
		}
		if !earliestSpend.IsZero() && (earliest.IsZero() || earliestSpend.Time.Before(earliest)) {
			earliest = earliestSpend.Time
		}
		if earliest.IsZero() {
			return nil, gorm.ErrRecordNotFound
		}
		fromLabel = earliest.Format("2006-01")
	}
	from, _, err := parseJarPeriod(fromLabel, timezone)
	if err != nil || to.Time.Before(from.Time) {
		return nil, contract.ErrJarInvalid
	}
	first, _, _ := entity.MonthRangeUTC(from.Time, location)
	_, last, _ := entity.MonthRangeUTC(to.Time, location)
	var summary entity.JarCumulativeSummary
	summary.JarID, summary.FromMonth, summary.ToMonth = jarID, fromLabel, toLabel
	var configs []entity.JarMonthConfig
	if err := r.db.Where("owner_id = ? AND jar_id = ? AND month >= ? AND month <= ?", owner, jarID, from, to).Order("month ASC").Find(&configs).Error; err != nil {
		return nil, err
	}
	var spendRows []struct {
		Month string `gorm:"column:month"`
		Spent int64  `gorm:"column:spent"`
	}
	if err := r.db.Raw(`SELECT to_char(date_trunc('month', t.occurred_at AT TIME ZONE ?), 'YYYY-MM') AS month, SUM(t.amount) AS spent
		FROM transactions t LEFT JOIN categories c ON c.id = t.category_id
		WHERE t.owner_id = ? AND t.jar_id = ? AND t.type = 'expense' AND t.included_in_reports = true
		AND COALESCE(c.system_key, '') <> 'expense_transfer_out' AND t.occurred_at >= ? AND t.occurred_at < ?
		GROUP BY 1`, timezone, owner, jarID, first, last).Scan(&spendRows).Error; err != nil {
		return nil, err
	}
	configByMonth := make(map[string]entity.JarMonthConfig, len(configs))
	nameByMonth := make(map[string]string, len(configs))
	for _, config := range configs {
		key := config.Month.String()[:7]
		configByMonth[key] = config
		nameByMonth[key] = config.Name
		summary.Name = config.Name
	}
	if summary.Name == "" && len(allConfigs) > 0 {
		summary.Name = allConfigs[len(allConfigs)-1].Name
	}
	spentByMonth := make(map[string]int64, len(spendRows))
	for _, row := range spendRows {
		spentByMonth[row.Month] = row.Spent
		summary.TotalSpent += row.Spent
	}
	for month := from.Time; !month.After(to.Time); month = month.AddDate(0, 1, 0) {
		key := month.Format("2006-01")
		config, configured := configByMonth[key]
		item := entity.JarCumulativeMonth{Month: key, Name: nameByMonth[key], Spent: spentByMonth[key]}
		if configured && config.AllocationMode != entity.JarAllocationNone {
			item.AllocationCovered = true
			summary.AllocationMonths++
			switch config.AllocationMode {
			case entity.JarAllocationFixed:
				value := *config.AllocationAmount
				item.CalculatedAllocation = &value
			case entity.JarAllocationPercent:
				var income int64
				monthStart, monthNext, _ := entity.MonthRangeUTC(month, location)
				if err := r.db.Raw(`SELECT COALESCE(SUM(t.amount), 0) FROM transactions t LEFT JOIN categories c ON c.id = t.category_id
					WHERE t.owner_id = ? AND t.type = 'income' AND t.included_in_reports = true AND COALESCE(c.system_key, '') <> 'income_transfer_in'
					AND t.occurred_at >= ? AND t.occurred_at < ?`, owner, monthStart, monthNext).Scan(&income).Error; err != nil {
					return nil, err
				}
				bps := int64(*config.AllocationPercentBPS)
				value := income/10000*bps + income%10000*bps/10000
				item.CalculatedAllocation = &value
			}
			summary.TotalAllocated += *item.CalculatedAllocation
		}
		summary.Months = append(summary.Months, item)
		summary.MonthsInRange++
	}
	if summary.AllocationMonths > 0 {
		coveredSpend := int64(0)
		for _, month := range summary.Months {
			if month.AllocationCovered {
				coveredSpend += month.Spent
			}
		}
		variance := summary.TotalAllocated - coveredSpend
		summary.AllocationVariance = &variance
	}
	if summary.Name == "" {
		summary.Name = "Hũ đã gỡ"
	}
	return &summary, nil
}
