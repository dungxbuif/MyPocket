package entity

import "testing"

func TestFeedbackValidationAndTransitions(t *testing.T) {
	valid := Feedback{Type: FeedbackTypeBug, Title: "Wrong category", Description: "Grab is classified incorrectly."}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid feedback rejected: %v", err)
	}
	for _, pair := range [][2]string{
		{FeedbackStatusOpen, FeedbackStatusTriaged},
		{FeedbackStatusTriaged, FeedbackStatusInProgress},
		{FeedbackStatusInProgress, FeedbackStatusRejected},
		{FeedbackStatusInProgress, FeedbackStatusFixed},
	} {
		if !CanTransitionFeedback(pair[0], pair[1]) {
			t.Fatalf("expected transition %s -> %s", pair[0], pair[1])
		}
	}
	for _, pair := range [][2]string{
		{FeedbackStatusOpen, FeedbackStatusFixed},
		{FeedbackStatusFixed, FeedbackStatusInProgress},
		{FeedbackStatusRejected, FeedbackStatusOpen},
	} {
		if CanTransitionFeedback(pair[0], pair[1]) {
			t.Fatalf("unexpected transition %s -> %s", pair[0], pair[1])
		}
	}

	tooLong := valid
	tooLong.Title = string(make([]byte, 201))
	if err := tooLong.Validate(); err == nil {
		t.Fatal("title over 200 characters must be rejected")
	}
	tooLong = valid
	tooLong.Description = string(make([]byte, 10001))
	if err := tooLong.Validate(); err == nil {
		t.Fatal("description over 10000 characters must be rejected")
	}
}
