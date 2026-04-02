package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	observe "github.com/macole16/kaimoku-observe"
)

func nullStr(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func nullInt64(v *int64) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *v, Valid: true}
}

func nullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func nullJSON(data json.RawMessage) []byte {
	if len(data) == 0 {
		return nil
	}
	return data
}

// CreateObservation inserts a new observation and populates its ID and Ref.
func (s *Store) CreateObservation(ctx context.Context, obs *observe.Observation) error {
	var seq int64
	if err := s.db.QueryRowContext(ctx, "SELECT nextval('obs_ref_seq')").Scan(&seq); err != nil {
		return fmt.Errorf("obs_ref_seq: %w", err)
	}
	obs.Ref = observe.FormatRef("OBS", seq)

	autoCtx := obs.AutoContext
	if len(autoCtx) == 0 {
		autoCtx = json.RawMessage("{}")
	}

	err := s.db.QueryRowContext(ctx, `
		INSERT INTO observations (
			ref, user_id, domain_id, title, description,
			category_user, severity_user, location, auto_context, status
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id, created_at, updated_at`,
		obs.Ref, obs.UserID, obs.DomainID, obs.Title, obs.Description,
		obs.CategoryUser, nullStr(obs.SeverityUser), nullStr(obs.Location),
		autoCtx, obs.Status,
	).Scan(&obs.ID, &obs.CreatedAt, &obs.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert observation: %w", err)
	}
	return nil
}

// GetObservation retrieves a single observation by ID.
func (s *Store) GetObservation(ctx context.Context, id int64) (*observe.Observation, error) {
	var (
		o            observe.Observation
		severityUser sql.NullString
		location     sql.NullString
		triage       []byte
		resInternal  sql.NullString
		resUser      sql.NullString
		relatedObs   []byte
		faqID        sql.NullInt64
		resolvedAt   sql.NullTime
	)

	err := s.db.QueryRowContext(ctx, `
		SELECT id, ref, user_id, domain_id, title, description,
			category_user, severity_user, location, auto_context,
			triage, status, resolution_internal, resolution_user,
			related_observations, faq_article_id,
			created_at, updated_at, resolved_at
		FROM observations WHERE id = $1`, id,
	).Scan(
		&o.ID, &o.Ref, &o.UserID, &o.DomainID, &o.Title, &o.Description,
		&o.CategoryUser, &severityUser, &location, &o.AutoContext,
		&triage, &o.Status, &resInternal, &resUser,
		&relatedObs, &faqID,
		&o.CreatedAt, &o.UpdatedAt, &resolvedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get observation %d: %w", id, err)
	}

	o.SeverityUser = severityUser.String
	o.Location = location.String
	o.Triage = triage
	o.ResolutionInternal = resInternal.String
	o.ResolutionUser = resUser.String
	o.RelatedObservations = relatedObs
	if faqID.Valid {
		o.FaqArticleID = &faqID.Int64
	}
	if resolvedAt.Valid {
		o.ResolvedAt = &resolvedAt.Time
	}

	return &o, nil
}

// GetObservationByRef retrieves an observation by its human-readable ref.
func (s *Store) GetObservationByRef(ctx context.Context, ref string) (*observe.Observation, error) {
	var id int64
	err := s.db.QueryRowContext(ctx, "SELECT id FROM observations WHERE ref = $1", ref).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("get observation by ref %s: %w", ref, err)
	}
	return s.GetObservation(ctx, id)
}

