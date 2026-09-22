package ai

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/mypocket/backend/internal/config"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/infrastructure/storage"
)

// TestLiveQwenImageAndTransactionEval is opt-in: it sends a synthetic receipt
// and synthetic transaction descriptions to the configured OCR, S3 and LLM.
// It creates and deletes one private S3 object, and never writes ledger data.
func TestLiveQwenImageAndTransactionEval(t *testing.T) {
	if os.Getenv("AI_EVAL_LIVE") != "1" {
		t.Skip("set AI_EVAL_LIVE=1 to run the live OCR/S3/Qwen evaluation")
	}
	if err := config.LoadLocalEnv(filepath.Join("..", "..", "..", ".env.local")); err != nil {
		t.Fatal("load local provider configuration:", err)
	}
	cfg := config.Load()
	if strings.TrimSpace(os.Getenv("AI_EVAL_IMAGE")) == "" {
		t.Fatal("AI_EVAL_IMAGE must point to a synthetic receipt image")
	}
	data, err := os.ReadFile(os.Getenv("AI_EVAL_IMAGE"))
	if err != nil || len(data) == 0 {
		t.Fatal("could not read synthetic image fixture")
	}
	mime := http.DetectContentType(data)
	if mime != "image/png" && mime != "image/jpeg" {
		t.Fatalf("unsupported evaluation image MIME %q", mime)
	}
	client := NewClient(Config{BaseURL: cfg.AIBaseURL, APIKey: cfg.AIAPIKey, Model: cfg.AIModel, OCRURL: cfg.OCRAPIURL, OCRKey: cfg.OCRAPIKey})
	if !client.Configured() || !client.OCRConfigured() {
		t.Fatal("AI/OCR provider configuration is incomplete")
	}

	store, err := storage.NewS3(storage.S3Config{Endpoint: cfg.S3Endpoint, Region: cfg.S3Region, Bucket: cfg.S3Bucket, Prefix: cfg.S3Prefix, Environment: cfg.AppEnv, AccessKeyID: cfg.S3AccessKeyID, SecretAccessKey: cfg.S3SecretAccessKey, ForcePathStyle: cfg.S3ForcePathStyle})
	if err != nil {
		t.Fatal("configure private S3:", err)
	}
	batch := fmt.Sprintf("eval-%d", time.Now().UTC().UnixNano())
	key, err := store.Key("uat-owner", batch, "receipt", "mock-receipt.png")
	if err != nil {
		t.Fatal("build private S3 key:", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	if err := store.Put(ctx, key, mime, data); err != nil {
		t.Fatal("upload synthetic receipt to private S3:", err)
	}
	defer func() {
		if err := store.Delete(context.Background(), key); err != nil {
			t.Errorf("delete synthetic S3 object: %v", err)
		}
	}()

	signedURL, err := store.SignedGet(ctx, key)
	if err != nil {
		t.Fatal("create private signed GET URL:", err)
	}
	getStart := time.Now()
	response, err := (&http.Client{Timeout: 30 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}).Get(signedURL)
	if err != nil {
		t.Fatal("read back private S3 object:", err)
	}
	readBack, readErr := io.ReadAll(io.LimitReader(response.Body, int64(len(data))+1))
	_ = response.Body.Close()
	if readErr != nil || response.StatusCode != http.StatusOK || string(readBack) != string(data) {
		var providerError struct {
			Code             string `xml:"Code"`
			Message          string `xml:"Message"`
			CanonicalRequest string `xml:"CanonicalRequest"`
			StringToSign     string `xml:"StringToSign"`
		}
		_ = xml.Unmarshal(readBack, &providerError)
		serverCanonicalHash := sha256.Sum256([]byte(providerError.CanonicalRequest))
		serverHash := hex.EncodeToString(serverCanonicalHash[:])
		var localHash string
		if parsed, parseErr := url.Parse(signedURL); parseErr == nil {
			query := parsed.Query()
			query.Del("X-Amz-Signature")
			canonical := strings.Join([]string{http.MethodGet, parsed.EscapedPath(), query.Encode(), "host:" + parsed.Host + "\n", "host", "UNSIGNED-PAYLOAD"}, "\n")
			localCanonicalHash := sha256.Sum256([]byte(canonical))
			localHash = hex.EncodeToString(localCanonicalHash[:])
		}
		redirectHost := ""
		if location, parseErr := url.Parse(response.Header.Get("Location")); parseErr == nil {
			redirectHost = location.Host
		}
		if response.StatusCode == http.StatusForbidden {
			alternateConfig := storage.S3Config{Endpoint: cfg.S3Endpoint, Region: "us-east-1", Bucket: cfg.S3Bucket, Prefix: cfg.S3Prefix, Environment: cfg.AppEnv, AccessKeyID: cfg.S3AccessKeyID, SecretAccessKey: cfg.S3SecretAccessKey, ForcePathStyle: cfg.S3ForcePathStyle}
			if alternate, alternateErr := storage.NewS3(alternateConfig); alternateErr == nil {
				if alternateURL, signErr := alternate.SignedGet(ctx, key); signErr == nil {
					alternateResponse, getErr := (&http.Client{Timeout: 30 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}).Get(alternateURL)
					if getErr == nil {
						alternateBody, _ := io.ReadAll(io.LimitReader(alternateResponse.Body, int64(len(data))+1))
						_ = alternateResponse.Body.Close()
						if alternateResponse.StatusCode == http.StatusOK && string(alternateBody) == string(data) {
							t.Log("diagnostic: signed GET succeeds with standard S3 region us-east-1; configured signing region is incompatible with this endpoint")
						}
					}
				}
			}
		}
		t.Fatalf("private S3 readback failed: status=%d code=%q message=%q redirect_host=%q server_canonical_hash=%s local_canonical_hash=%s string_to_sign=%q", response.StatusCode, providerError.Code, providerError.Message, redirectHost, serverHash, localHash, providerError.StringToSign)
	}
	t.Logf("S3 put+signed-get: pass, bytes=%d, latency=%s", len(data), time.Since(getStart).Round(time.Millisecond))

	ocrStart := time.Now()
	ocrText, err := client.ocr(ctx, Image{Name: "mock-receipt.png", MIMEType: mime, SourceURL: signedURL})
	ocrLatency := time.Since(ocrStart)
	if err != nil {
		t.Fatal("live OCR of private S3 image:", err)
	}
	if !strings.Contains(strings.ReplaceAll(strings.ReplaceAll(ocrText, ".", ""), ",", ""), "45000") {
		t.Fatalf("OCR did not recognize the synthetic total; OCR text: %q", ocrText)
	}
	t.Logf("OCR: pass, latency=%s, chars=%d, recognized mock total", ocrLatency.Round(time.Millisecond), len(ocrText))

	now := time.Date(2026, 9, 22, 10, 0, 0, 0, time.FixedZone("ICT", 7*60*60))
	wallets := []entity.Wallet{
		{ID: "wallet-cash", Name: "Ví tiền mặt", Type: entity.WalletTypeBasic, Currency: "VND"},
		{ID: "wallet-bank", Name: "Tài khoản ngân hàng", Type: entity.WalletTypeBasic, Currency: "VND"},
	}
	categories := []entity.Category{
		{ID: "category-food", Name: "Ăn uống", Kind: entity.TransactionTypeExpense},
		{ID: "category-coffee", Name: "Cà phê", Kind: entity.TransactionTypeExpense},
		{ID: "category-salary", Name: "Lương", Kind: entity.TransactionTypeIncome},
	}
	type evaluationCase struct {
		name          string
		text          string
		count         int
		types         []string
		amounts       []int64
		walletID      string
		datePrefix    string
		needsQuestion bool
	}
	cases := []evaluationCase{
		{name: "single expense / VND shorthand", text: "Hôm nay ăn trưa hết 35.000đ bằng Ví tiền mặt.", count: 1, types: []string{"expense"}, amounts: []int64{35000}, walletID: "wallet-cash", datePrefix: "2026-09-22"},
		{name: "two independent expenses / relative date", text: "Hôm qua ăn sáng 45.000đ và uống cà phê 30.000đ bằng Ví tiền mặt.", count: 2, types: []string{"expense", "expense"}, amounts: []int64{45000, 30000}, walletID: "wallet-cash", datePrefix: "2026-09-21"},
		{name: "income / large VND amount", text: "Ngày 20/09/2026 nhận lương 15.000.000đ vào Tài khoản ngân hàng.", count: 1, types: []string{"income"}, amounts: []int64{15000000}, walletID: "wallet-bank", datePrefix: "2026-09-20"},
		{name: "internal transfer classification", text: "Chuyển 500.000đ từ Tài khoản ngân hàng sang Ví tiền mặt.", count: 1, types: []string{"transfer"}, amounts: []int64{500000}, needsQuestion: true},
		{name: "OCR receipt image", text: "Tạo một giao dịch từ hóa đơn thử nghiệm sau, thanh toán bằng Ví tiền mặt. Nội dung OCR: " + ocrText, count: 1, types: []string{"expense"}, amounts: []int64{45000}, walletID: "wallet-cash", datePrefix: "2026-09-22"},
	}

	latencies := make([]time.Duration, 0, len(cases))
	checked, passed := 0, 0
	transferSafe := false
	for _, tc := range cases {
		start := time.Now()
		out, err := client.Extract(ctx, Input{Text: tc.text, Timezone: "Asia/Ho_Chi_Minh", Now: now, Wallets: wallets, Categories: categories})
		latency := time.Since(start)
		latencies = append(latencies, latency)
		if err != nil {
			t.Errorf("case %q failed: %v (latency=%s)", tc.name, err, latency.Round(time.Millisecond))
			checked += 3
			continue
		}
		fields := 0
		fieldPass := 0
		check := func(ok bool) {
			fields++
			if ok {
				fieldPass++
			}
		}
		check(len(out.Drafts) == tc.count)
		check(equalStrings(draftTypes(out.Drafts), tc.types))
		check(equalAmounts(draftAmounts(out.Drafts), tc.amounts))
		if tc.walletID != "" {
			allWalletsMatch := len(out.Drafts) == tc.count
			for _, draft := range out.Drafts {
				allWalletsMatch = allWalletsMatch && draft.WalletID == tc.walletID
			}
			check(allWalletsMatch)
		}
		if tc.datePrefix != "" {
			allDatesMatch := len(out.Drafts) == tc.count
			for _, draft := range out.Drafts {
				allDatesMatch = allDatesMatch && strings.HasPrefix(draft.OccurredAt, tc.datePrefix)
			}
			check(allDatesMatch)
		}
		if tc.needsQuestion {
			asked := strings.TrimSpace(out.Reply) != ""
			for _, draft := range out.Drafts {
				asked = asked || len(draft.Questions) > 0
			}
			check(asked)
			transferSafe = len(out.Drafts) == 1 && out.Drafts[0].Type == "transfer" && asked
		}
		checked += fields
		passed += fieldPass
		t.Logf("case=%q latency=%s field_score=%d/%d drafts=%d types=%v amounts=%v wallets=%v", tc.name, latency.Round(time.Millisecond), fieldPass, fields, len(out.Drafts), draftTypes(out.Drafts), draftAmounts(out.Drafts), draftWallets(out.Drafts))
	}
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	t.Logf("EVAL SUMMARY model=%s cases=%d field_accuracy=%d/%d (%.1f%%) llm_p50=%s llm_p95=%s ocr=%s pipeline_ocr_plus_receipt_llm=%s transfer_safety=%t", cfg.AIModel, len(cases), passed, checked, percent(passed, checked), percentile(latencies, 50).Round(time.Millisecond), percentile(latencies, 95).Round(time.Millisecond), ocrLatency.Round(time.Millisecond), (ocrLatency + latencies[len(latencies)-1]).Round(time.Millisecond), transferSafe)
	if percent(passed, checked) < 80 || !transferSafe {
		t.Errorf("quality gate failed: require >=80%% field accuracy and transfer clarification; got %.1f%%, transfer_safe=%t", percent(passed, checked), transferSafe)
	}
}

func draftTypes(drafts []entity.AIExtractDraft) []string {
	values := make([]string, 0, len(drafts))
	for _, draft := range drafts {
		values = append(values, draft.Type)
	}
	sort.Strings(values)
	return values
}

func draftAmounts(drafts []entity.AIExtractDraft) []int64 {
	values := make([]int64, 0, len(drafts))
	for _, draft := range drafts {
		values = append(values, draft.Amount)
	}
	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
	return values
}

func draftWallets(drafts []entity.AIExtractDraft) []string {
	values := make([]string, 0, len(drafts))
	for _, draft := range drafts {
		values = append(values, draft.WalletID)
	}
	return values
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	a, b := append([]string(nil), left...), append([]string(nil), right...)
	sort.Strings(a)
	sort.Strings(b)
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func equalAmounts(left, right []int64) bool {
	if len(left) != len(right) {
		return false
	}
	a, b := append([]int64(nil), left...), append([]int64(nil), right...)
	sort.Slice(a, func(i, j int) bool { return a[i] < a[j] })
	sort.Slice(b, func(i, j int) bool { return b[i] < b[j] })
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func percent(passed, checked int) float64 {
	if checked == 0 {
		return 0
	}
	return float64(passed) * 100 / float64(checked)
}

func percentile(values []time.Duration, p int) time.Duration {
	if len(values) == 0 {
		return 0
	}
	index := (p*len(values)+99)/100 - 1
	if index < 0 {
		index = 0
	}
	if index >= len(values) {
		index = len(values) - 1
	}
	return values[index]
}
