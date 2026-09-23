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
	JarID             *string `json:"jar_id,omitempty"`
	OccurredAt        string  `json:"occurred_at"`
	Note              string  `json:"note"`
	IncludedInReports bool    `json:"included_in_reports"`
}

type AIExtractDraft struct {
	AIEntryDraft
	Questions []string `json:"questions"`
}
type AIImage struct {
	Name         string `json:"name"`
	MIMEType     string `json:"mime_type"`
	Base64       string `json:"base64"`
	SourceURL    string `json:"-"`
	AttachmentID string `json:"-"`
}

type AIEntryAttachment struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	OwnerID     string    `json:"-"`
	ProcessID   string    `json:"-" gorm:"column:session_id"`
	ObjectKey   string    `json:"-"`
	Filename    string    `json:"filename"`
	MIMEType    string    `json:"mime_type"`
	SizeBytes   int64     `json:"size_bytes"`
	SHA256      string    `json:"-"`
	OCRStatus   string    `json:"-"`
	OCRText     string    `json:"-"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"-"`
	DeleteAfter time.Time `json:"-"`
}

func (AIEntryAttachment) TableName() string { return "transaction_attachments" }

type AIEntryAttachmentLink struct {
	TransactionID string `gorm:"primaryKey"`
	AttachmentID  string `gorm:"primaryKey"`
	CreatedAt     time.Time
}

func (AIEntryAttachmentLink) TableName() string { return "transaction_attachment_links" }

type AIExtractInput struct {
	Text, Instruction, Timezone string
	Now                         time.Time
	Wallets                     []Wallet
	Categories                  []Category
	Images                      []AIImage
}
type AIExtractOutput struct {
	Reply           string
	Drafts          []AIExtractDraft
	SourceText      string
	AttachmentTexts []string
	OCRComplete     bool
	ModelUsage      map[string]any
}

type AIEntrySession struct {
	ID              string            `json:"id" gorm:"primaryKey"`
	OwnerID         string            `json:"-"`
	RequestToken    string            `json:"-"`
	ProcessingUntil *time.Time        `json:"-"`
	Processing      bool              `json:"processing" gorm:"-"`
	Error           string            `json:"error,omitempty"`
	ErrorCode       string            `json:"error_code,omitempty"`
	Reply           string            `json:"reply,omitempty"`
	ModelUsage      map[string]any    `json:"-" gorm:"serializer:json;type:jsonb"`
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
	SessionID     string       `json:"process_id"`
	OwnerID       string       `json:"-"`
	Version       int          `json:"version"`
	Status        string       `json:"status"`
	Draft         AIEntryDraft `json:"draft" gorm:"serializer:json;type:jsonb"`
	Questions     []string     `json:"questions" gorm:"serializer:json;type:jsonb"`
	TransactionID *string      `json:"transaction_id,omitempty"`
	AttachmentIDs []string     `json:"attachment_ids,omitempty" gorm:"-"`
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
