package usecase

import (
	"testing"

	"github.com/mypocket/backend/internal/entity"
)

func TestHasJarContextAndNoEvidenceReply(t *testing.T) {
	with := &entity.AdvisorMessage{Parts: []entity.AdvisorPart{{Type: "text", Text: "q"}, {Type: "jar_context", Data: map[string]any{"month": "2026-10"}}}}
	without := &entity.AdvisorMessage{Parts: []entity.AdvisorPart{{Type: "text", Text: "q"}}}
	if !hasJarContext(with) || hasJarContext(without) || hasJarContext(nil) {
		t.Fatal("hasJarContext must detect only jar_context messages")
	}
	if _, _, ok := jarContextSuggestions(with); ok {
		t.Fatal("a jar_context without suggestions is not a fallback candidate")
	}
	if jarNoEvidenceReply == "" {
		t.Fatal("reply must exist")
	}
}
