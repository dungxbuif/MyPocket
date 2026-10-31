package entity

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAIEntryAttachmentDiagnosticMapsSafePerFileState(t *testing.T) {
	tests := []struct {
		name       string
		attachment AIEntryAttachment
		want       AIEntryAttachmentDiagnostic
	}{
		{
			name:       "completed OCR",
			attachment: AIEntryAttachment{InputOrder: 1, Filename: "second.pdf", MIMEType: "application/pdf", OCRStatus: "completed"},
			want:       AIEntryAttachmentDiagnostic{Position: 2, Filename: "second.pdf", Stage: "ocr", Status: "succeeded"},
		},
		{
			name:       "failed OCR",
			attachment: AIEntryAttachment{InputOrder: 0, Filename: "first.pdf", MIMEType: "application/pdf", OCRStatus: "failed", OCRText: "private OCR"},
			want:       AIEntryAttachmentDiagnostic{Position: 1, Filename: "first.pdf", Stage: "ocr", Status: "failed", ErrorCode: "ocr_failed"},
		},
		{
			name:       "structured file",
			attachment: AIEntryAttachment{InputOrder: 2, Filename: "statement.xlsx", MIMEType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", OCRStatus: "completed"},
			want:       AIEntryAttachmentDiagnostic{Position: 3, Filename: "statement.xlsx", Stage: "validation", Status: "succeeded"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.attachment.Diagnostic(); got != tt.want {
				t.Fatalf("Diagnostic() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestAIEntryAttachmentDiagnosticDoesNotExposePrivateStorageOrOCRFields(t *testing.T) {
	diagnostic := (AIEntryAttachment{
		InputOrder: 0, Filename: "bank.pdf", MIMEType: "application/pdf", OCRStatus: "failed",
		ObjectKey: "private/owner/secret", OCRText: "salary 90000000", SHA256: strings.Repeat("a", 64),
	}).Diagnostic()
	encoded, err := json.Marshal(diagnostic)
	if err != nil {
		t.Fatal(err)
	}
	body := string(encoded)
	for _, forbidden := range []string{"private/owner/secret", "salary 90000000", strings.Repeat("a", 64), "attachment_id", "object_key"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("diagnostic leaked %q: %s", forbidden, body)
		}
	}
	if !strings.Contains(body, `"position":1`) || !strings.Contains(body, `"error_code":"ocr_failed"`) {
		t.Fatalf("diagnostic omitted safe fields: %s", body)
	}
}
