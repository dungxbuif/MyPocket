package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"sort"
	"strconv"
	"strings"
)

var errSchema = errors.New("AI response does not match the draft schema")

type SchemaIssue struct {
	Code     string `json:"code"`
	Path     string `json:"path"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
}

type SchemaMismatchError struct {
	Issues []SchemaIssue `json:"issues"`
}

func (e *SchemaMismatchError) Error() string { return errSchema.Error() }
func (e *SchemaMismatchError) Unwrap() error { return errSchema }

func parseOutput(data []byte) (Output, error) {
	out, err := parseOutputStrict(data)
	if err == nil || !errors.Is(err, errSchema) {
		return out, err
	}
	issues := diagnoseOutput(data)
	if len(issues) == 0 {
		issues = []SchemaIssue{{Code: "schema_constraint", Path: "$", Expected: "valid transaction draft schema", Actual: "constraint_violation"}}
	}
	return Output{}, &SchemaMismatchError{Issues: issues}
}

func parseOutputStrict(data []byte) (Output, error) {
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

// diagnoseOutput reports only response structure and JSON types, never field values.
func diagnoseOutput(data []byte) []SchemaIssue {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return []SchemaIssue{{Code: "invalid_json", Path: "$", Expected: "JSON object", Actual: "malformed_json"}}
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return []SchemaIssue{{Code: "trailing_content", Path: "$", Expected: "one JSON value", Actual: "additional_content"}}
	}
	root, ok := value.(map[string]any)
	if !ok {
		issues := []SchemaIssue{{Code: "root_type", Path: "$", Expected: "object", Actual: jsonType(value)}}
		if items, isArray := value.([]any); isArray {
			issues[0].Actual = "array(length=" + strconv.Itoa(len(items)) + ")"
			for i, item := range items {
				if i == 3 {
					issues = append(issues, SchemaIssue{Code: "array_items_truncated", Path: "$", Expected: "at most 3 item shapes", Actual: "more_items"})
					break
				}
				path := "$[" + strconv.Itoa(i) + "]"
				if object, isObject := item.(map[string]any); isObject {
					issues = append(issues, SchemaIssue{Code: "array_item_shape", Path: path, Expected: "reply/drafts envelope or transaction draft", Actual: "object fields=" + strings.Join(diagnosticKeys(object), ",")})
				} else {
					issues = append(issues, SchemaIssue{Code: "array_item_shape", Path: path, Expected: "object", Actual: jsonType(item)})
				}
			}
		}
		return issues
	}
	issues := make([]SchemaIssue, 0, 8)
	checkObjectFields(&issues, "$", root, []string{"reply", "drafts"})
	if reply, exists := root["reply"]; exists {
		if _, ok := reply.(string); !ok {
			addTypeIssue(&issues, "$.reply", "string", reply)
		}
	}
	draftValue, exists := root["drafts"]
	if !exists {
		return issues
	}
	drafts, ok := draftValue.([]any)
	if !ok {
		addTypeIssue(&issues, "$.drafts", "array", draftValue)
		return issues
	}
	if len(drafts) > 30 {
		issues = append(issues, SchemaIssue{Code: "array_limit", Path: "$.drafts", Expected: "at most 30 items", Actual: "too_many_items"})
	}
	for i, item := range drafts {
		path := "$.drafts[" + strconv.Itoa(i) + "]"
		draft, ok := item.(map[string]any)
		if !ok {
			addTypeIssue(&issues, path, "object", item)
			continue
		}
		fields := []string{"type", "amount", "wallet_id", "category_id", "occurred_at", "note", "included_in_reports", "questions"}
		checkObjectFields(&issues, path, draft, fields)
		for _, field := range fields {
			v, present := draft[field]
			if !present {
				continue
			}
			fieldPath := path + "." + field
			switch field {
			case "type", "wallet_id", "occurred_at", "note":
				if _, ok := v.(string); !ok {
					addTypeIssue(&issues, fieldPath, "string", v)
				}
			case "amount":
				amount, ok := v.(json.Number)
				if !ok {
					addTypeIssue(&issues, fieldPath, "integer", v)
				} else if _, err := amount.Int64(); err != nil {
					issues = append(issues, SchemaIssue{Code: "field_type", Path: fieldPath, Expected: "integer", Actual: "non_integer_number"})
				}
			case "category_id":
				if v != nil {
					if _, ok := v.(string); !ok {
						addTypeIssue(&issues, fieldPath, "string or null", v)
					}
				}
			case "included_in_reports":
				if _, ok := v.(bool); !ok {
					addTypeIssue(&issues, fieldPath, "boolean", v)
				}
			case "questions":
				questions, ok := v.([]any)
				if !ok {
					addTypeIssue(&issues, fieldPath, "array of strings", v)
				} else {
					for q, question := range questions {
						if _, ok := question.(string); !ok {
							addTypeIssue(&issues, fieldPath+"["+strconv.Itoa(q)+"]", "string", question)
						}
					}
				}
			}
			if len(issues) >= 12 {
				return issues[:12]
			}
		}
	}
	if len(issues) > 12 {
		issues = issues[:12]
	}
	return issues
}

func checkObjectFields(issues *[]SchemaIssue, path string, actual map[string]any, expected []string) {
	keys := diagnosticKeys(actual)
	keysTruncated := len(actual) > len(keys)
	missing := make([]string, 0)
	for _, key := range expected {
		if _, ok := actual[key]; !ok {
			missing = append(missing, key)
		}
	}
	unexpected := make([]string, 0)
	for _, key := range keys {
		found := false
		for _, want := range expected {
			if key == want {
				found = true
				break
			}
		}
		if !found {
			unexpected = append(unexpected, key)
		}
	}
	if len(missing) > 0 || len(unexpected) > 0 {
		actualFields := []string{"present=" + strings.Join(keys, ",")}
		if keysTruncated {
			actualFields = append(actualFields, "present_keys_truncated=true")
		}
		if len(missing) > 0 {
			actualFields = append(actualFields, "missing="+strings.Join(missing, ","))
		}
		if len(unexpected) > 0 {
			actualFields = append(actualFields, "unexpected="+strings.Join(unexpected, ","))
		}
		*issues = append(*issues, SchemaIssue{Code: "object_fields", Path: path, Expected: strings.Join(expected, ","), Actual: strings.Join(actualFields, ";")})
	}
}

func diagnosticKeys(object map[string]any) []string {
	keys := make([]string, 0, len(object))
	for key := range object {
		keys = append(keys, safeDiagnosticKey(key))
	}
	sort.Strings(keys)
	if len(keys) > 16 {
		keys = keys[:16]
	}
	return keys
}

func safeDiagnosticKey(key string) string {
	switch key {
	case "reply", "drafts", "type", "amount", "wallet_id", "category_id", "occurred_at", "note", "included_in_reports", "questions",
		"id", "name", "currency", "wallets", "categories", "now", "timezone", "untrusted_source_text",
		"choices", "message", "content", "role", "finish_reason", "tool_calls", "function_call", "usage", "prompt_tokens", "completion_tokens":
		return key
	default:
		return "<unrecognized_field>"
	}
}

func addTypeIssue(issues *[]SchemaIssue, path, expected string, actual any) {
	*issues = append(*issues, SchemaIssue{Code: "field_type", Path: path, Expected: expected, Actual: jsonType(actual)})
}

func jsonType(value any) string {
	switch value.(type) {
	case nil:
		return "null"
	case map[string]any:
		return "object"
	case []any:
		return "array"
	case string:
		return "string"
	case bool:
		return "boolean"
	case json.Number:
		return "number"
	default:
		return "unknown"
	}
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
