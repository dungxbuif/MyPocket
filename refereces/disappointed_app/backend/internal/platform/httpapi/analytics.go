package httpapi

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"mypocket/internal/analytics"
	"mypocket/internal/finance"
	"mypocket/internal/platform/config"
)

type analyticsEnvelope struct {
	Status        string    `json:"status"`
	Report        any       `json:"report"`
	GeneratedAt   time.Time `json:"generated_at"`
	Timezone      string    `json:"timezone"`
	From          string    `json:"from"`
	To            string    `json:"to"`
	DataVersion   int64     `json:"data_version"`
	CorrelationID string    `json:"correlation_id"`
}
type searchResult struct {
	Kind   string `json:"kind"`
	ID     string `json:"id"`
	Label  string `json:"label"`
	Detail string `json:"detail,omitempty"`
}
type searchEnvelope struct {
	Status        string         `json:"status"`
	Results       []searchResult `json:"results"`
	Stale         bool           `json:"stale"`
	CorrelationID string         `json:"correlation_id"`
}

func dashboard(cfg config.Config, financeRepo FinanceRepository, analyticsRepo AnalyticsRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		if analyticsRepo == nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Analytics unavailable", correlationID(r.Context())))
			return
		}
		filter, err := parseAnalyticsFilter(r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", err.Error(), correlationID(r.Context())))
			return
		}
		d, err := analyticsRepo.Dashboard(r.Context(), userID, filter)
		if err != nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Dashboard unavailable", correlationID(r.Context())))
			return
		}
		writeJSON(w, http.StatusOK, analyticsEnvelope{Status: "ok", Report: d, GeneratedAt: d.Summary.GeneratedAt, Timezone: d.Summary.Timezone, From: d.Summary.From, To: d.Summary.To, DataVersion: d.Summary.DataVersion, CorrelationID: correlationID(r.Context())})
	}
}

func reports(cfg config.Config, repo AnalyticsRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		if repo == nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Analytics unavailable", correlationID(r.Context())))
			return
		}
		filter, err := parseAnalyticsFilter(r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorEnvelope("VALIDATION_FAILED", err.Error(), correlationID(r.Context())))
			return
		}
		kind := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/reports/"), "/")
		if kind == "insider" {
			insider, insiderErr := repo.Insider(r.Context(), userID, filter)
			if insiderErr != nil {
				writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Money Insider unavailable", correlationID(r.Context())))
				return
			}
			writeJSON(w, http.StatusOK, analyticsEnvelope{Status: "ok", Report: insider, GeneratedAt: insider.GeneratedAt, Timezone: insider.Timezone, From: insider.From, To: insider.To, DataVersion: insider.DataVersion, CorrelationID: correlationID(r.Context())})
			return
		}
		var report analytics.Report
		report.Summary, err = repo.Summary(r.Context(), userID, filter)
		if err == nil && kind == "categories" {
			report.Categories, err = repo.Categories(r.Context(), userID, filter)
		}
		if err == nil && kind == "comparison" {
			periodFilters := comparisonPeriods(filter)
			report.Periods = make([]analytics.Summary, 0, len(periodFilters))
			for index, periodFilter := range periodFilters {
				periodSummary := report.Summary
				if index != len(periodFilters)-1 {
					periodSummary, err = repo.Summary(r.Context(), userID, periodFilter)
					if err != nil {
						break
					}
				}
				report.Periods = append(report.Periods, periodSummary)
			}
			if err == nil {
				priorSummary := report.Periods[len(report.Periods)-2]
				report.Prior = &priorSummary
				report.Summary.NotComparable = priorSummary.IncomeVND == 0 && priorSummary.ExpenseVND == 0
				if priorSummary.IncomeVND > 0 {
					report.Summary.IncomeChangePercent = float64(report.Summary.IncomeVND-priorSummary.IncomeVND) * 100 / float64(priorSummary.IncomeVND)
				}
				if priorSummary.ExpenseVND > 0 {
					report.Summary.ExpenseChangePercent = float64(report.Summary.ExpenseVND-priorSummary.ExpenseVND) * 100 / float64(priorSummary.ExpenseVND)
				}
			}
		}
		if err == nil && kind == "cumulative" && r.URL.Query().Get("from") == "" && r.URL.Query().Get("to") == "" {
			loc := time.FixedZone("Asia/Ho_Chi_Minh", 7*60*60)
			if loaded, loadErr := time.LoadLocation("Asia/Ho_Chi_Minh"); loadErr == nil {
				loc = loaded
			}
			now := time.Now().In(loc)
			filter.From = time.Date(now.Year(), now.Month()-2, 1, 0, 0, 0, 0, loc)
			filter.To = time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 999999999, loc)
			report.Summary, err = repo.Summary(r.Context(), userID, filter)
		}
		if err == nil && (kind == "daily" || kind == "cumulative" || kind == "cash-flow") {
			report.Daily, err = repo.Daily(r.Context(), userID, filter)
		}
		if err != nil {
			writeJSON(w, http.StatusServiceUnavailable, ErrorEnvelope("INTERNAL_RETRYABLE", "Report unavailable", correlationID(r.Context())))
			return
		}
		if kind != "categories" && kind != "daily" && kind != "cumulative" && kind != "cash-flow" && kind != "comparison" {
			writeJSON(w, http.StatusNotFound, ErrorEnvelope("NOT_FOUND", "Report not found", correlationID(r.Context())))
			return
		}
		writeJSON(w, http.StatusOK, analyticsEnvelope{Status: "ok", Report: report, GeneratedAt: report.Summary.GeneratedAt, Timezone: report.Summary.Timezone, From: report.Summary.From, To: report.Summary.To, DataVersion: report.Summary.DataVersion, CorrelationID: correlationID(r.Context())})
	}
}

