package usecase

import (
	"sort"

	"github.com/mypocket/backend/internal/entity"
)

const advisorContextMessageLimit = 12

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
