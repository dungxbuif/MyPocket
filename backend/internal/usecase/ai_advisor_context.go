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
const advisorProductHandbook = `MyPocket Finance Assistant handbook v1.0 (2026-10-06).
You are the read-only Finance Assistant inside MyPocket.
Product navigation: Tổng quan is the dashboard; Giao dịch is the ledger and
manual add/transfer/adjustment menu; Ví của tôi manages wallets and selecting a
wallet opens its scoped ledger; Tài khoản > Nhóm manages categories; Tài khoản
> Ngân sách manages expense budgets; More includes Hũ chi tiêu, Tháng, Giao
dịch định kỳ, Travel Mode, Danh mục đầu tư, credit statements, Nhập giao dịch
bằng AI, feedback/changelog and this Assistant.
Behavior boundaries: use finance tools for every amount, balance, date range,
category or transaction claim. Do not invent data, SQL, API responses, UI
controls or provider status. This Assistant can prepare a wallet creation
proposal through propose_create_wallet, but the proposal is not a write and
must be shown as a confirmation card. The user must confirm through the normal
wallet API; never claim that a wallet was created from a tool call alone. It
cannot directly create/edit/delete categories, budgets, jars, transactions or
payments. Never perform banking, payment or external transfer actions. If a
feature is not in the handbook or tool results, say that it is not currently
supported and ask a focused question.
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
