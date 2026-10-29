package ai

import (
	"net/url"
	"strings"
	"testing"
)

func TestValidateImagesAcceptsTwentyPrivateReceiptFiles(t *testing.T) {
	images := make([]Image, 20)
	for i := range images {
		images[i] = Image{Name: "receipt.png", MIMEType: "image/png", SourceURL: "https://storage.example.test/object?signature=" + url.QueryEscape(string(rune('a'+i)))}
	}
	if err := validateImages(images); err != nil {
		t.Fatalf("twenty valid receipt files must be accepted: %v", err)
	}
}

func TestValidateImagesRejectsTheTwentyFirstReceiptFile(t *testing.T) {
	images := make([]Image, 21)
	for i := range images {
		images[i] = Image{Name: "receipt.png", MIMEType: "image/png", SourceURL: "https://storage.example.test/object?signature=x"}
	}
	if err := validateImages(images); err == nil {
		t.Fatal("the twenty-first receipt file must be rejected")
	} else if !strings.Contains(err.Error(), "20 receipt files") {
		t.Fatalf("limit error should explain the active limit: %v", err)
	}
}
