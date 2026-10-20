package lifecycle

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"time"
)

type ExportRow struct {
	Dataset, ID, Type, Name, SourceWalletID, DestinationWalletID, CategoryID, Note string
	AmountVND                                                                      int64
	OccurredAt                                                                     time.Time
}

type Snapshot struct{ Rows []ExportRow }

type ExportFormatter interface {
	Write(context.Context, Snapshot, io.Writer) error
}

type CSVExportFormatter struct{}

func (CSVExportFormatter) Write(_ context.Context, snapshot Snapshot, out io.Writer) error {
	w := csv.NewWriter(out)
	if err := w.Write([]string{"dataset", "id", "type", "name", "amount_vnd", "source_wallet_id", "destination_wallet_id", "category_id", "occurred_at", "note"}); err != nil {
		return err
	}
	for _, row := range snapshot.Rows {
		occurred := ""
		if !row.OccurredAt.IsZero() {
			occurred = row.OccurredAt.UTC().Format(time.RFC3339Nano)
		}
		if err := w.Write([]string{row.Dataset, row.ID, row.Type, row.Name, strconv.FormatInt(row.AmountVND, 10), row.SourceWalletID, row.DestinationWalletID, row.CategoryID, occurred, row.Note}); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func (r *Repository) ExportSnapshot(ctx context.Context, userID string, request ExportRequest) (Snapshot, error) {
	if request.SnapshotAt.IsZero() {
		request.SnapshotAt = time.Now().UTC()
	}
	allowed := map[string]bool{"wallets": true, "categories": true, "transactions": true}
	selected := map[string]bool{}
	for _, dataset := range request.Datasets {
		if !allowed[dataset] {
			return Snapshot{}, ErrInvalid
		}
		selected[dataset] = true
	}
	if len(selected) == 0 {
		selected = allowed
	}
	result := Snapshot{}
	if selected["wallets"] {
		rows, err := r.db.QueryContext(ctx, `SELECT id::text,type,name,balance_vnd FROM wallets WHERE user_id=$1 AND created_at<=$2 ORDER BY created_at,id`, userID, request.SnapshotAt)
		if err != nil {
			return Snapshot{}, err
		}
		for rows.Next() {
			var x ExportRow
			x.Dataset = "wallets"
			if err := rows.Scan(&x.ID, &x.Type, &x.Name, &x.AmountVND); err != nil {
				rows.Close()
				return Snapshot{}, err
			}
			result.Rows = append(result.Rows, x)
		}
		rows.Close()
	}
	if selected["categories"] {
		rows, err := r.db.QueryContext(ctx, `SELECT id::text,kind,name FROM categories WHERE user_id=$1 AND created_at<=$2 ORDER BY created_at,id`, userID, request.SnapshotAt)
		if err != nil {
			return Snapshot{}, err
		}
		for rows.Next() {
			var x ExportRow
			x.Dataset = "categories"
			if err := rows.Scan(&x.ID, &x.Type, &x.Name); err != nil {
				rows.Close()
				return Snapshot{}, err
			}
			result.Rows = append(result.Rows, x)
		}
		rows.Close()
	}
	if selected["transactions"] {
		query := `SELECT id::text,type,amount_vnd,source_wallet_id::text,coalesce(destination_wallet_id::text,''),coalesce(category_id::text,''),occurred_at,note FROM transactions WHERE user_id=$1 AND created_at<=$2`
		args := []any{userID, request.SnapshotAt}
		n := 3
		if request.From != nil {
			query += fmt.Sprintf(" AND occurred_at >= $%d", n)
			args = append(args, request.From.UTC())
			n++
		}
		if request.To != nil {
			query += fmt.Sprintf(" AND occurred_at <= $%d", n)
			args = append(args, request.To.UTC())
		}
		query += " ORDER BY occurred_at,id"
		rows, err := r.db.QueryContext(ctx, query, args...)
		if err != nil {
			return Snapshot{}, err
		}
		for rows.Next() {
			var x ExportRow
			x.Dataset = "transactions"
			if err := rows.Scan(&x.ID, &x.Type, &x.AmountVND, &x.SourceWalletID, &x.DestinationWalletID, &x.CategoryID, &x.OccurredAt, &x.Note); err != nil {
				rows.Close()
				return Snapshot{}, err
			}
			result.Rows = append(result.Rows, x)
		}
		rows.Close()
	}
	return result, nil
}
