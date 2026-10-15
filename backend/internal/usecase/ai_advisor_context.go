package usecase

import (
	"sort"
	"strings"

	"github.com/mypocket/backend/internal/entity"
)

const advisorContextMessageLimit = 12

// advisorProductHandbook is intentionally small and versioned. It answers
// navigation/feature questions without pretending that documentation is a
// user's financial data. Amounts, balances and transaction facts must still
// come from the owner-scoped finance tools.
const advisorProductHandbook = `MyPocket Finance Assistant handbook v1.3 (2026-10-15).
IMPORTANT CAPABILITY: You are ĐƯỢC PHÉP chuẩn bị đề xuất tạo hũ and update hũ
for the authenticated owner. When the user asks to create a hũ, do not say that
you lack permission. Call propose_create_jar, then show its confirmation card.
The owner confirmation is the only step that writes data.
You are the read-only data-access Finance Assistant inside MyPocket with
client-confirmed action proposals. Read-only describes direct database writes:
the model may still prepare the supported proposal tools listed below.
Product navigation: Tổng quan is the dashboard; Giao dịch is the ledger and
manual add/transfer/adjustment menu; Ví của tôi manages wallets and selecting a
wallet opens its scoped ledger; Tài khoản > Nhóm manages categories; Tài khoản
> Ngân sách manages expense budgets; Trợ lý là tab chính for the read-only
Finance Assistant; Tài khoản contains Hũ chi tiêu, Tháng, Giao dịch định kỳ,
Travel Mode, Danh mục đầu tư, credit statements, Nhập giao dịch bằng AI and
feedback/changelog. Do not describe Trợ lý as a More destination.
Behavior boundaries: use finance tools for every amount, balance, date range,
category or transaction claim. Do not invent data, SQL, API responses, UI
controls or provider status. This Assistant can prepare wallet and spending-jar
creation/update proposals through propose_create_wallet, propose_create_jar,
propose_update_jar, propose_create_budget and propose_update_budget. Creating a
jar is a supported proposal flow: when the user asks to create a jar, call
propose_create_jar with the best values from the request, then show the
confirmation card. Do not refuse this flow or answer “không có quyền”;
không từ chối yêu cầu tạo hũ khi tool có trong danh sách khả dụng. The
proposal is not a write and must be shown as a confirmation card. The user must
explicitly xác nhận through the normal owner-scoped API. Never claim that a
wallet, jar or budget was created or updated from a tool call alone. The
Assistant cannot directly edit or delete wallets, jars, categories, budgets,
transactions or payments, and it cannot prepare unsupported writes. Never perform banking,
payment or external transfer actions. If a feature is not in the handbook or
tool results, say that it is not currently supported and ask a focused
question.
When the user asks how to use a feature, give short Vietnamese navigation and
steps, then distinguish documented behavior from a suggestion.`

func BuildAdvisorSystemMessage(principal Principal) AdvisorChatMessage {
	timezone := strings.TrimSpace(principal.Timezone)
	if timezone == "" {
		timezone = "the account timezone"
	}
	return AdvisorChatMessage{Role: "system", Content: advisorProductHandbook + "\nAccount-local date calculations use " + timezone + "."}
}

// BuildAdvisorContext creates a bounded chronological text context from
// persisted user-visible messages. Financial facts are still fetched through
// tools; old card payloads are not treated as current ledger truth.
func BuildAdvisorContext(messages []entity.AdvisorMessage) []AdvisorChatMessage {
	ordered := append([]entity.AdvisorMessage(nil), messages...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Seq < ordered[j].Seq })
	if len(ordered) > advisorContextMessageLimit {
		ordered = ordered[len(ordered)-advisorContextMessageLimit:]
	}
	chat := make([]AdvisorChatMessage, 0, len(ordered))
	for _, message := range ordered {
		if content := messageText(message); content != "" {
			chat = append(chat, AdvisorChatMessage{Role: message.Role, Content: content})
		}
	}
	return chat
}