// ListObservations returns a filtered, paginated list of observations and the total count.
func (s *Store) ListObservations(ctx context.Context, f observe.ObservationFilter) ([]observe.Observation, int64, error) {
	var (
		where []string
		args  []any
		idx   int
	)

	addWhere := func(clause string, val any) {
		idx++
		where = append(where, fmt.Sprintf(clause, idx))
		args = append(args, val)
	}

	if f.UserID != 0 {
		addWhere("user_id = $%d", f.UserID)
	}
	if f.DomainID != 0 {
		addWhere("domain_id = $%d", f.DomainID)
	}
	if f.Status != "" {
		addWhere("status = $%d", f.Status)
	}

	whereSQL := ""
	if len(where) > 0 {
		whereSQL = " WHERE " + strings.Join(where, " AND ")
	}

	// Count query.
	var total int64
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM observations"+whereSQL, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count observations: %w", err)
	}

	if total == 0 {
		return nil, 0, nil
	}

	// Paginated fetch.
	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}

	idx++
	limitPlaceholder := fmt.Sprintf("$%d", idx)
	args = append(args, limit)
	idx++
	offsetPlaceholder := fmt.Sprintf("$%d", idx)
	args = append(args, offset)

	query := fmt.Sprintf(`
		SELECT id, ref, user_id, domain_id, title, description,
			category_user, severity_user, location, auto_context,
			triage, status, resolution_internal, resolution_user,
			related_observations, faq_article_id,
			created_at, updated_at, resolved_at
		FROM observations%s
		ORDER BY created_at DESC
		LIMIT %s OFFSET %s`, whereSQL, limitPlaceholder, offsetPlaceholder)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list observations: %w", err)
	}
	defer rows.Close()

	var result []observe.Observation
	for rows.Next() {
		var (
			o            observe.Observation
			severityUser sql.NullString
			location     sql.NullString
			triage       []byte
			resInternal  sql.NullString
			resUser      sql.NullString
			relatedObs   []byte
			faqID        sql.NullInt64
			resolvedAt   sql.NullTime
		)

		if err := rows.Scan(
			&o.ID, &o.Ref, &o.UserID, &o.DomainID, &o.Title, &o.Description,
			&o.CategoryUser, &severityUser, &location, &o.AutoContext,
			&triage, &o.Status, &resInternal, &resUser,
			&relatedObs, &faqID,
			&o.CreatedAt, &o.UpdatedAt, &resolvedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan observation: %w", err)
		}

		o.SeverityUser = severityUser.String
		o.Location = location.String
		o.Triage = triage
		o.ResolutionInternal = resInternal.String
		o.ResolutionUser = resUser.String
		o.RelatedObservations = relatedObs
		if faqID.Valid {
			o.FaqArticleID = &faqID.Int64
		}
		if resolvedAt.Valid {
			o.ResolvedAt = &resolvedAt.Time
		}

		result = append(result, o)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration: %w", err)
	}

	return result, total, nil
}

