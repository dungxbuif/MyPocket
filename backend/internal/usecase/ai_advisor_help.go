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

const advisorHelpVersion = "1.3"

type advisorHelpSection struct {
	ID       string
	Title    string
	Excerpt  string
	Keywords []string
}

// Keep this catalog deliberately small and source-controlled. It is product
// documentation, not a place to mirror user data or provider instructions.
var advisorHelpSections = []advisorHelpSection{
	{ID: "handbook.navigation.v1.3", Title: "Điều hướng chính", Excerpt: "Tổng quan là dashboard; Giao dịch là sổ giao dịch; Trợ lý là tab chính; Tài khoản mở các màn quản lý như Ví, Ngân sách, Báo cáo, Danh mục, Hũ chi tiêu, Tháng, Giao dịch định kỳ, Travel Mode, Danh mục đầu tư, Sổ tín dụng, Thông báo, Nhập giao dịch bằng AI và Phản hồi & changelog.", Keywords: []string{"điều hướng", "tab", "màn hình", "trợ lý", "tổng quan", "giao dịch", "ví", "tài khoản", "more", "navigation"}},
	{ID: "handbook.transactions.v1.3", Title: "Thêm và sửa giao dịch", Excerpt: "Mở Giao dịch để thêm khoản thu, chi, chuyển ví hoặc điều chỉnh. Kiểm tra ví, ngày, số tiền và danh mục trước khi lưu; chuyển ví tạo một cặp giao dịch. Giao dịch liên kết tín dụng được xử lý trong Sổ tín dụng.", Keywords: []string{"thêm giao dịch", "sửa giao dịch", "xóa giao dịch", "chuyển ví", "điều chỉnh", "giao dịch", "thu", "chi", "transfer"}},
	{ID: "handbook.wallets.v1.3", Title: "Quản lý ví", Excerpt: "Mở Ví của tôi để quản lý ví thường, ví tiết kiệm/mục tiêu và ví tín dụng. Chọn một ví để xem số dư và sổ giao dịch theo ví; ví mục tiêu có số tiền đích, ví tín dụng có hạn mức và kỳ sao kê.", Keywords: []string{"ví", "ví của tôi", "tiết kiệm", "mục tiêu", "tín dụng", "số dư", "wallet", "credit"}},
	{ID: "handbook.budgets.v1.3", Title: "Ngân sách", Excerpt: "Mở Tài khoản > Ngân sách để tạo giới hạn chi theo danh mục, ví và khoảng thời gian. Ngân sách hỗ trợ tuần, tháng, quý hoặc năm; báo cáo dùng các giao dịch được đánh dấu tính vào báo cáo.", Keywords: []string{"ngân sách", "budget", "giới hạn", "hạn mức", "danh mục", "ví", "tuần", "tháng", "quý", "năm"}},
	{ID: "handbook.reports.v1.3", Title: "Báo cáo", Excerpt: "Mở Tài khoản > Báo cáo để xem tổng thu, chi, ròng, xu hướng và phân bổ theo danh mục. Chỉ giao dịch được đánh dấu tính vào báo cáo được tổng hợp; chuyển ví và điều chỉnh số dư không làm tăng chi tiêu.", Keywords: []string{"báo cáo", "report", "thống kê", "phân tích", "xu hướng", "tổng thu", "tổng chi", "cashflow"}},
	{ID: "handbook.categories.v1.3", Title: "Danh mục", Excerpt: "Mở Tài khoản > Danh mục để quản lý nhóm thu, chi hoặc nợ, cây danh mục cha-con, phạm vi ví, biểu tượng và lưu trữ. Danh mục hệ thống không thể sửa tùy ý.", Keywords: []string{"danh mục", "category", "nhóm", "danh mục cha", "danh mục con", "phân loại", "lưu trữ"}},
	{ID: "handbook.jars.v1.3", Title: "Hũ chi tiêu", Excerpt: "Mở Tài khoản > Hũ chi tiêu để quản lý hũ theo tháng, phân bổ không dùng phân bổ, số tiền cố định hoặc tỷ lệ phần trăm. Hũ chỉ là kế hoạch phân bổ và không thay đổi số dư ví; yêu cầu Trợ lý tạo hoặc sửa hũ cần kiểm tra rồi xác nhận trên thẻ.", Keywords: []string{"hũ", "hũ chi tiêu", "jar", "jars", "phân bổ", "tạo hũ", "sửa hũ", "allocation"}},
	{ID: "handbook.months.v1.3", Title: "Tổng hợp theo tháng", Excerpt: "Mở Tài khoản > Tháng để xem tổng thu, chi, ròng, số giao dịch, chi tiết hũ và ghi chú của một tháng tài chính.", Keywords: []string{"tháng", "month", "tổng hợp tháng", "thu nhập tháng", "chi tiêu tháng", "ghi chú tháng"}},
	{ID: "handbook.recurring.v1.3", Title: "Giao dịch định kỳ", Excerpt: "Mở Tài khoản > Giao dịch định kỳ để tạo lịch thu hoặc chi với ví, loại, danh mục, tần suất, lần chạy tiếp theo, ngày kết thúc, trạng thái và ghi chú. Lịch chỉ tạo giao dịch thường khi người dùng chạy kỳ đến hạn; mỗi kỳ không tạo trùng.", Keywords: []string{"giao dịch định kỳ", "recurring", "lặp lại", "tần suất", "lịch", "subscription", "đến hạn"}},
	{ID: "handbook.travel.v1.3", Title: "Travel Mode", Excerpt: "Travel Mode gắn giao dịch thu/chi thường mới vào chuyến đang bật; đây không phải ví và không đổi số dư. Chuyển ví, điều chỉnh và giao dịch định kỳ không tự gắn chuyến.", Keywords: []string{"travel", "travel mode", "chuyến đi", "du lịch", "gắn chuyến"}},
	{ID: "handbook.portfolio.v1.3", Title: "Danh mục đầu tư", Excerpt: "Mở Tài khoản > Danh mục đầu tư để theo dõi tài sản, giao dịch đầu tư, giá vốn, giá trị và lãi/lỗ chưa thực hiện. Trợ lý không có proposal sửa giao dịch đầu tư trực tiếp.", Keywords: []string{"đầu tư", "portfolio", "tài sản", "giao dịch đầu tư", "giá vốn", "lãi lỗ"}},
	{ID: "handbook.credit.v1.3", Title: "Sổ tín dụng", Excerpt: "Mở Tài khoản > Sổ tín dụng để xem sao kê ví tín dụng, dư nợ, hạn mức còn lại, số phải trả và các khoản mua, hoàn tiền, phí hoặc lãi. Thanh toán tín dụng dùng luồng riêng; Trợ lý không tự thanh toán.", Keywords: []string{"tín dụng", "sổ tín dụng", "credit", "sao kê", "dư nợ", "hạn mức", "thanh toán", "phí", "lãi"}},
	{ID: "handbook.notifications.v1.3", Title: "Thông báo", Excerpt: "Mở Tài khoản > Thông báo để xem các thông báo trong ứng dụng và đánh dấu đã đọc. Trợ lý không tự gửi thông báo hoặc APNs.", Keywords: []string{"thông báo", "notification", "đã đọc", "chưa đọc", "apns"}},
	{ID: "handbook.assistant.v1.3", Title: "Finance Assistant", Excerpt: "Trợ lý là tab chính, đọc số liệu owner-scoped và có thể chuẩn bị đề xuất tạo ví, hũ, ngân sách, giao dịch, chuyển ví hoặc danh mục. Kiểm tra rồi xác nhận trên thẻ để lưu; trợ lý không tự ghi dữ liệu, thanh toán hay chuyển tiền bên ngoài.", Keywords: []string{"ai", "assistant", "trợ lý", "tool", "xác nhận", "tạo ví", "tạo giao dịch", "phân tích", "thanh toán"}},
	{ID: "handbook.ai_entry.v1.3", Title: "Nhập giao dịch bằng AI", Excerpt: "Mở Tài khoản > Nhập giao dịch bằng AI để gửi mô tả hoặc ảnh/PDF chứng từ. AI chỉ tạo danh sách giao dịch đề xuất; xem, sửa, chọn ví/danh mục và duyệt từng dòng trước khi lưu.", Keywords: []string{"nhập giao dịch bằng ai", "ai entry", "ocr", "ảnh", "pdf", "chứng từ", "đề xuất giao dịch", "duyệt"}},
	{ID: "handbook.feedback.v1.3", Title: "Phản hồi và changelog", Excerpt: "Mở Tài khoản > Phản hồi & changelog để gửi lỗi, yêu cầu tính năng hoặc cải tiến và xem trạng thái xử lý cùng ghi chú phát hành. Đây là luồng phản hồi, không phải luồng sửa dữ liệu tài chính.", Keywords: []string{"phản hồi", "feedback", "changelog", "release note", "báo lỗi", "yêu cầu tính năng", "cải tiến"}},
	{ID: "handbook.account.v1.3", Title: "Tài khoản và bảo mật", Excerpt: "Mở Tài khoản để sửa hồ sơ, múi giờ, quản lý API key, bật Face ID và đăng xuất. API key có cùng quyền với phiên đăng nhập của chủ tài khoản; giữ kín và thu hồi khi không dùng. Face ID mở phiên đã lưu trong Keychain.", Keywords: []string{"tài khoản", "account", "hồ sơ", "múi giờ", "timezone", "api key", "face id", "đăng xuất", "bảo mật"}},
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
		titleTokens := make(map[string]struct{})
		for _, token := range helpTokens(section.Title) {
			titleTokens[token] = struct{}{}
		}
		for _, token := range helpTokens(section.Title + " " + section.Excerpt + " " + strings.Join(section.Keywords, " ")) {
			searchTokens[token] = struct{}{}
		}
		score := 0
		for _, token := range queryTokens {
			if _, ok := searchTokens[token]; ok {
				score++
			}
			// Prefer the feature's own instructions over a navigation section
			// that merely mentions the same feature in passing.
			if _, ok := titleTokens[token]; ok {
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
