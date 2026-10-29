package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mypocket/backend/internal/entity"
	"github.com/mypocket/backend/internal/usecase"
)

type pagedAdvisorActionStore struct{ advisorActionStore }

func (s *pagedAdvisorActionStore) ListMessages(_ context.Context, ownerID, conversationID string, beforeSeq int64, limit int) ([]entity.AdvisorMessage, error) {
	var page []entity.AdvisorMessage
	for _, message := range s.messages {
		if ownerID == "owner-1" && message.ConversationID == conversationID && (beforeSeq == 0 || message.Seq < beforeSeq) {
			page = append(page, message)
			if len(page) == limit {
				break
			}
		}
	}
	return page, nil
}

func TestAdvisorConfirmOlderJarProposalKeepsOwnerGenerationAndReplay(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name       string
		owner      string
		generation int
		want       int
	}{
		{"current owner", "owner-1", 4, http.StatusCreated},
		{"foreign owner", "owner-2", 4, http.StatusNotFound},
		{"stale generation", "owner-1", 3, http.StatusConflict},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := &pagedAdvisorActionStore{}
			for seq := int64(61); seq > 1; seq-- {
				store.messages = append(store.messages, entity.AdvisorMessage{Seq: seq, ConversationID: "conversation-1", Generation: 4, Role: entity.AdvisorRoleUser, Parts: []entity.AdvisorPart{{Type: "text", Text: "later question"}}})
			}
			store.messages = append(store.messages, entity.AdvisorMessage{
				Seq: 1, ConversationID: "conversation-1", Generation: 4, Role: entity.AdvisorRoleAssistant,
				Parts: []entity.AdvisorPart{{Type: "jar_confirmation", Data: map[string]any{"view": map[string]any{
					"action_id": "old-jar-action", "action": "create_jar", "conversation_id": "conversation-1", "generation": int64(4),
					"month": "2026-10", "name": "Ăn uống", "allocation_mode": "fixed", "allocation_amount": int64(400000),
				}}}},
			})
			jars := &advisorJarRepositoryStub{}
			engine := gin.New()
			engine.Use(func(c *gin.Context) { c.Set(contextUserIDKey, tc.owner); c.Next() })
			engine.POST("/actions/:action_id/confirm", (&AdvisorHandler{Service: &usecase.AdvisorService{Store: store}, Users: advisorOverviewUsers{}, Jars: jars}).ConfirmAction)
			confirm := func() *httptest.ResponseRecorder {
				recorder := httptest.NewRecorder()
				request := httptest.NewRequest(http.MethodPost, "/actions/old-jar-action/confirm", strings.NewReader(fmt.Sprintf(`{"conversation_id":"conversation-1","generation":%d}`, tc.generation)))
				request.Header.Set("Content-Type", "application/json")
				engine.ServeHTTP(recorder, request)
				return recorder
			}
			recorder := confirm()
			if recorder.Code != tc.want {
				t.Fatalf("older proposal status=%d want=%d body=%s", recorder.Code, tc.want, recorder.Body.String())
			}
			if tc.want != http.StatusCreated {
				if jars.created != nil {
					t.Fatal("rejected confirmation created a jar")
				}
				return
			}
			if jars.created == nil || jars.created.Name != "Ăn uống" || jars.idempotencyKey != "old-jar-action" {
				t.Fatalf("older confirmation did not use persisted proposal/key: %+v", jars.created)
			}
			if replay := confirm(); replay.Code != http.StatusOK {
				t.Fatalf("older proposal replay=%d body=%s", replay.Code, replay.Body.String())
			}
		})
	}
}
