package httpapi

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/infrastructure/ai"
	"github.com/mypocket/backend/internal/infrastructure/financialsource"
	repo "github.com/mypocket/backend/internal/infrastructure/repository"
	"github.com/mypocket/backend/internal/usecase"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"strings"
	"testing"
)

type aiEntryCapabilityStorage struct{}

func (aiEntryCapabilityStorage) Key(string, string, string, string) (string, error) {
	return "key", nil
}
func (aiEntryCapabilityStorage) Put(context.Context, string, string, []byte) error { return nil }
func (aiEntryCapabilityStorage) Delete(context.Context, string) error              { return nil }
func (aiEntryCapabilityStorage) SignedGet(context.Context, string) (string, error) { return "", nil }

func TestAIEntryRoutesExposeOneShotProcessInsteadOfSessions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := &Router{Engine: gin.New(), AuthMiddleware: &AuthMiddleware{}}
	router.RegisterAIEntryRoutes(&AIEntryHandler{})
	routes := map[string]bool{}
	for _, route := range router.Engine.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	if !routes["POST /api/v1/ai/entry/process"] || !routes["GET /api/v1/ai/entry/requests/:request_id"] {
		t.Fatalf("missing one-shot processing/recovery routes: %+v", routes)
	}
	if !routes["GET /api/v1/transactions/:id/attachments/:attachmentId/download"] {
		t.Fatalf("missing owner-scoped transaction attachment download route: %+v", routes)
	}
	for route := range routes {
		if strings.Contains(route, "/sessions") || strings.Contains(route, "/messages") {
			t.Fatalf("conversation/session route must not be public: %s", route)
		}
	}
}

func TestAIEntryProcessResponseIncludesSafeAttachmentDiagnostics(t *testing.T) {
	response := processResponse(&entity.AIEntrySession{
		ID: "request-1",
		AttachmentDiagnostics: []entity.AIEntryAttachmentDiagnostic{{
			Position: 2, Filename: "second.pdf", Stage: "ocr", Status: "failed", ErrorCode: "ocr_failed",
		}},
	})
	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		AttachmentDiagnostics []entity.AIEntryAttachmentDiagnostic `json:"attachment_diagnostics"`
	}
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded.AttachmentDiagnostics) != 1 || decoded.AttachmentDiagnostics[0].Position != 2 || decoded.AttachmentDiagnostics[0].ErrorCode != "ocr_failed" {
		t.Fatalf("attachment diagnostics missing from process response: %s", encoded)
	}
}

