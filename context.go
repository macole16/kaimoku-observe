package observe

import "context"

type ContextData struct {
	UserID     int64                  `json:"user_id"`
	Email      string                 `json:"email"`
	DomainID   int64                  `json:"domain_id"`
	DomainName string                 `json:"domain_name"`
	Role       string                 `json:"role"`
	CapturedAt string                 `json:"captured_at"`
	Extra      map[string]interface{} `json:"extra,omitempty"`
}

type ContextProvider interface {
	CaptureContext(ctx context.Context, userID, domainID int64) (*ContextData, error)
}
