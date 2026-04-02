package observe

import "context"

type TriageResult struct {
	Category       string   `json:"category"`
	Severity       string   `json:"severity"`
	Confidence     float64  `json:"confidence"`
	Analysis       string   `json:"preliminary_analysis"`
	SuggestedRes   string   `json:"suggested_resolution"`
	RequiresReview bool     `json:"requires_human_review"`
	UrgencyFlag    bool     `json:"urgency_flag"`
	RelatedRefs    []string `json:"related_observations,omitempty"`
	FAQMatch       string   `json:"faq_match,omitempty"`
}

type TriageProvider interface {
	Triage(ctx context.Context, obs Observation, contextData *ContextData) (*TriageResult, error)
}
