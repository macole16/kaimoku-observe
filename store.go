package observe

import "context"

// Store is the main persistence contract for the observe module.
type Store interface {
	// Observations
	CreateObservation(ctx context.Context, obs *Observation) error
	GetObservation(ctx context.Context, id int64) (*Observation, error)
	GetObservationByRef(ctx context.Context, ref string) (*Observation, error)
	ListObservations(ctx context.Context, filter ObservationFilter) ([]Observation, int64, error)
	UpdateObservationStatus(ctx context.Context, id int64, status string) error
	UpdateObservationTriage(ctx context.Context, id int64, triage []byte) error
	ResolveObservation(ctx context.Context, id int64, resolutionUser, resolutionInternal string) error
	LinkFAQ(ctx context.Context, obsID int64, faqID int64) error

	// Timeline
	AddTimeline(ctx context.Context, obsID int64, eventType, actor, content string) error
	GetTimeline(ctx context.Context, obsID int64) ([]Timeline, error)

	// Screenshots
	AddScreenshot(ctx context.Context, s *Screenshot) error
	GetScreenshots(ctx context.Context, obsID int64) ([]Screenshot, error)

	// FAQ
	CreateFAQ(ctx context.Context, faq *FAQArticle) error
	GetFAQ(ctx context.Context, id int64) (*FAQArticle, error)
	ListFAQ(ctx context.Context, category string, limit int) ([]FAQArticle, error)
	SearchFAQ(ctx context.Context, query string, limit int) ([]FAQArticle, error)
	IncrementFAQViews(ctx context.Context, id int64) error
	RecordFAQFeedback(ctx context.Context, id int64, helpful bool) error
	RecordFAQDeflection(ctx context.Context, id int64) error

	// Notifications
	CreateNotification(ctx context.Context, n *Notification) error
	ListNotifications(ctx context.Context, userID int64, limit int) ([]Notification, int, error)
	MarkNotificationRead(ctx context.Context, id int64) error
	UnreadCount(ctx context.Context, userID int64) (int, error)

	// Admin
	DashboardStats(ctx context.Context) (*DashboardStats, error)
	FrequencyReport(ctx context.Context, days int) (*FrequencyReport, error)
}

// ObservationFilter controls listing and pagination of observations.
type ObservationFilter struct {
	UserID   int64
	DomainID int64
	Status   string
	Limit    int
	Offset   int
}

// DashboardStats holds aggregate statistics for the admin dashboard.
type DashboardStats struct {
	Total        int            `json:"total"`
	ByStatus     map[string]int `json:"by_status"`
	RecentCount  int            `json:"recent_count"`
	HighSeverity int            `json:"high_severity"`
}

// FrequencyReport holds observation frequency analysis over a time period.
type FrequencyReport struct {
	Days               int            `json:"days"`
	Categories         map[string]int `json:"categories"`
	Resolutions        map[string]int `json:"resolutions"`
	AvgResolutionHours float64        `json:"avg_resolution_hours"`
}
