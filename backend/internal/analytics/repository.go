package analytics

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"time"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func NormalizeFilter(from, to time.Time) Filter {
	loc := hoChiMinhLocation()
	now := time.Now().In(loc)
	if from.IsZero() {
		from = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	}
	if to.IsZero() {
		to = time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 999999999, loc)
	}
	return Filter{From: from.In(loc), To: to.In(loc)}
}

func hoChiMinhLocation() *time.Location {
	if loc, err := time.LoadLocation("Asia/Ho_Chi_Minh"); err == nil {
		return loc
	}
	// Minimal Alpine images may not ship tzdata; Vietnam has no DST.
	return time.FixedZone("Asia/Ho_Chi_Minh", 7*60*60)
}

func (r *Repository) Summary(ctx context.Context, userID string, filter Filter) (Summary, error) {
	var s Summary
	s.GeneratedAt = time.Now().UTC()
	s.Timezone = "Asia/Ho_Chi_Minh"
	s.From = filter.From.Format("2006-01-02")
	s.To = filter.To.Format("2006-01-02")
	err := r.db.QueryRowContext(ctx, `SELECT COALESCE(sum(CASE WHEN type='income' THEN amount_vnd ELSE 0 END),0), COALESCE(sum(CASE WHEN type='expense' THEN amount_vnd ELSE 0 END),0), COALESCE(max(version),0) FROM transactions WHERE user_id=$1 AND archived_at IS NULL AND excluded_from_reports=false AND type IN ('income','expense') AND occurred_at >= $2 AND occurred_at <= $3 AND ($4='' OR source_wallet_id::text=$4 OR destination_wallet_id::text=$4)`, userID, filter.From.UTC(), filter.To.UTC(), filter.WalletID).Scan(&s.IncomeVND, &s.ExpenseVND, &s.DataVersion)
	if err != nil {
		return Summary{}, fmt.Errorf("summary: %w", err)
	}
	s.NetIncomeVND = s.IncomeVND - s.ExpenseVND
	days := int(filter.To.Sub(filter.From).Hours()/24) + 1
	if days < 1 {
		days = 1
	}
	s.DailyAverageVND = s.ExpenseVND / int64(days)
	return s, nil
}

