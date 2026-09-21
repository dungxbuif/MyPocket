package entity

import "time"

const (
	AIProposalPending  = "pending"
	AIProposalApproved = "approved"
	AIProposalRejected = "rejected"
)

type AIEntryDraft struct {
	Type              string  `json:"type"`
	Amount            int64   `json:"amount"`
	WalletID          string  `json:"wallet_id"`
	CategoryID        *string `json:"category_id"`
	OccurredAt        string  `json:"occurred_at"`
	Note              string  `json:"note"`
	IncludedInReports bool    `json:"included_in_reports"`
}

type AIExtractDraft struct {
	AIEntryDraft
	Questions []string `json:"questions"`
}
type AIImage struct {
	Name     string `json:"name"`
	MIMEType string `json:"mime_type"`
	Base64   string `json:"base64"`
}
type AIHistoryMessage struct{ Role, Content string }
type AIExtractInput struct {
	Text, Timezone string
	Now            time.Time
	Wallets        []Wallet
	Categories     []Category
	History        []AIHistoryMessage
	Images         []AIImage
}
type AIExtractOutput struct {
	Reply      string
	Drafts     []AIExtractDraft
	SourceText string
}

type AIEntrySession struct {
	ID              string            `json:"id" gorm:"primaryKey"`
	OwnerID         string            `json:"-"`
	RequestToken    string            `json:"-"`
	ProcessingUntil *time.Time        `json:"-"`
	Processing      bool              `json:"processing" gorm:"-"`
	Error           string            `json:"error,omitempty"`
	Messages        []AIEntryMessage  `json:"messages" gorm:"-"`
	Proposals       []AIEntryProposal `json:"proposals" gorm:"-"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}
type AIEntryMessage struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	SessionID  string    `json:"-"`
	Role       string    `json:"role"`
	Content    string    `json:"content"`
	SourceText string    `json:"-"`
	CreatedAt  time.Time `json:"created_at"`
}
type AIEntryProposal struct {
	ID            string       `json:"id" gorm:"primaryKey"`
	SessionID     string       `json:"session_id"`
	OwnerID       string       `json:"-"`
	Version       int          `json:"version"`
	Status        string       `json:"status"`
	Draft         AIEntryDraft `json:"draft" gorm:"serializer:json;type:jsonb"`
	Questions     []string     `json:"questions" gorm:"serializer:json;type:jsonb"`
	TransactionID *string      `json:"transaction_id,omitempty"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}
type AIEntryRequest struct {
	SessionID string `gorm:"primaryKey"`
	RequestID string `gorm:"primaryKey"`
	Hash      string
	Token     string
	CreatedAt time.Time
}
