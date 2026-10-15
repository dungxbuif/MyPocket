package usecase

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"unicode"

	"github.com/mypocket/backend/internal/entity"
	"golang.org/x/text/unicode/norm"
)

const advisorHelpVersion = "1.1"

type advisorHelpSection struct {
	ID       string
	Title    string
	Excerpt  string
	Keywords []string
}

// Keep this catalog deliberately small and source-controlled. It is product
// documentation, not a place to mirror user data or provider instructions.
var advisorHelpSections = []advisorHelpSection{
	{ID: "handbook.navigation.v1.1", Title: "Điều hướng chính", Excerpt: "Tổng quan là dashboard; Giao dịch là sổ giao dịch; Ví của tôi mở sổ theo ví; Trợ lý là tab chính; Tài khoản chứa ngân sách, hũ chi tiêu và các công cụ quản lý.", Keywords: []string{"điều hướng", "tab", "màn hình", "trợ lý", "tổng quan", "giao dịch", "ví", "tài khoản"}},
	{ID: "handbook.transactions.v1.1", Title: "Thêm và sửa giao dịch", Excerpt: "Mở Giao dịch để thêm khoản thu, chi, chuyển ví hoặc điều chỉnh. Kiểm tra ví, ngày, số tiền và danh mục trước khi lưu; chuyển ví tạo một cặp giao dịch.", Keywords: []string{"thêm giao dịch", "sửa giao dịch", "chuyển ví", "điều chỉnh", "giao dịch"}},
	{ID: "handbook.budgets.v1.1", Title: "Ngân sách", Excerpt: "Mở Tài khoản > Ngân sách để tạo giới hạn chi theo danh mục và thời gian. Báo cáo dùng các giao dịch được đánh dấu tính vào báo cáo.", Keywords: []string{"ngân sách", "budget", "giới hạn", "danh mục", "báo cáo"}},
	{ID: "handbook.wallets.v1.1", Title: "Quản lý ví", Excerpt: "Ví của tôi quản lý ví thường, ví tiết kiệm/mục tiêu và ví tín dụng. Chọn một ví để xem số dư và các giao dịch thuộc ví đó.", Keywords: []string{"ví", "tiết kiệm", "mục tiêu", "tín dụng", "số dư"}},
	{ID: "handbook.assistant.v1.1", Title: "Finance Assistant", Excerpt: "Trợ lý chỉ đọc số liệu qua các tool owner-scoped. Đề xuất tạo ví chỉ được lưu sau khi bạn xác nhận trên thẻ; trợ lý không tự tạo giao dịch hay thực hiện thanh toán.", Keywords: []string{"ai", "assistant", "trợ lý", "tool", "xác nhận", "tạo ví", "thanh toán"}},
}

type productHelpArgs struct {
	Query      string `json:"query"`
	MaxResults int    `json:"max_results"`
}

type productHelpSource struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Excerpt string `json:"excerpt"`
}

type productHelpView struct {
	Version string              `json:"version"`
	Sources []productHelpSource `json:"sources"`
}

type productHelpTool struct{}

func (t *productHelpTool) Name() string { return "get_product_help" }

func (t *productHelpTool) Definition() ToolDefinition {
	return readDefinition(t.Name(), "Find a short, versioned answer in the MyPocket product handbook", map[string]any{
		"query":       map[string]any{"type": "string", "minLength": 1, "maxLength": 240},
		"max_results": map[string]any{"type": "integer", "minimum": 1, "maximum": 5},
	}, []string{"query"})
}

func (t *productHelpTool) Execute(_ context.Context, _ Principal, raw json.RawMessage) (entity.FinanceResult, error) {
	var args productHelpArgs
	if err := DecodeAdvisorToolArgs(raw, &args); err != nil {
		return entity.FinanceResult{}, err
	}
	args.Query = strings.TrimSpace(args.Query)
	if args.Query == "" || len(args.Query) > 240 || args.MaxResults < 0 || args.MaxResults > 5 {
		return entity.FinanceResult{}, ErrAdvisorToolInvalid
	}
	if args.MaxResults == 0 {
		args.MaxResults = 3
	}
	queryTokens := helpTokens(args.Query)
	if len(queryTokens) == 0 {
		return entity.FinanceResult{}, ErrAdvisorToolInvalid
	}
	type scoredSection struct {
		section advisorHelpSection
		score   int
		index   int
	}
	scored := make([]scoredSection, 0, len(advisorHelpSections))
	for index, section := range advisorHelpSections {
		searchTokens := make(map[string]struct{})
		for _, token := range helpTokens(section.Title + " " + section.Excerpt + " " + strings.Join(section.Keywords, " ")) {
			searchTokens[token] = struct{}{}
		}
		score := 0
		for _, token := range queryTokens {
			if _, ok := searchTokens[token]; ok {
				score++
			}
		}
		if score > 0 {
			scored = append(scored, scoredSection{section: section, score: score, index: index})
		}
	}
	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].score != scored[j].score {
			return scored[i].score > scored[j].score
		}
		return scored[i].index < scored[j].index
	})
	sources := make([]productHelpSource, 0, minInt(args.MaxResults, len(scored)))
	for _, item := range scored[:minInt(args.MaxResults, len(scored))] {
		sources = append(sources, productHelpSource{ID: item.section.ID, Title: item.section.Title, Excerpt: item.section.Excerpt})
	}
	view, err := json.Marshal(productHelpView{Version: advisorHelpVersion, Sources: sources})
	if err != nil {
		return entity.FinanceResult{}, ErrAdvisorToolInvalid
	}
	status := "ok"
	if len(sources) == 0 {
		status = "not_found"
	}
	return entity.FinanceResult{QueryKey: "product_help", Status: status, ViewKind: "product_help", View: view}, nil
}

func helpNormalize(value string) string {
	value = norm.NFD.String(strings.ToLower(strings.TrimSpace(value)))
	var builder strings.Builder
	for _, runeValue := range value {
		switch {
		case unicode.Is(unicode.Mn, runeValue):
		case unicode.IsLetter(runeValue) || unicode.IsDigit(runeValue):
			builder.WriteRune(runeValue)
		default:
			builder.WriteByte(' ')
		}
	}
	return strings.Join(strings.Fields(builder.String()), " ")
}

func helpTokens(value string) []string {
	normalized := helpNormalize(value)
	stopWords := map[string]struct{}{
		"co": {}, "cho": {}, "cua": {}, "la": {}, "lam": {}, "nao": {},
		"nang": {}, "nhu": {}, "sao": {}, "the": {}, "thoi": {}, "toi": {},
		"tai": {}, "tinh": {}, "ton": {}, "khong": {}, "va": {},
	}
	seen := make(map[string]struct{})
	result := make([]string, 0, 8)
	for _, token := range strings.Fields(normalized) {
		if len([]rune(token)) < 2 {
			continue
		}
		if _, ok := stopWords[token]; ok {
			continue
		}
		if _, ok := seen[token]; ok {
			continue
		}
		seen[token] = struct{}{}
		result = append(result, token)
	}
	return result
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}