func (r *Repository) Categories(ctx context.Context, userID string, filter Filter) ([]CategoryTotal, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT COALESCE(parent.id::text, c.id::text, ''), COALESCE(parent.name, c.name, 'Chưa phân loại'), COALESCE(sum(t.amount_vnd),0) FROM transactions t LEFT JOIN categories c ON c.id=t.category_id LEFT JOIN categories parent ON parent.id=c.parent_id WHERE t.user_id=$1 AND t.archived_at IS NULL AND t.excluded_from_reports=false AND t.type='expense' AND t.occurred_at >= $2 AND t.occurred_at <= $3 AND ($4='' OR t.source_wallet_id::text=$4 OR t.destination_wallet_id::text=$4) GROUP BY COALESCE(parent.id::text,c.id::text,''), COALESCE(parent.name,c.name,'Chưa phân loại') ORDER BY 3 DESC`, userID, filter.From.UTC(), filter.To.UTC(), filter.WalletID)
	if err != nil {
		return nil, fmt.Errorf("category report: %w", err)
	}
	defer rows.Close()
	result := []CategoryTotal{}
	var total int64
	for rows.Next() {
		var c CategoryTotal
		if err := rows.Scan(&c.CategoryID, &c.CategoryName, &c.AmountVND); err != nil {
			return nil, err
		}
		total += c.AmountVND
		result = append(result, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range result {
		if total > 0 {
			result[i].SharePercent = float64(result[i].AmountVND) * 100 / float64(total)
		}
	}
	if len(result) == 0 {
		result = append(result, CategoryTotal{CategoryName: "Chưa phân loại"})
	}
	return result, nil
}

func (r *Repository) Insider(ctx context.Context, userID string, filter Filter) (InsiderReport, error) {
	result := InsiderReport{
		GeneratedAt: time.Now().UTC(), Timezone: "Asia/Ho_Chi_Minh",
		From: filter.From.Format("2006-01-02"), To: filter.To.Format("2006-01-02"),
		NotComparable: true,
	}
	now := time.Now().In(filter.From.Location())
	result.ElapsedDays = elapsedPeriodDays(filter, now)
	var category InsiderCategory
	err := r.db.QueryRowContext(ctx, `SELECT COALESCE(parent.id::text,c.id::text,''), COALESCE(parent.name,c.name,'Chưa phân loại'), count(*), COALESCE(sum(t.amount_vnd),0), COALESCE(max(t.version),0)
		FROM transactions t
		LEFT JOIN categories c ON c.id=t.category_id
		LEFT JOIN categories parent ON parent.id=c.parent_id
		WHERE t.user_id=$1 AND t.archived_at IS NULL AND t.excluded_from_reports=false AND t.type='expense'
		AND t.occurred_at >= $2 AND t.occurred_at <= $3
		AND ($4='' OR t.source_wallet_id::text=$4 OR t.destination_wallet_id::text=$4)
		GROUP BY COALESCE(parent.id::text,c.id::text,''), COALESCE(parent.name,c.name,'Chưa phân loại')
		ORDER BY count(*) DESC, sum(t.amount_vnd) DESC, COALESCE(parent.id::text,c.id::text,'')
		LIMIT 1`, userID, filter.From.UTC(), filter.To.UTC(), filter.WalletID).Scan(&category.ID, &category.Name, &category.TransactionCount, &result.SpentVND, &result.DataVersion)
	if errors.Is(err, sql.ErrNoRows) {
		return result, nil
	}
	if err != nil {
		return InsiderReport{}, fmt.Errorf("insider category: %w", err)
	}
	result.SelectedCategory = &category
	result.AverageDailyVND = result.SpentVND / int64(result.ElapsedDays)

	span := filter.To.Sub(filter.From) + time.Nanosecond
	priorFrom, priorTo := filter.From.Add(-span), filter.To.Add(-span)
	priorDays := elapsedPeriodDays(Filter{From: priorFrom, To: priorTo}, priorTo)
	var priorSpent int64
	err = r.db.QueryRowContext(ctx, `SELECT COALESCE(sum(t.amount_vnd),0)
		FROM transactions t
		LEFT JOIN categories c ON c.id=t.category_id
		LEFT JOIN categories parent ON parent.id=c.parent_id
		WHERE t.user_id=$1 AND t.archived_at IS NULL AND t.excluded_from_reports=false AND t.type='expense'
		AND t.occurred_at >= $2 AND t.occurred_at <= $3
		AND ($4='' OR t.source_wallet_id::text=$4 OR t.destination_wallet_id::text=$4)
		AND COALESCE(parent.id::text,c.id::text,'')=$5`, userID, priorFrom.UTC(), priorTo.UTC(), filter.WalletID, category.ID).Scan(&priorSpent)
	if err != nil {
		return InsiderReport{}, fmt.Errorf("insider prior category: %w", err)
	}
	result.PriorAverageDailyVND = priorSpent / int64(priorDays)
	if result.PriorAverageDailyVND > 0 {
		result.NotComparable = false
		result.ChangePercent = float64(result.AverageDailyVND-result.PriorAverageDailyVND) * 100 / float64(result.PriorAverageDailyVND)
	}
	return result, nil
}

func elapsedPeriodDays(filter Filter, now time.Time) int {
	end := filter.To
	if !now.Before(filter.From) && now.Before(filter.To) {
		end = now
	}
	days := int(end.Sub(filter.From).Hours()/24) + 1
	if days < 1 {
		return 1
	}
	return days
}

func (r *Repository) Daily(ctx context.Context, userID string, filter Filter) ([]DailyTotal, error) {
	rows, err := r.db.QueryContext(ctx, `WITH days AS (SELECT generate_series(($2 AT TIME ZONE 'Asia/Ho_Chi_Minh')::date, ($3 AT TIME ZONE 'Asia/Ho_Chi_Minh')::date, interval '1 day')::date AS day), vals AS (SELECT (occurred_at AT TIME ZONE 'Asia/Ho_Chi_Minh')::date AS day, COALESCE(sum(CASE WHEN type='income' THEN amount_vnd ELSE 0 END),0) income, COALESCE(sum(CASE WHEN type='expense' THEN amount_vnd ELSE 0 END),0) expense FROM transactions WHERE user_id=$1 AND archived_at IS NULL AND excluded_from_reports=false AND type IN ('income','expense') AND occurred_at >= $2 AND occurred_at <= $3 AND ($4='' OR source_wallet_id::text=$4 OR destination_wallet_id::text=$4) GROUP BY 1) SELECT days.day, COALESCE(vals.income,0), COALESCE(vals.expense,0) FROM days LEFT JOIN vals USING(day) ORDER BY days.day`, userID, filter.From.UTC(), filter.To.UTC(), filter.WalletID)
	if err != nil {
		return nil, fmt.Errorf("daily report: %w", err)
	}
	defer rows.Close()
	result := []DailyTotal{}
	var cumulative int64
	for rows.Next() {
		var date time.Time
		var d DailyTotal
		if err := rows.Scan(&date, &d.IncomeVND, &d.ExpenseVND); err != nil {
			return nil, err
		}
		d.Date = date.Format("2006-01-02")
		d.NetIncomeVND = d.IncomeVND - d.ExpenseVND
		cumulative += d.NetIncomeVND
		d.CumulativeNetVND = cumulative
		result = append(result, d)
	}
	return result, rows.Err()
}

func (r *Repository) Dashboard(ctx context.Context, userID string, filter Filter) (Dashboard, error) {
	summary, err := r.Summary(ctx, userID, filter)
	if err != nil {
		return Dashboard{}, err
	}
	d := Dashboard{Summary: summary}
	rows, err := r.db.QueryContext(ctx, `SELECT id::text,name,balance_vnd,include_in_total FROM wallets WHERE user_id=$1 AND archived_at IS NULL ORDER BY created_at,id`, userID)
	if err != nil {
		return Dashboard{}, err
	}
	for rows.Next() {
		var w Wallet
		if err := rows.Scan(&w.ID, &w.Name, &w.BalanceVND, &w.IncludeInTotal); err != nil {
			rows.Close()
			return Dashboard{}, err
		}
		d.Wallets = append(d.Wallets, w)
		if w.IncludeInTotal {
			d.NetWorthVND += w.BalanceVND
		}
	}
	rows.Close()
	d.WalletNetWorthVND = d.NetWorthVND
	d.CombinedNetWorthVND = d.WalletNetWorthVND
	if err := r.attachPortfolioTotals(ctx, userID, &d); err != nil {
		return Dashboard{}, err
	}
	rows, err = r.db.QueryContext(ctx, `SELECT id::text,type,amount_vnd,note,occurred_at FROM transactions WHERE user_id=$1 AND archived_at IS NULL ORDER BY occurred_at DESC,created_at DESC LIMIT 10`, userID)
	if err != nil {
		return Dashboard{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var t RecentTransaction
		if err := rows.Scan(&t.ID, &t.Type, &t.AmountVND, &t.Note, &t.OccurredAt); err != nil {
			return Dashboard{}, err
		}
		d.Recent = append(d.Recent, t)
	}
	return d, rows.Err()
}

func (r *Repository) attachPortfolioTotals(ctx context.Context, userID string, d *Dashboard) error {
	rows, err := r.db.QueryContext(ctx, `
		WITH latest_prices AS (
			SELECT DISTINCT ON (asset_id) asset_id, unit_price_vnd
			FROM asset_price_history
			WHERE user_id = $1
			ORDER BY asset_id, priced_at DESC, created_at DESC, id DESC
		), latest_trades AS (
			SELECT DISTINCT ON (asset_id) asset_id, quantity_after
			FROM asset_trades
			WHERE user_id = $1 AND archived_at IS NULL
			ORDER BY asset_id, occurred_at DESC, created_at DESC, id DESC
		)
		SELECT coalesce(t.quantity_after::text, '0'), p.unit_price_vnd
		FROM asset_positions a
		LEFT JOIN latest_trades t ON t.asset_id = a.id
		LEFT JOIN latest_prices p ON p.asset_id = a.id
		WHERE a.user_id = $1 AND a.archived_at IS NULL AND a.include_in_net_worth = true
	`, userID)
	if err != nil {
		return fmt.Errorf("portfolio dashboard totals: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var quantityText string
		var price sql.NullInt64
		if err := rows.Scan(&quantityText, &price); err != nil {
			return err
		}
		if !price.Valid {
			d.MissingAssetPriceCount++
			continue
		}
		d.InvestmentMarketValueVND += roundDecimalMoney(quantityText, price.Int64)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	d.CombinedNetWorthVND += d.InvestmentMarketValueVND
	return nil
}

func roundDecimalMoney(quantity string, unitPriceVND int64) int64 {
	rat := new(big.Rat)
	if _, ok := rat.SetString(quantity); !ok {
		return 0
	}
	value := new(big.Rat).Mul(rat, big.NewRat(unitPriceVND, 1))
	num := new(big.Int).Set(value.Num())
	den := new(big.Int).Set(value.Denom())
	absNum := new(big.Int).Abs(num)
	q, rem := new(big.Int).QuoRem(absNum, den, new(big.Int))
	rem.Mul(rem, big.NewInt(2))
	if rem.Cmp(den) >= 0 {
		q.Add(q, big.NewInt(1))
	}
	if value.Sign() < 0 {
		q.Neg(q)
	}
	return q.Int64()
}
