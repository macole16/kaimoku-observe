package observe

import (
	"encoding/json"
	"time"
)

// Category constants (user-provided).
const (
	CategorySomethingOff = "something_off"
	CategoryNotWorking   = "not_working"
	CategoryQuestion     = "question"
	CategoryIdea         = "idea"
)

// Severity constants (user-provided, optional).
const (
	SeverityBlocking = "blocking"
	SeverityAnnoying = "annoying"
	SeverityFYI      = "fyi"
)

// Location constants.
const (
	LocationWebmail       = "webmail"
	LocationCompose       = "compose"
	LocationFolders       = "folders"
	LocationSearch        = "search"
	LocationCalendar      = "calendar"
	LocationContacts      = "contacts"
	LocationSettings      = "settings"
	LocationLogin         = "login"
	LocationNotifications = "notifications"
	LocationIMAP          = "imap"
	LocationDelivery      = "delivery"
	LocationOther         = "other"
)

// Status constants (observation lifecycle).
const (
	StatusSubmitted     = "submitted"
	StatusTriaged       = "triaged"
	StatusInvestigating = "investigating"
	StatusAwaitingInfo  = "awaiting_info"
	StatusResolved      = "resolved"
	StatusPublished     = "published"
)

// Timeline event type constants.
const (
	EventStatusChange      = "status_change"
	EventComment           = "comment"
	EventInfoRequested     = "info_requested"
	EventResolutionPosted  = "resolution_posted"
	EventFAQPublished      = "faq_published"
	EventInvestigationNote = "investigation_note"
)

// Actor constants.
const (
	ActorUser   = "user"
	ActorAdmin  = "admin"
	ActorSystem = "system"
)

// Notification event constants.
const (
	NotifySubmitted     = "submitted"
	NotifyStatusChange  = "status_change"
	NotifyInfoRequested = "info_requested"
	NotifyResolved      = "resolved"
	NotifyFAQPublished  = "faq_published"
)

// Observation represents a user-submitted observation about the product.
type Observation struct {
	ID                  int64           `json:"id"`
	Ref                 string          `json:"ref"`
	UserID              int64           `json:"user_id"`
	DomainID            int64           `json:"domain_id"`
	Title               string          `json:"title"`
	Description         string          `json:"description"`
	CategoryUser        string          `json:"category_user"`
	SeverityUser        string          `json:"severity_user,omitempty"`
	Location            string          `json:"location,omitempty"`
	AutoContext         json.RawMessage `json:"auto_context,omitempty"`
	Triage              json.RawMessage `json:"triage,omitempty"`
	Status              string          `json:"status"`
	ResolutionInternal  string          `json:"resolution_internal,omitempty"`
	ResolutionUser      string          `json:"resolution_user,omitempty"`
	RelatedObservations json.RawMessage `json:"related_observations,omitempty"`
	FaqArticleID        *int64          `json:"faq_article_id,omitempty"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
	ResolvedAt          *time.Time      `json:"resolved_at,omitempty"`
	UserEmail           string          `json:"user_email,omitempty"`
	DomainName          string          `json:"domain_name,omitempty"`
	ScreenshotCount     int             `json:"screenshot_count,omitempty"`
}

// Timeline represents an event in the lifecycle of an observation.
type Timeline struct {
	ID            int64     `json:"id"`
	ObservationID int64     `json:"observation_id"`
	EventType     string    `json:"event_type"`
	Actor         string    `json:"actor"`
	Content       string    `json:"content,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// FAQArticle represents a published FAQ article derived from observations.
type FAQArticle struct {
	ID                 int64           `json:"id"`
	Ref                string          `json:"ref"`
	Title              string          `json:"title"`
	Category           string          `json:"category"`
	Tags               json.RawMessage `json:"tags"`
	Problem            string          `json:"problem"`
	Resolution         string          `json:"resolution"`
	RelatedArticles    json.RawMessage `json:"related_articles,omitempty"`
	ObservationCount   int             `json:"observation_count"`
	DeflectionCount    int             `json:"deflection_count"`
	PublishedAt        time.Time       `json:"published_at"`
	LastTriggeredAt    *time.Time      `json:"last_triggered_at,omitempty"`
	ImprovementFlagged bool            `json:"improvement_flagged"`
	Views              int             `json:"views"`
	HelpfulYes         int             `json:"helpful_yes"`
	HelpfulNo          int             `json:"helpful_no"`
}

// Notification represents a notification sent to a user.
type Notification struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"user_id"`
	ObservationID *int64    `json:"observation_id,omitempty"`
	EventType     string    `json:"event_type"`
	Title         string    `json:"title"`
	Body          string    `json:"body,omitempty"`
	IsRead        bool      `json:"is_read"`
	CreatedAt     time.Time `json:"created_at"`
}

// Screenshot represents a screenshot attached to an observation.
type Screenshot struct {
	ID            int64     `json:"id"`
	ObservationID int64     `json:"observation_id"`
	Filename      string    `json:"filename"`
	ContentType   string    `json:"content_type"`
	SizeBytes     int64     `json:"size_bytes"`
	StorageKey    string    `json:"storage_key"`
	CreatedAt     time.Time `json:"created_at"`
}
