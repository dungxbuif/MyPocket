package entity

// AdvisorClaim binds a rendered answer slot to a fact returned by a tool in
// the same run. The model supplies references; the server supplies values.
type AdvisorClaim struct {
	Slot    string `json:"slot"`
	FactRef string `json:"fact_ref"`
}

type AdvisorFinalAnswer struct {
	Mode        string         `json:"mode"`
	TemplateID  string         `json:"template_id,omitempty"`
	Claims      []AdvisorClaim `json:"claims,omitempty"`
	Explanation string         `json:"explanation,omitempty"`
	SourceID    string         `json:"source_id,omitempty"`
	QuestionID  string         `json:"question_id,omitempty"`
}