// UpdateObservationStatus updates the status of an observation.
func (s *Store) UpdateObservationStatus(ctx context.Context, id int64, status string) error {
	res, err := s.db.ExecContext(ctx,
		"UPDATE observations SET status = $1, updated_at = now() WHERE id = $2",
		status, id)
	if err != nil {
		return fmt.Errorf("update observation status: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// UpdateObservationTriage sets the triage data and moves status to triaged.
func (s *Store) UpdateObservationTriage(ctx context.Context, id int64, triage []byte) error {
	res, err := s.db.ExecContext(ctx, `
		UPDATE observations
		SET triage = $1, status = 'triaged', updated_at = now()
		WHERE id = $2`,
		triage, id)
	if err != nil {
		return fmt.Errorf("update observation triage: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ResolveObservation marks an observation as resolved with both user-facing and internal notes.
func (s *Store) ResolveObservation(ctx context.Context, id int64, resolutionUser, resolutionInternal string) error {
	res, err := s.db.ExecContext(ctx, `
		UPDATE observations
		SET status = 'resolved',
			resolution_user = $1,
			resolution_internal = $2,
			resolved_at = now(),
			updated_at = now()
		WHERE id = $3`,
		resolutionUser, resolutionInternal, id)
	if err != nil {
		return fmt.Errorf("resolve observation: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// LinkFAQ links an observation to a FAQ article and sets status to published.
func (s *Store) LinkFAQ(ctx context.Context, obsID int64, faqID int64) error {
	res, err := s.db.ExecContext(ctx, `
		UPDATE observations
		SET faq_article_id = $1, status = 'published', updated_at = now()
		WHERE id = $2`,
		faqID, obsID)
	if err != nil {
		return fmt.Errorf("link faq: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// AddTimeline appends a timeline event to an observation.
func (s *Store) AddTimeline(ctx context.Context, obsID int64, eventType, actor, content string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO observation_timeline (observation_id, event_type, actor, content)
		VALUES ($1, $2, $3, $4)`,
		obsID, eventType, actor, content)
	if err != nil {
		return fmt.Errorf("add timeline: %w", err)
	}
	return nil
}

// GetTimeline returns all timeline events for an observation ordered by time.
func (s *Store) GetTimeline(ctx context.Context, obsID int64) ([]observe.Timeline, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, observation_id, event_type, actor, content, created_at
		FROM observation_timeline
		WHERE observation_id = $1
		ORDER BY created_at`, obsID)
	if err != nil {
		return nil, fmt.Errorf("get timeline: %w", err)
	}
	defer rows.Close()

	var result []observe.Timeline
	for rows.Next() {
		var (
			t       observe.Timeline
			content sql.NullString
		)
		if err := rows.Scan(&t.ID, &t.ObservationID, &t.EventType, &t.Actor, &content, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan timeline: %w", err)
		}
		t.Content = content.String
		result = append(result, t)
	}
	return result, rows.Err()
}

// AddScreenshot records a screenshot attachment for an observation.
func (s *Store) AddScreenshot(ctx context.Context, sc *observe.Screenshot) error {
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO observation_screenshots (observation_id, filename, content_type, size_bytes, storage_key)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`,
		sc.ObservationID, sc.Filename, sc.ContentType, sc.SizeBytes, sc.StorageKey,
	).Scan(&sc.ID, &sc.CreatedAt)
	if err != nil {
		return fmt.Errorf("add screenshot: %w", err)
	}
	return nil
}

// GetScreenshots returns all screenshots for an observation.
func (s *Store) GetScreenshots(ctx context.Context, obsID int64) ([]observe.Screenshot, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, observation_id, filename, content_type, size_bytes, storage_key, created_at
		FROM observation_screenshots
		WHERE observation_id = $1
		ORDER BY created_at`, obsID)
	if err != nil {
		return nil, fmt.Errorf("get screenshots: %w", err)
	}
	defer rows.Close()

	var result []observe.Screenshot
	for rows.Next() {
		var sc observe.Screenshot
		if err := rows.Scan(&sc.ID, &sc.ObservationID, &sc.Filename, &sc.ContentType, &sc.SizeBytes, &sc.StorageKey, &sc.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan screenshot: %w", err)
		}
		result = append(result, sc)
	}
	return result, rows.Err()
}

// CreateFAQ inserts a new FAQ article and populates its ID and Ref.
func (s *Store) CreateFAQ(ctx context.Context, faq *observe.FAQArticle) error {
	var seq int64
	if err := s.db.QueryRowContext(ctx, "SELECT nextval('faq_ref_seq')").Scan(&seq); err != nil {
		return fmt.Errorf("faq_ref_seq: %w", err)
	}
	faq.Ref = observe.FormatRef("FAQ", seq)

	tags := faq.Tags
	if len(tags) == 0 {
		tags = json.RawMessage("[]")
	}

	err := s.db.QueryRowContext(ctx, `
		INSERT INTO faq_articles (ref, title, category, tags, problem, resolution, related_articles)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, published_at`,
		faq.Ref, faq.Title, faq.Category, tags, faq.Problem, faq.Resolution, nullJSON(faq.RelatedArticles),
	).Scan(&faq.ID, &faq.PublishedAt)
	if err != nil {
		return fmt.Errorf("insert faq: %w", err)
	}
	return nil
}

// GetFAQ retrieves a single FAQ article by ID.
func (s *Store) GetFAQ(ctx context.Context, id int64) (*observe.FAQArticle, error) {
	var (
		f               observe.FAQArticle
		relatedArticles []byte
		lastTriggeredAt sql.NullTime
	)

	err := s.db.QueryRowContext(ctx, `
		SELECT id, ref, title, category, tags, problem, resolution,
			related_articles, observation_count, deflection_count,
			published_at, last_triggered_at, improvement_flagged,
			views, helpful_yes, helpful_no
		FROM faq_articles WHERE id = $1`, id,
	).Scan(
		&f.ID, &f.Ref, &f.Title, &f.Category, &f.Tags, &f.Problem, &f.Resolution,
		&relatedArticles, &f.ObservationCount, &f.DeflectionCount,
		&f.PublishedAt, &lastTriggeredAt, &f.ImprovementFlagged,
		&f.Views, &f.HelpfulYes, &f.HelpfulNo,
	)
	if err != nil {
		return nil, fmt.Errorf("get faq %d: %w", id, err)
	}

	f.RelatedArticles = relatedArticles
	if lastTriggeredAt.Valid {
		f.LastTriggeredAt = &lastTriggeredAt.Time
	}

	return &f, nil
}

// ListFAQ returns FAQ articles filtered by category.
func (s *Store) ListFAQ(ctx context.Context, category string, limit int) ([]observe.FAQArticle, error) {
	if limit <= 0 {
		limit = 50
	}

	var (
		query string
		args  []any
	)

	if category != "" {
		query = `SELECT id, ref, title, category, tags, problem, resolution,
			related_articles, observation_count, deflection_count,
			published_at, last_triggered_at, improvement_flagged,
			views, helpful_yes, helpful_no
		FROM faq_articles WHERE category = $1
		ORDER BY published_at DESC LIMIT $2`
		args = []any{category, limit}
	} else {
		query = `SELECT id, ref, title, category, tags, problem, resolution,
			related_articles, observation_count, deflection_count,
			published_at, last_triggered_at, improvement_flagged,
			views, helpful_yes, helpful_no
		FROM faq_articles
		ORDER BY published_at DESC LIMIT $1`
		args = []any{limit}
	}

	return s.scanFAQRows(ctx, query, args...)
}

// SearchFAQ performs a simple text search across FAQ title, problem, and resolution.
func (s *Store) SearchFAQ(ctx context.Context, query string, limit int) ([]observe.FAQArticle, error) {
	if limit <= 0 {
		limit = 20
	}
	pattern := "%" + query + "%"
	return s.scanFAQRows(ctx, `
		SELECT id, ref, title, category, tags, problem, resolution,
			related_articles, observation_count, deflection_count,
			published_at, last_triggered_at, improvement_flagged,
			views, helpful_yes, helpful_no
		FROM faq_articles
		WHERE title ILIKE $1 OR problem ILIKE $1 OR resolution ILIKE $1
		ORDER BY published_at DESC LIMIT $2`, pattern, limit)
}

func (s *Store) scanFAQRows(ctx context.Context, query string, args ...any) ([]observe.FAQArticle, error) {
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query faq: %w", err)
	}
	defer rows.Close()

	var result []observe.FAQArticle
	for rows.Next() {
		var (
			f               observe.FAQArticle
			relatedArticles []byte
			lastTriggeredAt sql.NullTime
		)
		if err := rows.Scan(
			&f.ID, &f.Ref, &f.Title, &f.Category, &f.Tags, &f.Problem, &f.Resolution,
			&relatedArticles, &f.ObservationCount, &f.DeflectionCount,
			&f.PublishedAt, &lastTriggeredAt, &f.ImprovementFlagged,
			&f.Views, &f.HelpfulYes, &f.HelpfulNo,
		); err != nil {
			return nil, fmt.Errorf("scan faq: %w", err)
		}
		f.RelatedArticles = relatedArticles
		if lastTriggeredAt.Valid {
			f.LastTriggeredAt = &lastTriggeredAt.Time
		}
		result = append(result, f)
	}
	return result, rows.Err()
}

// IncrementFAQViews increments the view counter for a FAQ article.
func (s *Store) IncrementFAQViews(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, "UPDATE faq_articles SET views = views + 1 WHERE id = $1", id)
	return err
}

// RecordFAQFeedback records a helpful yes/no vote on a FAQ article.
func (s *Store) RecordFAQFeedback(ctx context.Context, id int64, helpful bool) error {
	var col string
	if helpful {
		col = "helpful_yes"
	} else {
		col = "helpful_no"
	}
	_, err := s.db.ExecContext(ctx,
		fmt.Sprintf("UPDATE faq_articles SET %s = %s + 1 WHERE id = $1", col, col), id)
	return err
}

// RecordFAQDeflection increments the deflection counter for a FAQ article.
func (s *Store) RecordFAQDeflection(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE faq_articles
		SET deflection_count = deflection_count + 1, last_triggered_at = now()
		WHERE id = $1`, id)
	return err
}

// CreateNotification inserts a new notification.
func (s *Store) CreateNotification(ctx context.Context, n *observe.Notification) error {
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO observation_notifications (user_id, observation_id, event_type, title, body)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`,
		n.UserID, nullInt64(n.ObservationID), n.EventType, n.Title, n.Body,
	).Scan(&n.ID, &n.CreatedAt)
	if err != nil {
		return fmt.Errorf("create notification: %w", err)
	}
	return nil
}

// ListNotifications returns recent notifications for a user with unread count.
func (s *Store) ListNotifications(ctx context.Context, userID int64, limit int) ([]observe.Notification, int, error) {
	if limit <= 0 {
		limit = 20
	}

	var unread int
	if err := s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM observation_notifications WHERE user_id = $1 AND NOT is_read",
		userID).Scan(&unread); err != nil {
		return nil, 0, fmt.Errorf("count unread: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, user_id, observation_id, event_type, title, body, is_read, created_at
		FROM observation_notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2`, userID, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	var result []observe.Notification
	for rows.Next() {
		var (
			n     observe.Notification
			obsID sql.NullInt64
			body  sql.NullString
		)
		if err := rows.Scan(&n.ID, &n.UserID, &obsID, &n.EventType, &n.Title, &body, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan notification: %w", err)
		}
		if obsID.Valid {
			n.ObservationID = &obsID.Int64
		}
		n.Body = body.String
		result = append(result, n)
	}
	return result, unread, rows.Err()
}

// MarkNotificationRead marks a single notification as read.
func (s *Store) MarkNotificationRead(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE observation_notifications SET is_read = true WHERE id = $1", id)
	return err
}

// UnreadCount returns the number of unread notifications for a user.
func (s *Store) UnreadCount(ctx context.Context, userID int64) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM observation_notifications WHERE user_id = $1 AND NOT is_read",
		userID).Scan(&count)
	return count, err
}

// DashboardStats returns aggregate observation statistics.
func (s *Store) DashboardStats(ctx context.Context) (*observe.DashboardStats, error) {
	stats := &observe.DashboardStats{
		ByStatus: make(map[string]int),
	}

	// Total count.
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM observations").Scan(&stats.Total); err != nil {
		return nil, fmt.Errorf("dashboard total: %w", err)
	}

	// By status.
	rows, err := s.db.QueryContext(ctx, "SELECT status, COUNT(*) FROM observations GROUP BY status")
	if err != nil {
		return nil, fmt.Errorf("dashboard by status: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("scan status count: %w", err)
		}
		stats.ByStatus[status] = count
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Recent (last 7 days).
	if err := s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM observations WHERE created_at > now() - INTERVAL '7 days'",
	).Scan(&stats.RecentCount); err != nil {
		return nil, fmt.Errorf("dashboard recent: %w", err)
	}

	// High severity.
	if err := s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM observations WHERE severity_user = 'blocking'",
	).Scan(&stats.HighSeverity); err != nil {
		return nil, fmt.Errorf("dashboard high severity: %w", err)
	}

	return stats, nil
}

// FrequencyReport returns observation frequency analysis over a time period.
func (s *Store) FrequencyReport(ctx context.Context, days int) (*observe.FrequencyReport, error) {
	report := &observe.FrequencyReport{
		Days:        days,
		Categories:  make(map[string]int),
		Resolutions: make(map[string]int),
	}

	interval := fmt.Sprintf("%d days", days)

	// Categories.
	rows, err := s.db.QueryContext(ctx, `
		SELECT category_user, COUNT(*)
		FROM observations
		WHERE created_at > now() - $1::INTERVAL
		GROUP BY category_user`, interval)
	if err != nil {
		return nil, fmt.Errorf("frequency categories: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var cat string
		var count int
		if err := rows.Scan(&cat, &count); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		report.Categories[cat] = count
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Resolutions by status.
	rows2, err := s.db.QueryContext(ctx, `
		SELECT status, COUNT(*)
		FROM observations
		WHERE created_at > now() - $1::INTERVAL
		GROUP BY status`, interval)
	if err != nil {
		return nil, fmt.Errorf("frequency resolutions: %w", err)
	}
	defer rows2.Close()
	for rows2.Next() {
		var status string
		var count int
		if err := rows2.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("scan resolution: %w", err)
		}
		report.Resolutions[status] = count
	}
	if err := rows2.Err(); err != nil {
		return nil, err
	}

	// Average resolution time.
	var avgHours sql.NullFloat64
	err = s.db.QueryRowContext(ctx, `
		SELECT AVG(EXTRACT(EPOCH FROM (resolved_at - created_at)) / 3600)
		FROM observations
		WHERE resolved_at IS NOT NULL
		AND created_at > now() - $1::INTERVAL`, interval).Scan(&avgHours)
	if err != nil {
		return nil, fmt.Errorf("frequency avg resolution: %w", err)
	}
	if avgHours.Valid {
		report.AvgResolutionHours = avgHours.Float64
	}

	return report, nil
}
