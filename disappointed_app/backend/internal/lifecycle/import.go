package lifecycle

import (
	"encoding/csv"
	"io"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const MaxImportRows = 10_000
const MaxImportBytes = 15 << 20

var importHeaders = []string{"occurred_at", "type", "amount_vnd", "source_wallet_id", "destination_wallet_id", "category_id", "note", "excluded_from_reports"}

type ImportParser struct{}

func (ImportParser) Parse(jobID string, version int64, input io.Reader, ownedWallets, ownedCategories map[string]bool) ImportPreview {
	preview := ImportPreview{ID: jobID, Version: version, ValidRows: []ImportRow{}, Errors: []ImportRowError{}}
	r := csv.NewReader(io.LimitReader(input, MaxImportBytes+1))
	r.FieldsPerRecord = -1
	header, err := r.Read()
	if err != nil || !equalHeaders(header, importHeaders) {
		preview.Errors = append(preview.Errors, ImportRowError{Row: 1, Code: "INVALID_HEADERS", Message: "CSV headers are invalid"})
		return preview
	}
	for rowNumber := 2; rowNumber <= MaxImportRows+2; rowNumber++ {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			preview.Errors = append(preview.Errors, ImportRowError{Row: rowNumber, Code: "INVALID_CSV", Message: "CSV row is invalid"})
			break
		}
		if rowNumber > MaxImportRows+1 {
			preview.Errors = append(preview.Errors, ImportRowError{Row: rowNumber, Code: "ROW_LIMIT", Message: "CSV exceeds 10000 rows"})
			break
		}
		row, rowErr := parseImportRow(rowNumber, record, ownedWallets, ownedCategories)
		if rowErr != nil {
			preview.Errors = append(preview.Errors, *rowErr)
		} else {
			preview.ValidRows = append(preview.ValidRows, row)
		}
	}
	preview.Confirmable = len(preview.ValidRows) > 0 && len(preview.Errors) == 0
	return preview
}

func parseImportRow(number int, record []string, wallets, categories map[string]bool) (ImportRow, *ImportRowError) {
	bad := func(code, message string) *ImportRowError {
		return &ImportRowError{Row: number, Code: code, Message: message}
	}
	if len(record) != len(importHeaders) {
		return ImportRow{}, bad("FIELD_COUNT", "CSV row has wrong field count")
	}
	for _, value := range record {
		if !utf8.ValidString(value) {
			return ImportRow{}, bad("INVALID_UTF8", "CSV must be UTF-8")
		}
	}
	at, err := time.Parse(time.RFC3339, strings.TrimSpace(record[0]))
	if err != nil {
		return ImportRow{}, bad("INVALID_DATE", "occurred_at must be RFC3339")
	}
	amount, err := strconv.ParseInt(strings.TrimSpace(record[2]), 10, 64)
	if err != nil || amount <= 0 {
		return ImportRow{}, bad("INVALID_AMOUNT", "amount_vnd must be a positive integer")
	}
	typeValue := strings.TrimSpace(record[1])
	if typeValue != "income" && typeValue != "expense" && typeValue != "transfer" {
		return ImportRow{}, bad("INVALID_TYPE", "type is invalid")
	}
	source, destination, category := strings.TrimSpace(record[3]), strings.TrimSpace(record[4]), strings.TrimSpace(record[5])
	if !wallets[source] {
		return ImportRow{}, bad("UNKNOWN_WALLET", "source wallet is not owned")
	}
	if destination != "" && !wallets[destination] {
		return ImportRow{}, bad("UNKNOWN_WALLET", "destination wallet is not owned")
	}
	if category != "" && !categories[category] {
		return ImportRow{}, bad("UNKNOWN_CATEGORY", "category is not available")
	}
	excluded, err := strconv.ParseBool(strings.TrimSpace(record[7]))
	if err != nil {
		return ImportRow{}, bad("INVALID_BOOLEAN", "excluded_from_reports must be true or false")
	}
	return ImportRow{Row: number, OccurredAt: at, Type: typeValue, AmountVND: amount, SourceWalletID: source, DestinationWalletID: destination, CategoryID: category, Note: strings.TrimSpace(record[6]), ExcludedFromReports: excluded}, nil
}

func equalHeaders(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if strings.TrimSpace(got[i]) != want[i] {
			return false
		}
	}
	return true
}
