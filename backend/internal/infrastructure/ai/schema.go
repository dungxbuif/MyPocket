package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

var errSchema = errors.New("AI response does not match the draft schema")

func parseOutput(data []byte) (Output, error) {
	// encoding/json otherwise accepts duplicate keys with last-value wins.
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := uniqueJSON(decoder); err != nil {
		return Output{}, errSchema
	}
	if _, err := decoder.Token(); err != io.EOF {
		return Output{}, errSchema
	}
	root, err := object(data, []string{"reply", "drafts"})
	if err != nil {
		return Output{}, errSchema
	}
	var out Output
	if json.Unmarshal(root["reply"], &out.Reply) != nil || len(out.Reply) > 8192 {
		return Output{}, errSchema
	}
	var drafts []json.RawMessage
	if json.Unmarshal(root["drafts"], &drafts) != nil || len(drafts) > 30 {
		return Output{}, errSchema
	}
	out.Drafts = make([]Draft, 0, len(drafts))
	for _, raw := range drafts {
		fields, err := object(raw, []string{"type", "amount", "wallet_id", "category_id", "occurred_at", "note", "included_in_reports", "questions"})
		if err != nil {
			return Output{}, errSchema
		}
		var draft Draft
		if json.Unmarshal(raw, &draft) != nil || draft.Amount < 0 || draft.Amount > 9007199254740991 || len(draft.Type) > 64 || len(draft.WalletID) > 128 || (draft.CategoryID != nil && len(*draft.CategoryID) > 128) || len(draft.OccurredAt) > 128 || len(draft.Note) > 4096 || len(draft.Questions) > 10 {
			return Output{}, errSchema
		}
		// Reject null question elements, which otherwise decode as empty strings.
		var questions []json.RawMessage
		if json.Unmarshal(fields["questions"], &questions) != nil {
			return Output{}, errSchema
		}
		for i, q := range questions {
			if bytes.Equal(bytes.TrimSpace(q), []byte("null")) || len(draft.Questions[i]) > 1024 {
				return Output{}, errSchema
			}
		}
		out.Drafts = append(out.Drafts, draft)
	}
	return out, nil
}

func object(data []byte, keys []string) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if json.Unmarshal(data, &fields) != nil || len(fields) != len(keys) {
		return nil, errSchema
	}
	for _, key := range keys {
		value, ok := fields[key]
		if !ok || (key != "category_id" && bytes.Equal(bytes.TrimSpace(value), []byte("null"))) {
			return nil, errSchema
		}
	}
	return fields, nil
}

func uniqueJSON(d *json.Decoder) error {
	tok, err := d.Token()
	if err != nil {
		return err
	}
	delim, ok := tok.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := map[string]bool{}
		for d.More() {
			key, err := d.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok || seen[name] {
				return errSchema
			}
			seen[name] = true
			if err := uniqueJSON(d); err != nil {
				return err
			}
		}
	case '[':
		for d.More() {
			if err := uniqueJSON(d); err != nil {
				return err
			}
		}
	default:
		return errSchema
	}
	_, err = d.Token()
	return err
}
