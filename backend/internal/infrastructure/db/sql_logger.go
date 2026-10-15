package db

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// financialSQLLogger keeps statement class/timing/rows useful without sending
// bound financial data or driver error details to diagnostic output.
type financialSQLLogger struct{ logger.Interface }

func (l financialSQLLogger) LogMode(level logger.LogLevel) logger.Interface {
	return financialSQLLogger{Interface: l.Interface.LogMode(level)}
}

// GORM invokes ParamsFilter before constructing the SQL string passed to Trace.
func (l financialSQLLogger) ParamsFilter(_ context.Context, sql string, _ ...any) (string, []any) {
	return sql, nil
}

func (l financialSQLLogger) Info(ctx context.Context, _ string, data ...any) {
	l.Interface.Info(ctx, "database info event (%d arguments omitted)", len(data))
}

func (l financialSQLLogger) Warn(ctx context.Context, _ string, data ...any) {
	l.Interface.Warn(ctx, "database warning event (%d arguments omitted)", len(data))
}

func (l financialSQLLogger) Error(ctx context.Context, _ string, data ...any) {
	l.Interface.Error(ctx, "database error event (%d arguments omitted)", len(data))
}

func (l financialSQLLogger) Trace(ctx context.Context, begin time.Time, query func() (string, int64), err error) {
	l.Interface.Trace(ctx, begin, func() (string, int64) {
		// Scan's internal recorder can expand SQL before this logger sees it,
		// bypassing ParamsFilter. Never forward any full SQL text, even when it
		// appears parameterized; retain only the allowlisted statement class.
		sql, rows := query()
		fields := strings.Fields(sql)
		kind := "SQL"
		if len(fields) > 0 {
			switch candidate := strings.ToUpper(fields[0]); candidate {
			case "SELECT", "INSERT", "UPDATE", "DELETE", "BEGIN", "COMMIT", "ROLLBACK", "SAVEPOINT", "RELEASE", "WITH", "EXPLAIN":
				kind = candidate
			}
		}
		return kind + " (statement redacted)", rows
	}, safeSQLDiagnosticError(err))
}

func safeSQLDiagnosticError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return gorm.ErrRecordNotFound
	}
	var state interface{ SQLState() string }
	if errors.As(err, &state) {
		code := state.SQLState()
		valid := len(code) == 5
		for _, character := range code {
			if !((character >= '0' && character <= '9') || (character >= 'A' && character <= 'Z')) {
				valid = false
			}
		}
		if valid {
			return fmt.Errorf("database query failed (SQLSTATE %s)", code)
		}
	}
	return errors.New("database query failed")
}
