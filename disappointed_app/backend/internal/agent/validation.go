package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

var ErrValidation = errors.New("agent validation failed")

func ValidateCreate(kind Kind, text, key string) error {
	text, key = strings.TrimSpace(text), strings.TrimSpace(key)
	if kind != KindTransactionDraft && kind != KindAnalysis && kind != KindIntake && kind != KindAdvisor {
		return fmt.Errorf("%w: unsupported kind", ErrValidation)
	}
	if len(text) == 0 || len(text) > 8000 {
		return fmt.Errorf("%w: message must be 1..8000 characters", ErrValidation)
	}
	if len(key) == 0 || len(key) > 200 {
		return fmt.Errorf("%w: idempotency key is required", ErrValidation)
	}
	return nil
}

func ParseModelResult(raw string, kind Kind) (ModelResult, error) {
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.DisallowUnknownFields()
	var result ModelResult
	if err := decoder.Decode(&result); err != nil {
		return ModelResult{}, fmt.Errorf("%w: malformed provider result", ErrValidation)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return ModelResult{}, fmt.Errorf("%w: trailing provider result", ErrValidation)
	}
	if kind == KindAnalysis || kind == KindAdvisor {
		if strings.TrimSpace(result.Answer) == "" || result.Transaction != nil || len(result.Drafts) > 0 {
			return ModelResult{}, fmt.Errorf("%w: analysis must contain answer only", ErrValidation)
		}
		return result, nil
	}
	drafts := result.Drafts
	if result.Transaction != nil {
		drafts = append(drafts, *result.Transaction)
	}
	if len(drafts) == 0 {
		return ModelResult{}, fmt.Errorf("%w: transaction proposal is required", ErrValidation)
	}
	if kind == KindTransactionDraft && len(drafts) != 1 {
		return ModelResult{}, fmt.Errorf("%w: transaction_draft requires one proposal", ErrValidation)
	}
	if len(drafts) > 20 {
		return ModelResult{}, fmt.Errorf("%w: too many draft proposals", ErrValidation)
	}
	for i := range drafts {
		tx := &drafts[i]
		if tx.Type != "income" && tx.Type != "expense" && tx.Type != "transfer" {
			return ModelResult{}, fmt.Errorf("%w: invalid transaction type", ErrValidation)
		}
		if kind == KindIntake && tx.Type == "transfer" {
			return ModelResult{}, fmt.Errorf("%w: intake does not support transfers", ErrValidation)
		}
		if tx.AmountVND <= 0 || tx.AmountVND > 9_000_000_000_000_000 {
			return ModelResult{}, fmt.Errorf("%w: invalid amount", ErrValidation)
		}
		if strings.TrimSpace(tx.SourceWalletID) == "" || len(tx.Note) > 500 {
			return ModelResult{}, fmt.Errorf("%w: invalid transaction fields", ErrValidation)
		}
		if _, err := time.Parse(time.RFC3339, tx.OccurredAt); err != nil {
			return ModelResult{}, fmt.Errorf("%w: occurred_at must be RFC3339", ErrValidation)
		}
		if tx.Type == "transfer" {
			if tx.DestinationWalletID == nil || strings.TrimSpace(*tx.DestinationWalletID) == "" || *tx.DestinationWalletID == tx.SourceWalletID || tx.CategoryID != nil {
				return ModelResult{}, fmt.Errorf("%w: invalid transfer shape", ErrValidation)
			}
		} else if tx.DestinationWalletID != nil || tx.CategoryID == nil || strings.TrimSpace(*tx.CategoryID) == "" {
			return ModelResult{}, fmt.Errorf("%w: invalid income/expense shape", ErrValidation)
		}
	}
	result.Drafts = drafts
	return result, nil
}

func TransactionSchema() json.RawMessage {
	return json.RawMessage(`{"type":"object","additionalProperties":false,"required":["transaction"],"properties":{"transaction":{"type":"object","additionalProperties":false,"required":["type","amount_vnd","source_wallet_id","occurred_at","note"],"properties":{"type":{"enum":["income","expense","transfer"]},"amount_vnd":{"type":"integer","minimum":1},"source_wallet_id":{"type":"string"},"destination_wallet_id":{"type":"string"},"category_id":{"type":"string"},"occurred_at":{"type":"string","format":"date-time"},"note":{"type":"string","maxLength":500}}}}}`)
}

func IntakeSchema() json.RawMessage {
	return json.RawMessage(`{"type":"object","additionalProperties":false,"required":["drafts"],"properties":{"drafts":{"type":"array","minItems":1,"maxItems":20,"items":{"type":"object","additionalProperties":false,"required":["type","amount_vnd","source_wallet_id","category_id","occurred_at","note"],"properties":{"type":{"enum":["income","expense"]},"amount_vnd":{"type":"integer","minimum":1},"source_wallet_id":{"type":"string"},"category_id":{"type":"string"},"occurred_at":{"type":"string","format":"date-time"},"note":{"type":"string","maxLength":500}}}}}}`)
}

func AnalysisSchema() json.RawMessage {
	return json.RawMessage(`{"type":"object","additionalProperties":false,"required":["answer"],"properties":{"answer":{"type":"string"}}}`)
}
