package httpapi

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mypocket/backend/internal/usecase"
)

// ExternalOperation describes one canonical authenticated REST operation. It
// deliberately has no dependency on an MCP transport or finance repository.
type ExternalOperation struct {
	Name            string
	Description     string
	Method          string
	PathTemplate    string
	InputSchema     map[string]any
	BodyEncoding    string
	Idempotency     string
	Async           bool
	SensitiveResult bool
	QueryParameters []string
}

type ExternalOperationCatalog struct {
	operations []ExternalOperation
	byName     map[string]ExternalOperation
}

func (c *ExternalOperationCatalog) List() []ExternalOperation {
	if c == nil {
		return nil
	}
	result := make([]ExternalOperation, len(c.operations))
	for i, op := range c.operations {
		result[i] = cloneExternalOperation(op)
	}
	return result
}

func (c *ExternalOperationCatalog) Lookup(name string) (ExternalOperation, bool) {
	if c == nil {
		return ExternalOperation{}, false
	}
	op, ok := c.byName[name]
	return cloneExternalOperation(op), ok
}

var externalOperationName = regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)+$`)
var externalPathParameter = regexp.MustCompile(`^\{([A-Za-z][A-Za-z0-9_]*)\}$`)
var externalLiteralSegment = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// NewExternalOperationCatalog validates the whole manifest before it becomes
// dispatch authority. Returned descriptors are copies, including their schema.
func NewExternalOperationCatalog(operations []ExternalOperation) (*ExternalOperationCatalog, error) {
	c := &ExternalOperationCatalog{byName: map[string]ExternalOperation{}}
	routes := map[string]bool{}
	for _, op := range operations {
		if err := validateExternalOperation(op); err != nil {
			return nil, fmt.Errorf("invalid operation %q: %w", op.Name, err)
		}
		if _, exists := c.byName[op.Name]; exists {
			return nil, fmt.Errorf("duplicate operation name %q", op.Name)
		}
		key := op.Method + " " + op.PathTemplate
		if routes[key] {
			return nil, fmt.Errorf("duplicate operation route %q", key)
		}
		routes[key] = true
		op = cloneExternalOperation(op)
		c.byName[op.Name] = op
		c.operations = append(c.operations, op)
	}
	sort.Slice(c.operations, func(i, j int) bool { return c.operations[i].Name < c.operations[j].Name })
	return c, nil
}

func validateExternalOperation(op ExternalOperation) error {
	if !externalOperationName.MatchString(op.Name) {
		return fmt.Errorf("name must be lowercase and domain-qualified")
	}
	switch op.Method {
	case "GET", "POST", "PUT", "PATCH", "DELETE":
	default:
		return fmt.Errorf("unsupported method")
	}
	if !strings.HasPrefix(op.PathTemplate, "/api/v1/") {
		return fmt.Errorf("path must be local and versioned")
	}
	relative := strings.TrimPrefix(op.PathTemplate, "/api/v1/")
	first := strings.Split(relative, "/")[0]
	switch first {
	case "agent", "internal", "mcp", "login", "health", "docs", "changelog":
		return fmt.Errorf("route is outside authenticated user parity")
	}
	if first == "auth" && relative != "auth/profile" && relative != "auth/devices" && relative != "auth/devices/{id}" && relative != "auth/logout" {
		return fmt.Errorf("credential exchange is outside authenticated user parity")
	}
	if op.InputSchema == nil || op.InputSchema["type"] != "object" || op.InputSchema["additionalProperties"] != false {
		return fmt.Errorf("input requires a closed object schema")
	}
	properties, ok := op.InputSchema["properties"].(map[string]any)
	if !ok {
		return fmt.Errorf("input properties missing")
	}
	if err := validateExternalSchema(op.InputSchema, 0); err != nil {
		return err
	}
	pathNames := map[string]bool{}
	for _, segment := range strings.Split(relative, "/") {
		if match := externalPathParameter.FindStringSubmatch(segment); match != nil {
			name := match[1]
			if pathNames[name] || externalOwnerSelector(name) {
				return fmt.Errorf("invalid repeated/owner path selector")
			}
			pathNames[name] = true
			property, ok := properties[name].(map[string]any)
			if !ok || property["type"] != "string" || !externalSchemaRequired(op.InputSchema, name) {
				return fmt.Errorf("path parameter %q requires a required string schema", name)
			}
		} else if !externalLiteralSegment.MatchString(segment) {
			return fmt.Errorf("invalid path segment")
		}
	}
	seenQuery := map[string]bool{}
	for _, name := range op.QueryParameters {
		if seenQuery[name] || pathNames[name] || properties[name] == nil || externalOwnerSelector(name) {
			return fmt.Errorf("invalid query parameter %q", name)
		}
		seenQuery[name] = true
	}
	switch op.BodyEncoding {
	case "json", "query", "multipart", "empty":
	default:
		return fmt.Errorf("explicit body encoding required")
	}
	if op.Method == "GET" && (op.BodyEncoding == "json" || op.BodyEncoding == "multipart") {
		return fmt.Errorf("GET cannot encode a body")
	}
	if op.BodyEncoding == "empty" || op.BodyEncoding == "query" {
		for name := range properties {
			if !pathNames[name] && !seenQuery[name] {
				return fmt.Errorf("argument %q has no declared encoding", name)
			}
		}
	}
	switch op.Idempotency {
	case "none", "optional", "required":
	default:
		return fmt.Errorf("explicit idempotency mode required")
	}
	return nil
}

func validateExternalSchema(schema map[string]any, depth int) error {
	if depth > 32 {
		return fmt.Errorf("schema too deeply nested")
	}
	if _, ok := schema["$ref"]; ok {
		return fmt.Errorf("schema references must be resolved")
	}
	if choices, ok := schema["anyOf"].([]any); ok {
		if len(choices) == 0 {
			return fmt.Errorf("empty anyOf schema")
		}
	} else {
		switch schema["type"] {
		case "null", "string", "boolean", "integer", "number":
		case "array":
			if _, ok := schema["items"].(map[string]any); !ok {
				return fmt.Errorf("array schema requires item schema")
			}
		case "object":
			properties, ok := schema["properties"].(map[string]any)
			if !ok || schema["additionalProperties"] != false {
				return fmt.Errorf("nested object requires closed properties")
			}
			if required, exists := schema["required"]; exists {
				var fields []string
				switch required := required.(type) {
				case []string:
					fields = required
				case []any:
					for _, field := range required {
						name, ok := field.(string)
						if !ok {
							return fmt.Errorf("required names must be strings")
						}
						fields = append(fields, name)
					}
				default:
					return fmt.Errorf("required names must be an array")
				}
				seen := map[string]bool{}
				for _, name := range fields {
					if seen[name] || properties[name] == nil {
						return fmt.Errorf("invalid required property %q", name)
					}
					seen[name] = true
				}
			}
		default:
			return fmt.Errorf("unsupported schema type")
		}
	}
	if properties, ok := schema["properties"].(map[string]any); ok {
		for name, value := range properties {
			if externalOwnerSelector(name) || name == "_meta" || name == "idempotency_key" {
				return fmt.Errorf("reserved schema property %q", name)
			}
			property, ok := value.(map[string]any)
			if !ok {
				return fmt.Errorf("invalid schema property %q", name)
			}
			if err := validateExternalSchema(property, depth+1); err != nil {
				return err
			}
		}
	}
	if items, ok := schema["items"].(map[string]any); ok {
		if err := validateExternalSchema(items, depth+1); err != nil {
			return err
		}
	}
	if choices, ok := schema["anyOf"].([]any); ok {
		for _, choice := range choices {
			s, ok := choice.(map[string]any)
			if !ok {
				return fmt.Errorf("invalid anyOf schema")
			}
			if err := validateExternalSchema(s, depth+1); err != nil {
				return err
			}
		}
	}
	return nil
}

func externalSchemaRequired(schema map[string]any, name string) bool {
	switch required := schema["required"].(type) {
	case []string:
		for _, field := range required {
			if field == name {
				return true
			}
		}
	case []any:
		for _, field := range required {
			if field == name {
				return true
			}
		}
	}
	return false
}

func cloneExternalOperation(op ExternalOperation) ExternalOperation {
	op.InputSchema = cloneExternalSchema(op.InputSchema)
	op.QueryParameters = append([]string(nil), op.QueryParameters...)
	return op
}

func cloneExternalSchema(schema map[string]any) map[string]any {
	if schema == nil {
		return nil
	}
	copy := make(map[string]any, len(schema))
	for key, value := range schema {
		switch value := value.(type) {
		case map[string]any:
			copy[key] = cloneExternalSchema(value)
		case []string:
			copy[key] = append([]string(nil), value...)
		case []any:
			items := make([]any, len(value))
			for i, item := range value {
				if object, ok := item.(map[string]any); ok {
					items[i] = cloneExternalSchema(object)
				} else {
					items[i] = item
				}
			}
			copy[key] = items
		default:
			copy[key] = value
		}
	}
	return copy
}

// Encode translates only the declared operation into an HTTP request. The
// dispatcher attaches the original bearer and context; no credential or owner
// supplied as a business argument can override that authenticated principal.
// Domain validity and required idempotency remain canonical-handler decisions.
func (op ExternalOperation) Encode(arguments map[string]any, meta map[string]any) (*http.Request, error) {
	if err := validateExternalOperation(op); err != nil {
		return nil, err
	}
	if arguments == nil {
		arguments = map[string]any{}
	}
	if err := rejectExternalOwnerSelectors(arguments, 0); err != nil {
		return nil, err
	}
	if err := rejectExternalOwnerSelectors(meta, 0); err != nil {
		return nil, err
	}
	data, err := json.Marshal(arguments)
	if err != nil {
		return nil, fmt.Errorf("arguments must be JSON")
	}
	var normalized map[string]any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&normalized); err != nil {
		return nil, fmt.Errorf("arguments must be an object")
	}
	if err := validateExternalValue(normalized, op.InputSchema, 0); err != nil {
		return nil, err
	}
	key := ""
	for name, value := range meta {
		if name != "idempotency_key" {
			return nil, fmt.Errorf("unsupported metadata field %q", name)
		}
		var ok bool
		key, ok = value.(string)
		if !ok || len(key) < 1 || len(key) > 128 {
			return nil, fmt.Errorf("invalid idempotency key")
		}
		for _, r := range key {
			if r < 32 || r > 126 {
				return nil, fmt.Errorf("invalid idempotency key")
			}
		}
	}
	path := op.PathTemplate
	for _, segment := range strings.Split(op.PathTemplate, "/") {
		match := externalPathParameter.FindStringSubmatch(segment)
		if match == nil {
			continue
		}
		name := match[1]
		value, ok := normalized[name].(string)
		if !ok || value == "" || len(value) > 256 || strings.TrimSpace(value) != value || value == "." || value == ".." || strings.ContainsAny(value, "/\\%?#\x00\r\n") {
			return nil, fmt.Errorf("invalid path argument %q", name)
		}
		path = strings.ReplaceAll(path, segment, url.PathEscape(value))
		delete(normalized, name)
	}
	query := url.Values{}
	for _, name := range op.QueryParameters {
		if value, exists := normalized[name]; exists {
			text, err := externalQueryValue(value)
			if err != nil {
				return nil, fmt.Errorf("invalid query argument %q", name)
			}
			query.Set(name, text)
			delete(normalized, name)
		}
	}
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var body *bytes.Reader
	contentType := ""
	switch op.BodyEncoding {
	case "json":
		data, err = json.Marshal(normalized)
		if err != nil {
			return nil, fmt.Errorf("cannot encode JSON")
		}
		body = bytes.NewReader(data)
		contentType = "application/json"
	case "multipart":
		if op.Name == "feedback.create" {
			data, contentType, err = encodeExternalFeedback(normalized)
		} else {
			data, contentType, err = encodeExternalMultipart(normalized)
		}
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(data)
	case "query", "empty":
		if len(normalized) != 0 {
			return nil, fmt.Errorf("unexpected request body")
		}
	}
	var request *http.Request
	if body == nil {
		request, err = http.NewRequest(op.Method, path, nil)
	} else {
		request, err = http.NewRequest(op.Method, path, body)
	}
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	if key != "" {
		request.Header.Set("Idempotency-Key", key)
	}
	return request, nil
}

func externalOwnerSelector(name string) bool {
	canonical := strings.ToLower(strings.NewReplacer("_", "", "-", "").Replace(name))
	return canonical == "ownerid" || canonical == "userid"
}

func rejectExternalOwnerSelectors(value any, depth int) error {
	if depth > 32 {
		return fmt.Errorf("arguments too deeply nested")
	}
	switch value := value.(type) {
	case map[string]any:
		for name, item := range value {
			if externalOwnerSelector(name) {
				return fmt.Errorf("owner selectors are not accepted")
			}
			if err := rejectExternalOwnerSelectors(item, depth+1); err != nil {
				return err
			}
		}
	case []any:
		for _, item := range value {
			if err := rejectExternalOwnerSelectors(item, depth+1); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateExternalValue(value any, schema map[string]any, depth int) error {
	if depth > 32 {
		return fmt.Errorf("arguments too deeply nested")
	}
	if choices, ok := schema["anyOf"].([]any); ok {
		for _, choice := range choices {
			if s, ok := choice.(map[string]any); ok && validateExternalValue(value, s, depth+1) == nil {
				return nil
			}
		}
		return fmt.Errorf("argument has an invalid nullable type")
	}
	valid := false
	switch schema["type"] {
	case "null":
		valid = value == nil
	case "string":
		_, valid = value.(string)
	case "boolean":
		_, valid = value.(bool)
	case "integer":
		if number, ok := value.(json.Number); ok {
			_, err := number.Int64()
			valid = err == nil
		}
	case "number":
		if number, ok := value.(json.Number); ok {
			n, err := number.Float64()
			valid = err == nil && !math.IsInf(n, 0) && !math.IsNaN(n)
		}
	case "array":
		items, ok := value.([]any)
		if ok {
			valid = true
			itemSchema, _ := schema["items"].(map[string]any)
			for _, item := range items {
				if err := validateExternalValue(item, itemSchema, depth+1); err != nil {
					return err
				}
			}
		}
	case "object":
		object, ok := value.(map[string]any)
		if ok {
			valid = true
			properties, _ := schema["properties"].(map[string]any)
			for name, item := range object {
				property, ok := properties[name].(map[string]any)
				if !ok {
					return fmt.Errorf("unknown argument %q", name)
				}
				if err := validateExternalValue(item, property, depth+1); err != nil {
					return fmt.Errorf("argument %q: %w", name, err)
				}
			}
			for name := range properties {
				if externalSchemaRequired(schema, name) {
					if _, exists := object[name]; !exists {
						return fmt.Errorf("required argument %q is missing", name)
					}
				}
			}
		}
	}
	if !valid {
		return fmt.Errorf("argument type must be %v", schema["type"])
	}
	return nil
}

func externalQueryValue(value any) (string, error) {
	switch value := value.(type) {
	case string:
		return value, nil
	case bool:
		return strconv.FormatBool(value), nil
	case json.Number:
		return value.String(), nil
	}
	return "", fmt.Errorf("query values must be scalar")
}

type externalUploadFile struct {
	Filename string `json:"filename"`
	MIMEType string `json:"mime_type"`
	Base64   string `json:"base64"`
}
type externalAIEntryInput struct {
	RequestID string               `json:"request_id"`
	Text      string               `json:"text"`
	Timezone  string               `json:"timezone"`
	Files     []externalUploadFile `json:"files"`
}

type externalFeedbackInput struct {
	feedbackInput
	Screenshot externalUploadFile `json:"screenshot"`
}

// Feedback's optional screenshot uses its own canonical file/field/body bounds,
// rather than the larger AI-entry upload contract.
func encodeExternalFeedback(arguments map[string]any) ([]byte, string, error) {
	value, screenshot := arguments["screenshot"]
	if !screenshot {
		data, err := json.Marshal(arguments)
		return data, "application/json", err
	}
	file := value.(map[string]any)
	filename, _ := file["filename"].(string)
	mimeType, _ := file["mime_type"].(string)
	encoded, _ := file["base64"].(string)
	if filename == "" || len(filename) > 255 || strings.ContainsAny(filename, "/\\\x00\r\n") || mimeType != "image/png" || len(encoded) > base64.StdEncoding.EncodedLen(1<<20) {
		return nil, "", fmt.Errorf("invalid feedback screenshot")
	}
	data, err := base64.StdEncoding.Strict().DecodeString(encoded)
	if err != nil || len(data) == 0 || len(data) > 1<<20 || http.DetectContentType(data) != "image/png" {
		return nil, "", fmt.Errorf("invalid feedback screenshot bytes")
	}
	var output bytes.Buffer
	writer := multipart.NewWriter(&output)
	for _, name := range []string{"type", "title", "description"} {
		if value, exists := arguments[name]; exists {
			text := value.(string)
			if len(text)+output.Len() > 2<<20 {
				return nil, "", fmt.Errorf("feedback multipart request too large")
			}
			if err := writer.WriteField(name, text); err != nil {
				return nil, "", err
			}
		}
	}
	header := make(textproto.MIMEHeader)
	escaped := strings.NewReplacer("\\", "\\\\", "\"", "\\\"").Replace(filename)
	header.Set("Content-Disposition", `form-data; name="screenshot"; filename="`+escaped+`"`)
	header.Set("Content-Type", "image/png")
	part, err := writer.CreatePart(header)
	if err != nil {
		return nil, "", err
	}
	if _, err := part.Write(data); err != nil {
		return nil, "", err
	}
	if err := writer.Close(); err != nil {
		return nil, "", err
	}
	if output.Len() > 2<<20 {
		return nil, "", fmt.Errorf("feedback multipart request too large")
	}
	return output.Bytes(), writer.FormDataContentType(), nil
}

func encodeExternalMultipart(arguments map[string]any) ([]byte, string, error) {
	var output bytes.Buffer
	writer := multipart.NewWriter(&output)
	for _, name := range []string{"request_id", "text", "timezone"} {
		if value, exists := arguments[name]; exists {
			if err := writer.WriteField(name, value.(string)); err != nil {
				return nil, "", err
			}
		}
	}
	files, _ := arguments["files"].([]any)
	if len(files) > 20 {
		return nil, "", fmt.Errorf("at most 20 files are accepted")
	}
	for _, value := range files {
		file := value.(map[string]any)
		filename, _ := file["filename"].(string)
		mimeType, _ := file["mime_type"].(string)
		encoded, _ := file["base64"].(string)
		if filename == "" || strings.ContainsAny(filename, "/\\\x00\r\n") || len(filename) > 255 || len(encoded) > base64.StdEncoding.EncodedLen(5*1024*1024) {
			return nil, "", fmt.Errorf("invalid upload file")
		}
		data, err := base64.StdEncoding.Strict().DecodeString(encoded)
		if err != nil || len(data) == 0 || len(data) > 5*1024*1024 {
			return nil, "", fmt.Errorf("invalid upload bytes")
		}
		detected := http.DetectContentType(data)
		if (detected != "image/jpeg" && detected != "image/png" && detected != "application/pdf") || mimeType != detected {
			return nil, "", fmt.Errorf("unsupported or mismatched upload MIME")
		}
		header := make(textproto.MIMEHeader)
		escaped := strings.NewReplacer("\\", "\\\\", "\"", "\\\"").Replace(filename)
		header.Set("Content-Disposition", `form-data; name="files"; filename="`+escaped+`"`)
		header.Set("Content-Type", mimeType)
		part, err := writer.CreatePart(header)
		if err != nil {
			return nil, "", err
		}
		if _, err := part.Write(data); err != nil {
			return nil, "", err
		}
		if output.Len() > 101*1024*1024 {
			return nil, "", fmt.Errorf("multipart request too large")
		}
	}
	if err := writer.Close(); err != nil {
		return nil, "", err
	}
	if output.Len() > 101*1024*1024 {
		return nil, "", fmt.Errorf("multipart request too large")
	}
	return output.Bytes(), writer.FormDataContentType(), nil
}

// externalInputSchema follows the handler DTO JSON tags so internal-only
// fields never become tool inputs. No business defaults or validations are
// inferred; those remain in the canonical handler.
func externalInputSchema(t reflect.Type) map[string]any {
	if t.Kind() == reflect.Pointer {
		return map[string]any{"anyOf": []any{externalInputSchema(t.Elem()), map[string]any{"type": "null"}}}
	}
	if t == reflect.TypeFor[time.Time]() {
		return map[string]any{"type": "string", "format": "date-time"}
	}
	switch t.Kind() {
	case reflect.Struct:
		properties := map[string]any{}
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			tag := strings.Split(field.Tag.Get("json"), ",")[0]
			if tag == "-" {
				continue
			}
			if field.Anonymous && tag == "" {
				embedded := externalInputSchema(field.Type)
				for name, value := range embedded["properties"].(map[string]any) {
					properties[name] = value
				}
				continue
			}
			if field.PkgPath != "" || tag == "" {
				continue
			}
			properties[tag] = externalInputSchema(field.Type)
		}
		return map[string]any{"type": "object", "properties": properties, "additionalProperties": false}
	case reflect.Slice:
		return map[string]any{"anyOf": []any{map[string]any{"type": "array", "items": externalInputSchema(t.Elem())}, map[string]any{"type": "null"}}}
	case reflect.Array:
		return map[string]any{"type": "array", "items": externalInputSchema(t.Elem())}
	case reflect.Bool:
		return map[string]any{"type": "boolean"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]any{"type": "integer"}
	case reflect.Float32, reflect.Float64:
		return map[string]any{"type": "number"}
	default:
		return map[string]any{"type": "string"}
	}
}

// DefaultExternalOperationCatalog is a static, complete authenticated-user
// manifest. Invalid descriptors are programming errors and fail startup.
func DefaultExternalOperationCatalog() *ExternalOperationCatalog {
	var operations []ExternalOperation
	add := func(name, method, path string, input any, query ...string) {
		schema := map[string]any{"type": "object", "properties": map[string]any{}, "additionalProperties": false}
		encoding := "empty"
		if input != nil {
			schema = externalInputSchema(reflect.TypeOf(input))
			encoding = "json"
		} else if len(query) > 0 {
			encoding = "query"
		}
		properties := schema["properties"].(map[string]any)
		var required []string
		for _, segment := range strings.Split(path, "/") {
			if match := externalPathParameter.FindStringSubmatch(segment); match != nil {
				properties[match[1]] = map[string]any{"type": "string"}
				required = append(required, match[1])
			}
		}
		for _, field := range query {
			kind := "string"
			switch field {
			case "before_seq", "after_seq", "limit", "offset":
				kind = "integer"
			}
			properties[field] = map[string]any{"type": kind}
		}
		if len(required) > 0 {
			schema["required"] = required
		}
		idempotency := "none"
		operations = append(operations, ExternalOperation{Name: name, Description: method + " " + path, Method: method, PathTemplate: path, InputSchema: schema, BodyEncoding: encoding, Idempotency: idempotency, SensitiveResult: true, QueryParameters: query})
	}
	add("account.profile.get", "GET", "/api/v1/auth/profile", nil)
	add("account.profile.update_timezone", "PATCH", "/api/v1/auth/profile", timezoneInput{})
	add("account.home", "GET", "/api/v1/home", nil)
	add("account.api_key.list", "GET", "/api/v1/api-keys", nil)
	add("account.api_key.create", "POST", "/api/v1/api-keys", apiKeyCreateInput{})
	add("account.api_key.revoke", "DELETE", "/api/v1/api-keys/{id}", nil)
	add("account.device.enroll", "POST", "/api/v1/auth/devices", usecase.DeviceEnrollmentInput{})
	add("account.device.revoke", "DELETE", "/api/v1/auth/devices/{id}", nil)
	add("account.logout", "POST", "/api/v1/auth/logout", refreshRequest{})
	add("categories.list", "GET", "/api/v1/categories", nil)
	add("categories.create", "POST", "/api/v1/categories", categoryInput{})
	add("categories.update", "PATCH", "/api/v1/categories/{id}", categoryInput{})
	add("categories.set_status", "PATCH", "/api/v1/categories/{id}/status", categoryStatusInput{})
	add("categories.set_wallets", "PATCH", "/api/v1/categories/{id}/wallets", categoryWalletInput{})
	add("categories.delete", "DELETE", "/api/v1/categories/{id}", nil)
	add("wallets.list", "GET", "/api/v1/wallets", nil)
	add("wallets.create", "POST", "/api/v1/wallets", walletInput{})
	add("wallets.update", "PATCH", "/api/v1/wallets/{id}", walletInput{})
	add("wallets.delete", "DELETE", "/api/v1/wallets/{id}", nil)
	add("transactions.list", "GET", "/api/v1/transactions", nil)
	add("transactions.create", "POST", "/api/v1/transactions", transactionCreateInput{})
	add("transactions.update", "PATCH", "/api/v1/transactions/{id}", transactionInput{})
	add("transactions.delete", "DELETE", "/api/v1/transactions/{id}", transactionDeleteInput{})
	add("transactions.adjustment", "POST", "/api/v1/transactions/adjustment", adjustmentInput{})
	add("transactions.bulk_delete", "POST", "/api/v1/transactions/bulk-delete", bulkDeleteInput{})
	add("transactions.transfer.create", "POST", "/api/v1/transactions/transfer", transferInput{})
	add("transactions.transfer.update", "PATCH", "/api/v1/transactions/transfer/{transfer_id}", transferUpdateInput{})
	add("transactions.transfer.delete", "DELETE", "/api/v1/transactions/transfer/{transfer_id}", nil)
	add("transactions.travel.update", "PATCH", "/api/v1/transactions/{id}/travel", travelTransactionInput{})
	add("transactions.attachment.download", "GET", "/api/v1/transactions/{id}/attachments/{attachmentId}/download", nil)
	add("budgets.list", "GET", "/api/v1/budgets", nil)
	add("budgets.create", "POST", "/api/v1/budgets", budgetInput{})
	add("budgets.update", "PATCH", "/api/v1/budgets/{id}", budgetInput{})
	add("budgets.delete", "DELETE", "/api/v1/budgets/{id}", nil)
	add("budgets.report", "GET", "/api/v1/budgets/{id}/report", nil)
	add("budgets.transactions", "GET", "/api/v1/budgets/{id}/transactions", nil, "limit", "offset")
	add("budget_series.list", "GET", "/api/v1/budget-series", nil)
	add("budget_series.create", "POST", "/api/v1/budget-series", budgetSeriesInput{})
	add("budget_series.update", "PATCH", "/api/v1/budget-series/{id}", budgetSeriesInput{})
	add("budget_series.stop", "POST", "/api/v1/budget-series/{id}/stop", budgetSeriesInput{})
	add("jars.list", "GET", "/api/v1/jars", nil, "month")
	add("jars.create", "POST", "/api/v1/jars", jarInput{})
	add("jars.month.update", "PUT", "/api/v1/jars/{jar_id}/months/{month}", jarInput{})
	add("jars.month.remove", "DELETE", "/api/v1/jars/{jar_id}/months/{month}", nil)
	add("jars.report", "GET", "/api/v1/jars/{jar_id}/report", nil, "from", "to")
	add("months.get", "GET", "/api/v1/months/{month}", nil)
	add("months.note.save", "PUT", "/api/v1/months/{month}/note", monthNoteInput{})
	add("months.note.delete", "DELETE", "/api/v1/months/{month}/note", nil)
	add("recurring.list", "GET", "/api/v1/recurring", nil)
	add("recurring.create", "POST", "/api/v1/recurring", recurringInput{})
	add("recurring.update", "PATCH", "/api/v1/recurring/{id}", recurringInput{})
	add("recurring.delete", "DELETE", "/api/v1/recurring/{id}", nil)
	add("recurring.run_due", "POST", "/api/v1/recurring/run-due", nil)
	add("credit.entry.create", "POST", "/api/v1/credit/wallets/{wallet_id}/entries", creditEntryInput{})
	add("credit.payment.create", "POST", "/api/v1/credit/wallets/{wallet_id}/payments", creditPaymentInput{})
	add("credit.statement", "GET", "/api/v1/credit/wallets/{wallet_id}/statement", nil, "from", "to")
	add("travel.list", "GET", "/api/v1/travel/events", nil)
	add("travel.create", "POST", "/api/v1/travel/events", travelEventInput{})
	add("travel.update", "PATCH", "/api/v1/travel/events/{id}", travelEventInput{})
	add("travel.delete", "DELETE", "/api/v1/travel/events/{id}", nil)
	add("travel.activate", "POST", "/api/v1/travel/events/{id}/activate", nil)
	add("travel.deactivate", "POST", "/api/v1/travel/events/{id}/deactivate", nil)
	add("portfolio.summary", "GET", "/api/v1/portfolio/summary", nil)
	add("portfolio.asset.list", "GET", "/api/v1/portfolio/assets", nil)
	add("portfolio.asset.create", "POST", "/api/v1/portfolio/assets", portfolioAssetInput{})
	add("portfolio.asset.update", "PATCH", "/api/v1/portfolio/assets/{id}", portfolioAssetUpdateInput{})
	add("portfolio.asset.delete", "DELETE", "/api/v1/portfolio/assets/{id}", nil)
	add("portfolio.asset.update_price", "PATCH", "/api/v1/portfolio/assets/{id}/price", portfolioPriceInput{})
	add("portfolio.trade.list", "GET", "/api/v1/portfolio/assets/{id}/trades", nil)
	add("portfolio.trade.create", "POST", "/api/v1/portfolio/assets/{id}/trades", portfolioTradeInput{})
	add("portfolio.trade.preview_correction", "POST", "/api/v1/portfolio/assets/{id}/trade-corrections/preview", portfolioTradeCorrectionInput{})
	add("portfolio.trade.correct", "POST", "/api/v1/portfolio/assets/{id}/trade-corrections", portfolioTradeCorrectionInput{})
	add("reports.savings", "GET", "/api/v1/reports/savings", nil, "month", "wallet_id")
	add("reports.savings.transactions", "GET", "/api/v1/reports/savings/transactions", nil, "month", "wallet_id", "section", "kind", "limit", "offset")
	add("reports.insider", "GET", "/api/v1/reports/insider", nil, "month", "wallet_id", "category_id")
	add("notifications.list", "GET", "/api/v1/notifications", nil)
	add("notifications.read", "POST", "/api/v1/notifications/{id}/read", nil)
	add("notifications.device.register", "POST", "/api/v1/notifications/devices", usecase.PushTokenInput{})
	add("notifications.device.revoke", "DELETE", "/api/v1/notifications/devices/{id}", nil)
	add("feedback.list", "GET", "/api/v1/feedback", nil)
	add("feedback.create", "POST", "/api/v1/feedback", externalFeedbackInput{})
	add("feedback.get", "GET", "/api/v1/feedback/{id}", nil)
	add("feedback.screenshot", "GET", "/api/v1/feedback/{id}/screenshot", nil)
	add("feedback.set_resolution", "PATCH", "/api/v1/feedback/{id}/resolution", feedbackResolutionInput{})
	add("ai_entry.capabilities", "GET", "/api/v1/ai/entry/capabilities", nil)
	add("ai_entry.process", "POST", "/api/v1/ai/entry/process", externalAIEntryInput{})
	add("ai_entry.request.get", "GET", "/api/v1/ai/entry/requests/{request_id}", nil)
	add("ai_entry.proposal.edit", "PATCH", "/api/v1/ai/entry/proposals/{id}", aiProposalEdit{})
	add("ai_entry.proposal.approve", "POST", "/api/v1/ai/entry/proposals/{id}/approve", aiProposalDecision{})
	add("ai_entry.proposal.reject", "POST", "/api/v1/ai/entry/proposals/{id}/reject", aiProposalDecision{})
	add("advisor.capabilities", "GET", "/api/v1/ai/advisor/capabilities", nil)
	add("advisor.overview", "GET", "/api/v1/ai/advisor/overview", nil, "month")
	add("advisor.history", "GET", "/api/v1/ai/advisor/conversation/messages", nil, "conversation_id", "before_seq")
	add("advisor.reset", "POST", "/api/v1/ai/advisor/conversation/reset", nil)
	add("advisor.messages", "POST", "/api/v1/ai/advisor/messages", advisorMessageInput{})
	add("advisor.messages_async", "POST", "/api/v1/ai/advisor/messages/async", advisorMessageInput{})
	add("advisor.run.get", "GET", "/api/v1/ai/advisor/runs/{id}", nil)
	add("advisor.run.events", "GET", "/api/v1/ai/advisor/runs/{id}/events", nil, "after_seq")
	add("advisor.run.cancel", "POST", "/api/v1/ai/advisor/runs/{id}/cancel", nil)
	add("advisor.action.confirm", "POST", "/api/v1/ai/advisor/actions/{action_id}/confirm", advisorConfirmActionInput{})
	add("sync.capabilities", "GET", "/api/v1/sync/capabilities", nil)
	for i := range operations {
		op := &operations[i]
		switch op.Name {
		case "feedback.create":
			op.BodyEncoding = "multipart"
			properties := op.InputSchema["properties"].(map[string]any)
			screenshot := properties["screenshot"].(map[string]any)
			screenshot["required"] = []any{"filename", "mime_type", "base64"}
			screenshot["description"] = "Optional PNG screenshot, at most 1 MiB decoded; standard base64."
			fields := screenshot["properties"].(map[string]any)
			fields["filename"] = map[string]any{"type": "string", "minLength": 1, "maxLength": 255}
			fields["mime_type"] = map[string]any{"type": "string", "enum": []any{"image/png"}}
			fields["base64"] = map[string]any{"type": "string", "minLength": 1, "maxLength": base64.StdEncoding.EncodedLen(1 << 20)}
		case "wallets.create", "categories.create", "budgets.create", "jars.create", "transactions.create", "transactions.transfer.create":
			op.Idempotency = "optional"
		case "transactions.transfer.update", "transactions.transfer.delete", "budget_series.create", "budget_series.update", "budget_series.stop", "portfolio.trade.correct", "advisor.reset":
			op.Idempotency = "required"
		case "advisor.messages_async":
			op.Async = true
		case "ai_entry.process":
			op.BodyEncoding = "multipart"
			op.Async = true
		}
	}
	catalog, err := NewExternalOperationCatalog(operations)
	if err != nil {
		panic(err)
	}
	return catalog
}