func comparisonPeriods(current analytics.Filter) []analytics.Filter {
	periods := make([]analytics.Filter, 6)
	periods[len(periods)-1] = current
	span := current.To.Sub(current.From) + time.Nanosecond
	for index := len(periods) - 2; index >= 0; index-- {
		next := periods[index+1]
		periods[index] = analytics.Filter{From: next.From.Add(-span), To: next.To.Add(-span), WalletID: current.WalletID}
	}
	return periods
}

func parseAnalyticsFilter(r *http.Request) (analytics.Filter, error) {
	from, to, err := parseDateRange(r.URL.Query().Get("from"), r.URL.Query().Get("to"))
	if err != nil {
		return analytics.Filter{}, err
	}
	filter := analytics.NormalizeFilter(from, to)
	if filter.From.After(filter.To) {
		return analytics.Filter{}, fmt.Errorf("report from date must not be after to date")
	}
	filter.WalletID = strings.TrimSpace(r.URL.Query().Get("wallet_id"))
	return filter, nil
}
func parseDateRange(from, to string) (time.Time, time.Time, error) {
	loc := time.FixedZone("Asia/Ho_Chi_Minh", 7*60*60)
	if loaded, loadErr := time.LoadLocation("Asia/Ho_Chi_Minh"); loadErr == nil {
		loc = loaded
	}
	var start, end time.Time
	var err error
	if from != "" {
		start, err = time.ParseInLocation("2006-01-02", from, loc)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}
	if to != "" {
		end, err = time.ParseInLocation("2006-01-02", to, loc)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		end = end.Add(24*time.Hour - time.Nanosecond)
	}
	return start, end, nil
}

func search(cfg config.Config, financeRepo FinanceRepository, planningRepo PlanningRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := authenticatedUserID(w, r, cfg)
		if !ok {
			return
		}
		q := strings.TrimSpace(r.URL.Query().Get("q"))
		if q == "" {
			writeJSON(w, http.StatusOK, searchEnvelope{Status: "ok", Results: []searchResult{}, CorrelationID: correlationID(r.Context())})
			return
		}
		results := []searchResult{}
		if financeRepo != nil {
			wallets, _ := financeRepo.ListWallets(r.Context(), userID)
			for _, item := range wallets {
				if strings.Contains(strings.ToLower(item.Name), strings.ToLower(q)) {
					results = append(results, searchResult{Kind: "wallet", ID: item.ID, Label: item.Name, Detail: "Ví"})
				}
			}
			cats, _ := financeRepo.ListCategories(r.Context(), userID)
			for _, item := range cats {
				if strings.Contains(strings.ToLower(item.Name), strings.ToLower(q)) {
					results = append(results, searchResult{Kind: "category", ID: item.ID, Label: item.Name, Detail: "Nhóm"})
				}
			}
			txs, _ := financeRepo.ListTransactions(r.Context(), userID, finance.TransactionFilters{Query: q})
			for _, item := range txs {
				results = append(results, searchResult{Kind: "transaction", ID: item.ID, Label: item.Note, Detail: formatVND(item.AmountVND)})
			}
		}
		if planningRepo != nil {
			events, _ := planningRepo.ListEvents(r.Context(), userID)
			for _, item := range events {
				if strings.Contains(strings.ToLower(item.Name), strings.ToLower(q)) {
					results = append(results, searchResult{Kind: "event", ID: item.ID, Label: item.Name, Detail: "Sự kiện"})
				}
			}
			debts, _ := planningRepo.ListObligations(r.Context(), userID)
			for _, item := range debts {
				if strings.Contains(strings.ToLower(item.Counterparty), strings.ToLower(q)) {
					results = append(results, searchResult{Kind: "obligation", ID: item.ID, Label: item.Counterparty, Detail: "Khoản nợ"})
				}
			}
		}
		if len(results) > 50 {
			results = results[:50]
		}
		writeJSON(w, http.StatusOK, searchEnvelope{Status: "ok", Results: results, CorrelationID: correlationID(r.Context())})
	}
}

func formatVND(amount int64) string { return fmt.Sprintf("%d đ", amount) }