func TestAIEntryCapabilitiesAdvertiseStructuredFilesWithoutOCR(t *testing.T) {
	handler := &AIEntryHandler{Service: &usecase.AIEntryService{
		Extractor: ai.NewClient(ai.Config{BaseURL: "https://llm.invalid/v1", APIKey: "test", Model: "test"}),
		Storage:   aiEntryCapabilityStorage{},
	}}
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	handler.Capabilities(context)
	if recorder.Code != http.StatusOK {
		t.Fatalf("Capabilities status = %d", recorder.Code)
	}
	var response struct {
		Data map[string]bool `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Data["ocr_configured"] || response.Data["files_configured"] || !response.Data["structured_files_configured"] {
		t.Fatalf("unexpected capability split: %+v", response.Data)
	}
}

func TestAIEntryProcessRejectsSpoofedPDFBeforeProviderWork(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(c *gin.Context) { c.Set(contextUserIDKey, "owner"); c.Next() })
	engine.POST("/process", (&AIEntryHandler{}).Process)
	body := &bytes.Buffer{}
	form := multipart.NewWriter(body)
	_ = form.WriteField("request_id", uuid.NewString())
	_ = form.WriteField("text", "receipt")
	_ = form.WriteField("timezone", "UTC")
	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", `form-data; name="files"; filename="receipt.pdf"`)
	header.Set("Content-Type", "application/pdf")
	part, err := form.CreatePart(header)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write([]byte("this is not a pdf"))
	_ = form.Close()
	req := httptest.NewRequest(http.MethodPost, "/process", body)
	req.Header.Set("Content-Type", form.FormDataContentType())
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("spoofed PDF status = %d; body=%s", rec.Code, rec.Body.String())
	}
}

func TestAIEntryProcessAcceptsPNGWhenClientPartMIMEIsBlankOrWrong(t *testing.T) {
	body := []byte("\x89PNG\r\n\x1a\nvalid-png-payload")
	formBody := &bytes.Buffer{}
	form := multipart.NewWriter(formBody)
	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", `form-data; name="files"; filename="IMG_7868.PNG"`)
	header.Set("Content-Type", "application/octet-stream")
	part, err := form.CreatePart(header)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(body); err != nil {
		t.Fatal(err)
	}
	if err := form.Close(); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/process", formBody)
	request.Header.Set("Content-Type", form.FormDataContentType())
	if err := request.ParseMultipartForm(2 << 20); err != nil {
		t.Fatal(err)
	}
	parsed, err := parseAIEntryFiles(request.MultipartForm.File["files"], "UTC")
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed) != 1 || parsed[0].MIMEType != "image/png" {
		t.Fatalf("unexpected parsed image: %+v", parsed)
	}
}

func TestParseAIEntryFilesRejectsUnsafeFilenames(t *testing.T) {
	for _, filename := range []string{"../statement.xlsx", `folder\statement.xlsx`, "statement\n.xlsx", "statement\x00.xlsx"} {
		t.Run(strings.ReplaceAll(filename, "/", "_"), func(t *testing.T) {
			header := &multipart.FileHeader{Filename: filename, Size: 1}
			if _, err := parseAIEntryFiles([]*multipart.FileHeader{header}, "UTC"); err == nil {
				t.Fatal("parseAIEntryFiles() error = nil, want unsafe filename rejection")
			}
		})
	}
}

func TestParseAIEntryFilesAcceptsStructuredXLSXAndRejectsArbitraryZIP(t *testing.T) {
	valid := aiEntryTestWorkbook(t)
	if _, err := financialsource.ParseWorkbook("statement.xlsx", valid, "Asia/Ho_Chi_Minh"); err != nil {
		t.Fatalf("test workbook is invalid: %v", err)
	}
	if got := http.DetectContentType(valid); got != "application/zip" {
		t.Fatalf("test workbook MIME = %q, want application/zip", got)
	}
	for _, tt := range []struct {
		name    string
		body    []byte
		wantErr bool
	}{
		{name: "structured workbook", body: valid},
		{name: "arbitrary zip", body: aiEntryTestZIP(t, map[string]string{"payload.txt": "not a workbook"}), wantErr: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			formBody := &bytes.Buffer{}
			form := multipart.NewWriter(formBody)
			part, err := form.CreateFormFile("files", "statement.xlsx")
			if err != nil {
				t.Fatal(err)
			}
			if _, err := part.Write(tt.body); err != nil {
				t.Fatal(err)
			}
			if err := form.Close(); err != nil {
				t.Fatal(err)
			}
			request := httptest.NewRequest(http.MethodPost, "/process", formBody)
			request.Header.Set("Content-Type", form.FormDataContentType())
			if err := request.ParseMultipartForm(6 << 20); err != nil {
				t.Fatal(err)
			}
			parsed, err := parseAIEntryFiles(request.MultipartForm.File["files"], "Asia/Ho_Chi_Minh")
			if tt.wantErr {
				if err == nil {
					t.Fatal("parseAIEntryFiles() error = nil, want rejection")
				}
				return
			}
			if err != nil {
				t.Fatalf("parseAIEntryFiles() error = %v", err)
			}
			if len(parsed) != 1 || parsed[0].MIMEType != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
				t.Fatalf("unexpected parsed workbook: %+v", parsed)
			}
		})
	}
}

func aiEntryTestWorkbook(t *testing.T) []byte {
	t.Helper()
	return aiEntryTestZIP(t, map[string]string{
		"[Content_Types].xml":        `<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/><Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/></Types>`,
		"_rels/.rels":                `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>`,
		"xl/workbook.xml":            `<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets><sheet name="Transactions" sheetId="1" r:id="rId1"/></sheets></workbook>`,
		"xl/_rels/workbook.xml.rels": `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/></Relationships>`,
		"xl/worksheets/sheet1.xml":   `<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData><row r="1"><c r="A1" t="inlineStr"><is><t>Ngay giao dich</t></is></c><c r="B1" t="inlineStr"><is><t>So tham chieu</t></is></c><c r="C1" t="inlineStr"><is><t>Dien giai</t></is></c><c r="D1" t="inlineStr"><is><t>Ghi no</t></is></c><c r="E1" t="inlineStr"><is><t>Ghi co</t></is></c></row><row r="2"><c r="A2" t="inlineStr"><is><t>26/10/2026</t></is></c><c r="B2" t="inlineStr"><is><t>ref-1</t></is></c><c r="C2" t="inlineStr"><is><t>coffee</t></is></c><c r="D2" t="inlineStr"><is><t>30,000</t></is></c><c r="E2" t="inlineStr"><is><t></t></is></c></row></sheetData></worksheet>`,
	})
}

func aiEntryTestZIP(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for name, contents := range files {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write([]byte(contents)); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func TestAIEntryHTTPReviewEditApproveRejectWithRealDatabase(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires migrated TEST_DATABASE_URL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	owner := uuid.NewString()
	wallet := entity.Wallet{ID: uuid.NewString(), OwnerID: owner, Name: "review test", Type: "basic", Currency: "VND"}
	if err = db.Create(&entity.User{ID: owner, GoogleSubject: owner, Email: owner + "@test.invalid"}).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Where("id = ?", wallet.ID).Delete(&entity.Wallet{})
		db.Where("id = ?", wallet.OwnerID).Delete(&entity.User{})
	})
	if err = db.Create(&wallet).Error; err != nil {
		t.Fatal(err)
	}
	providerCalls := 0
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		providerCalls++
		if r.URL.Path != "/chat/completions" {
			t.Errorf("wrong path %s", r.URL.Path)
		}
		draft := map[string]any{"type": "expense", "amount": 35000, "wallet_id": wallet.ID, "category_id": nil, "occurred_at": "2026-09-21T10:00:00+07:00", "note": "ăn sáng", "included_in_reports": true, "questions": []string{}, "source_currency": "VND"}
		secondDraft := map[string]any{"type": "expense", "amount": 36000, "wallet_id": wallet.ID, "category_id": nil, "occurred_at": "2026-09-21T12:00:00+07:00", "note": "ăn trưa", "included_in_reports": true, "questions": []string{}, "source_currency": "VND"}
		content, _ := json.Marshal(map[string]any{"reply": "Bạn kiểm tra các đề xuất trước khi duyệt nhé.", "drafts": []any{draft, secondDraft}})
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]any{"content": string(content)}, "finish_reason": "stop"}}})
	}))
	defer provider.Close()
	service := &usecase.AIEntryService{Entries: repo.NewAIEntryPostgresRepository(db), Wallets: repo.NewWalletPostgresRepository(db), Categories: repo.NewCategoryPostgresRepository(db), Extractor: ai.NewClient(ai.Config{BaseURL: provider.URL, APIKey: "test-key", Model: "test-model"})}
	h := &AIEntryHandler{Service: service}
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(c *gin.Context) { c.Set(contextUserIDKey, c.GetHeader("X-Test-Owner")); c.Next() })
	engine.POST("/process", h.Process)
	engine.GET("/requests/:request_id", h.Request)
	engine.PATCH("/proposals/:id", h.EditProposal)
	engine.POST("/proposals/:id/approve", h.ApproveProposal)
	engine.POST("/proposals/:id/reject", h.RejectProposal)
	request := func(method, path string, body any, want int, out any) {
		t.Helper()
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(method, path, bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Test-Owner", owner)
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, req)
		if rec.Code != want {
			t.Fatalf("%s %s: %d %s", method, path, rec.Code, rec.Body.String())
		}
		if out != nil {
			var envelope struct {
				Data json.RawMessage `json:"data"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(envelope.Data, out); err != nil {
				t.Fatal(err)
			}
		}
	}
	requestID := uuid.NewString()
	process := func(want int, out any) {
		t.Helper()
		body := &bytes.Buffer{}
		form := multipart.NewWriter(body)
		_ = form.WriteField("request_id", requestID)
		_ = form.WriteField("text", "ăn sáng 35k, ăn trưa 35k")
		_ = form.WriteField("timezone", "Asia/Ho_Chi_Minh")
		_ = form.Close()
		req := httptest.NewRequest("POST", "/process", body)
		req.Header.Set("Content-Type", form.FormDataContentType())
		req.Header.Set("X-Test-Owner", owner)
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, req)
		if rec.Code != want {
			t.Fatalf("POST /process: %d %s", rec.Code, rec.Body.String())
		}
		var envelope struct {
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(envelope.Data, out); err != nil {
			t.Fatal(err)
		}
	}
	requestID = uuid.NewString()
	var processView aiEntryProcessResponse
	process(200, &processView)
	if len(processView.Proposals) != 2 || processView.ID != requestID {
		t.Fatalf("missing one-shot review list: %+v", processView)
	}
	var messageCount int64
	db.Model(&entity.AIEntryMessage{}).Where("session_id = ?", requestID).Count(&messageCount)
	if messageCount != 0 {
		t.Fatalf("one-shot request stored chat messages: %d", messageCount)
	}
	var count int64
	db.Model(&entity.Transaction{}).Where("owner_id = ?", owner).Count(&count)
	if count != 0 {
		t.Fatal("extraction wrote ledger")
	}
	process(200, &processView)
	if providerCalls != 1 || len(processView.Proposals) != 2 {
		t.Fatal("request replay repeated extraction")
	}
	p := processView.Proposals[0]
	p.Draft.Amount = 45000
	p.Draft.IncludedInReports = false
	request("PATCH", "/proposals/"+p.ID, map[string]any{"version": p.Version, "draft": p.Draft}, 200, &p)
	request("POST", "/proposals/"+p.ID+"/approve", map[string]any{"version": 1}, 409, nil)
	request("POST", "/proposals/"+p.ID+"/approve", map[string]any{"version": p.Version}, 200, &p)
	if p.Status != entity.AIProposalApplied || p.TransactionID == nil {
		t.Fatal("approval not persisted")
	}
	request("POST", "/proposals/"+p.ID+"/approve", map[string]any{"version": 2}, 200, &p)
	request("POST", "/proposals/"+processView.Proposals[1].ID+"/reject", map[string]any{"version": 1}, 200, nil)
	var rows []entity.Transaction
	db.Where("owner_id = ?", owner).Find(&rows)
	if len(rows) != 1 || rows[0].Amount != 45000 || rows[0].IncludedInReports {
		t.Fatalf("wrong approved ledger: %+v", rows)
	}
	request("GET", "/requests/"+requestID, nil, 200, &processView)
	if processView.Proposals[0].Status == "pending" || processView.Proposals[1].Status == "pending" {
		t.Fatal("review state not restored")
	}
	owner = "other"
	request("GET", "/requests/"+requestID, nil, 404, nil)
	owner = wallet.OwnerID // Cleanup always uses the test fixture's owner.
}
