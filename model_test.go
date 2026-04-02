package observe

import (
	"encoding/json"
	"testing"
	"time"
)

func TestObservationJSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	resolvedAt := now.Add(time.Hour)
	faqID := int64(42)

	orig := Observation{
		ID:                  1,
		Ref:                 "OBS-0001",
		UserID:              10,
		DomainID:            20,
		Title:               "Email not loading",
		Description:         "When I open the webmail, nothing loads.",
		CategoryUser:        CategoryNotWorking,
		SeverityUser:        SeverityBlocking,
		Location:            LocationWebmail,
		AutoContext:         json.RawMessage(`{"browser":"Firefox","version":"130"}`),
		Triage:              json.RawMessage(`{"priority":"high"}`),
		Status:              StatusSubmitted,
		ResolutionInternal:  "Cache cleared",
		ResolutionUser:      "We fixed the issue.",
		RelatedObservations: json.RawMessage(`[2,3]`),
		FaqArticleID:        &faqID,
		CreatedAt:           now,
		UpdatedAt:           now,
		ResolvedAt:          &resolvedAt,
		UserEmail:           "user@example.com",
		DomainName:          "example.com",
		ScreenshotCount:     2,
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got Observation
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.ID != orig.ID {
		t.Errorf("ID: got %d, want %d", got.ID, orig.ID)
	}
	if got.Ref != orig.Ref {
		t.Errorf("Ref: got %q, want %q", got.Ref, orig.Ref)
	}
	if got.Title != orig.Title {
		t.Errorf("Title: got %q, want %q", got.Title, orig.Title)
	}
	if got.CategoryUser != orig.CategoryUser {
		t.Errorf("CategoryUser: got %q, want %q", got.CategoryUser, orig.CategoryUser)
	}
	if got.Status != orig.Status {
		t.Errorf("Status: got %q, want %q", got.Status, orig.Status)
	}
	if got.FaqArticleID == nil || *got.FaqArticleID != *orig.FaqArticleID {
		t.Errorf("FaqArticleID: got %v, want %v", got.FaqArticleID, orig.FaqArticleID)
	}
	if got.ResolvedAt == nil || !got.ResolvedAt.Equal(*orig.ResolvedAt) {
		t.Errorf("ResolvedAt: got %v, want %v", got.ResolvedAt, orig.ResolvedAt)
	}
	if string(got.AutoContext) != string(orig.AutoContext) {
		t.Errorf("AutoContext: got %s, want %s", got.AutoContext, orig.AutoContext)
	}
}

func TestObservationOmitEmpty(t *testing.T) {
	obs := Observation{
		ID:           1,
		Ref:          "OBS-0002",
		UserID:       10,
		DomainID:     20,
		Title:        "Question about filters",
		Description:  "How do I create a filter?",
		CategoryUser: CategoryQuestion,
		Status:       StatusSubmitted,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	data, err := json.Marshal(obs)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal raw: %v", err)
	}

	omitted := []string{"severity_user", "location", "auto_context", "triage",
		"resolution_internal", "resolution_user", "related_observations",
		"faq_article_id", "resolved_at", "user_email", "domain_name", "screenshot_count"}
	for _, key := range omitted {
		if _, ok := raw[key]; ok {
			t.Errorf("expected %q to be omitted, but it was present", key)
		}
	}
}

func TestFAQArticleJSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	triggered := now.Add(2 * time.Hour)

	orig := FAQArticle{
		ID:                 5,
		Ref:                "FAQ-0005",
		Title:              "How to set up email forwarding",
		Category:           "settings",
		Tags:               json.RawMessage(`["forwarding","setup"]`),
		Problem:            "Users don't know how to forward email.",
		Resolution:         "Go to Settings > Forwarding.",
		RelatedArticles:    json.RawMessage(`[1,2]`),
		ObservationCount:   12,
		DeflectionCount:    8,
		PublishedAt:        now,
		LastTriggeredAt:    &triggered,
		ImprovementFlagged: true,
		Views:              150,
		HelpfulYes:         40,
		HelpfulNo:          3,
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got FAQArticle
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.ID != orig.ID {
		t.Errorf("ID: got %d, want %d", got.ID, orig.ID)
	}
	if got.Title != orig.Title {
		t.Errorf("Title: got %q, want %q", got.Title, orig.Title)
	}
	if got.ImprovementFlagged != orig.ImprovementFlagged {
		t.Errorf("ImprovementFlagged: got %v, want %v", got.ImprovementFlagged, orig.ImprovementFlagged)
	}
	if string(got.Tags) != string(orig.Tags) {
		t.Errorf("Tags: got %s, want %s", got.Tags, orig.Tags)
	}
	if got.LastTriggeredAt == nil || !got.LastTriggeredAt.Equal(*orig.LastTriggeredAt) {
		t.Errorf("LastTriggeredAt: got %v, want %v", got.LastTriggeredAt, orig.LastTriggeredAt)
	}
}

func TestTimelineJSONRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)

	orig := Timeline{
		ID:            1,
		ObservationID: 10,
		EventType:     EventStatusChange,
		Actor:         ActorAdmin,
		Content:       "Changed status to triaged",
		CreatedAt:     now,
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got Timeline
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.ID != orig.ID {
		t.Errorf("ID: got %d, want %d", got.ID, orig.ID)
	}
	if got.ObservationID != orig.ObservationID {
		t.Errorf("ObservationID: got %d, want %d", got.ObservationID, orig.ObservationID)
	}
	if got.EventType != orig.EventType {
		t.Errorf("EventType: got %q, want %q", got.EventType, orig.EventType)
	}
	if got.Actor != orig.Actor {
		t.Errorf("Actor: got %q, want %q", got.Actor, orig.Actor)
	}
	if got.Content != orig.Content {
		t.Errorf("Content: got %q, want %q", got.Content, orig.Content)
	}
}

func TestConstants(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		// Categories
		{"CategorySomethingOff", CategorySomethingOff, "something_off"},
		{"CategoryNotWorking", CategoryNotWorking, "not_working"},
		{"CategoryQuestion", CategoryQuestion, "question"},
		{"CategoryIdea", CategoryIdea, "idea"},
		// Severities
		{"SeverityBlocking", SeverityBlocking, "blocking"},
		{"SeverityAnnoying", SeverityAnnoying, "annoying"},
		{"SeverityFYI", SeverityFYI, "fyi"},
		// Locations
		{"LocationWebmail", LocationWebmail, "webmail"},
		{"LocationCompose", LocationCompose, "compose"},
		{"LocationFolders", LocationFolders, "folders"},
		{"LocationSearch", LocationSearch, "search"},
		{"LocationCalendar", LocationCalendar, "calendar"},
		{"LocationContacts", LocationContacts, "contacts"},
		{"LocationSettings", LocationSettings, "settings"},
		{"LocationLogin", LocationLogin, "login"},
		{"LocationNotifications", LocationNotifications, "notifications"},
		{"LocationIMAP", LocationIMAP, "imap"},
		{"LocationDelivery", LocationDelivery, "delivery"},
		{"LocationOther", LocationOther, "other"},
		// Statuses
		{"StatusSubmitted", StatusSubmitted, "submitted"},
		{"StatusTriaged", StatusTriaged, "triaged"},
		{"StatusInvestigating", StatusInvestigating, "investigating"},
		{"StatusAwaitingInfo", StatusAwaitingInfo, "awaiting_info"},
		{"StatusResolved", StatusResolved, "resolved"},
		{"StatusPublished", StatusPublished, "published"},
		// Event types
		{"EventStatusChange", EventStatusChange, "status_change"},
		{"EventComment", EventComment, "comment"},
		{"EventInfoRequested", EventInfoRequested, "info_requested"},
		{"EventResolutionPosted", EventResolutionPosted, "resolution_posted"},
		{"EventFAQPublished", EventFAQPublished, "faq_published"},
		{"EventInvestigationNote", EventInvestigationNote, "investigation_note"},
		// Actors
		{"ActorUser", ActorUser, "user"},
		{"ActorAdmin", ActorAdmin, "admin"},
		{"ActorSystem", ActorSystem, "system"},
		// Notification events
		{"NotifySubmitted", NotifySubmitted, "submitted"},
		{"NotifyStatusChange", NotifyStatusChange, "status_change"},
		{"NotifyInfoRequested", NotifyInfoRequested, "info_requested"},
		{"NotifyResolved", NotifyResolved, "resolved"},
		{"NotifyFAQPublished", NotifyFAQPublished, "faq_published"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}
