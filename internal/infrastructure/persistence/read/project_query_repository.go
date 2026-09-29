package read

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"go-api/internal/domain/paginate"
	domainproject "go-api/internal/domain/project"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type projectListRow struct {
	ID           uuid.UUID
	Name         string
	Status       string
	CreatedAt    time.Time
	MediaFileID  *uuid.UUID `gorm:"column:media_file_id"`
	ThumbnailKey string     `gorm:"column:thumbnail_key"`
}

func (projectListRow) TableName() string { return "project" }

type projectDetailRow struct {
	ID        uuid.UUID
	Name      string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (projectDetailRow) TableName() string { return "project" }

type projectMediaFileRow struct {
	ID               uuid.UUID
	OriginalFilename *string
	MimeType         *string
	SizeBytes        *int64
	DurationMs       int64
	Width            *int
	Height           *int
	FPS              *float64
	VideoCodec       *string
	AudioCodec       *string
	AudioSampleRate  *int
	AudioChannels    *int
	StorageKey       string
	ThumbnailKey     string
	Status           string
	CreatedAt        time.Time
}

func (projectMediaFileRow) TableName() string { return "media_file" }

type projectJobRow struct {
	ID           uuid.UUID
	ProjectID    uuid.UUID  `gorm:"column:project_id"`
	MediaFileID  uuid.UUID  `gorm:"column:media_file_id"`
	Name         string
	Status       string
	ErrorMessage *string    `gorm:"column:error_message"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	StartedAt    *time.Time
	CompletedAt  *time.Time
}

func (projectJobRow) TableName() string { return "job" }

type projectTimelineRow struct {
	ID            uuid.UUID
	MediaFileID   uuid.UUID `gorm:"column:media_file_id"`
	Version       int
	DurationMs    int64     `gorm:"column:duration_ms"`
	Fingerprint   string
	EngineVersion string    `gorm:"column:engine_version"`
	IsActive      bool      `gorm:"column:is_active"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (projectTimelineRow) TableName() string { return "timeline" }

type projectTimelineSegmentRow struct {
	TimelineID    uuid.UUID `gorm:"column:timeline_id"`
	MediaFileID   uuid.UUID `gorm:"column:media_file_id"`
	SegmentIndex  int       `gorm:"column:segment_index"`
	SourceStartMs int64     `gorm:"column:source_start_ms"`
	SourceEndMs   int64     `gorm:"column:source_end_ms"`
	OutputStartMs int64     `gorm:"column:output_start_ms"`
	OutputEndMs   int64     `gorm:"column:output_end_ms"`
}

func (projectTimelineSegmentRow) TableName() string { return "timeline_segment" }

type projectEditDecisionRow struct {
	ID            uuid.UUID
	TimelineID    *uuid.UUID      `gorm:"column:timeline_id"`
	MediaFileID   uuid.UUID       `gorm:"column:media_file_id"`
	DecisionType  string          `gorm:"column:decision_type"`
	SourceStartMs int64           `gorm:"column:source_start_ms"`
	SourceEndMs   int64           `gorm:"column:source_end_ms"`
	Action        string
	Source        string
	Confidence    *float64
	Reasons       json.RawMessage `gorm:"column:reasons;type:jsonb"`
}

func (projectEditDecisionRow) TableName() string { return "edit_decision" }

var projectListSortColumns = map[string]string{
	"created_at": "project.created_at",
	"updated_at": "project.updated_at",
	"name":       "project.name",
	"status":     "project.status",
}

type projectReadRepository struct {
	db *gorm.DB
}

func NewProjectReadRepository(db *gorm.DB) domainproject.ProjectReadRepository {
	return &projectReadRepository{db: db}
}

func (r *projectReadRepository) FindByID(ctx context.Context, id, userID uuid.UUID) (*domainproject.ProjectDetailView, error) {
	var row projectDetailRow
	err := r.db.WithContext(ctx).
		Select("id", "name", "status", "created_at", "updated_at").
		Where("id = ? AND user_id = ?", id, userID).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	var mediaRows []projectMediaFileRow
	err = r.db.WithContext(ctx).
		Select(
			"id",
			"original_filename",
			"mime_type",
			"size_bytes",
			"duration_ms",
			"width",
			"height",
			"fps",
			"video_codec",
			"audio_codec",
			"audio_sample_rate",
			"audio_channels",
			"storage_key",
			"thumbnail_key",
			"status",
			"created_at",
		).
		Where("project_id = ?", id).
		Order("created_at ASC").
		Find(&mediaRows).Error
	if err != nil {
		return nil, err
	}

	mediaFiles := make([]domainproject.ProjectMediaFileView, 0, len(mediaRows))
	for _, media := range mediaRows {
		view := domainproject.ProjectMediaFileView{
			ID:              media.ID,
			DurationMs:      media.DurationMs,
			Width:           media.Width,
			Height:          media.Height,
			FPS:             media.FPS,
			AudioSampleRate: media.AudioSampleRate,
			AudioChannels:   media.AudioChannels,
			StorageKey:      media.StorageKey,
			ThumbnailKey:    media.ThumbnailKey,
			Status:          media.Status,
			CreatedAt:       media.CreatedAt,
		}
		if media.OriginalFilename != nil {
			view.OriginalFilename = *media.OriginalFilename
		}
		if media.MimeType != nil {
			view.MimeType = *media.MimeType
		}
		if media.SizeBytes != nil {
			view.SizeBytes = *media.SizeBytes
		}
		if media.VideoCodec != nil {
			view.VideoCodec = *media.VideoCodec
		}
		if media.AudioCodec != nil {
			view.AudioCodec = *media.AudioCodec
		}
		mediaFiles = append(mediaFiles, view)
	}

	var jobRows []projectJobRow
	err = r.db.WithContext(ctx).
		Select(
			"id",
			"project_id",
			"media_file_id",
			"name",
			"status",
			"error_message",
			"created_at",
			"updated_at",
			"started_at",
			"completed_at",
		).
		Where("project_id = ?", id).
		Order("created_at ASC").
		Find(&jobRows).Error
	if err != nil {
		return nil, err
	}

	jobs := make([]domainproject.ProjectJobView, 0, len(jobRows))
	for _, job := range jobRows {
		jobs = append(jobs, projectJobViewFromRow(job))
	}

	timelines, err := r.loadActiveTimelines(ctx, id)
	if err != nil {
		return nil, err
	}

	return &domainproject.ProjectDetailView{
		ID:         row.ID,
		Name:       row.Name,
		Status:     row.Status,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
		MediaFiles: mediaFiles,
		Jobs:       jobs,
		Timelines:  timelines,
	}, nil
}

func (r *projectReadRepository) loadActiveTimelines(
	ctx context.Context,
	projectID uuid.UUID,
) ([]domainproject.ProjectTimelineView, error) {
	var timelineRows []projectTimelineRow
	err := r.db.WithContext(ctx).
		Select(
			"id",
			"media_file_id",
			"version",
			"duration_ms",
			"fingerprint",
			"engine_version",
			"is_active",
			"created_at",
			"updated_at",
		).
		Where("project_id = ? AND is_active = TRUE", projectID).
		Order("created_at ASC").
		Find(&timelineRows).Error
	if err != nil {
		return nil, err
	}
	if len(timelineRows) == 0 {
		return []domainproject.ProjectTimelineView{}, nil
	}

	timelineIDs := make([]uuid.UUID, len(timelineRows))
	indexByID := make(map[uuid.UUID]int, len(timelineRows))
	timelines := make([]domainproject.ProjectTimelineView, 0, len(timelineRows))
	for i, row := range timelineRows {
		timelineIDs[i] = row.ID
		indexByID[row.ID] = i
		timelines = append(timelines, domainproject.ProjectTimelineView{
			ID:            row.ID,
			MediaFileID:   row.MediaFileID,
			Version:       row.Version,
			DurationMs:    row.DurationMs,
			Fingerprint:   row.Fingerprint,
			EngineVersion: row.EngineVersion,
			IsActive:      row.IsActive,
			Segments:      []domainproject.ProjectTimelineSegmentView{},
			Decisions:     []domainproject.ProjectEditDecisionView{},
			CreatedAt:     row.CreatedAt,
			UpdatedAt:     row.UpdatedAt,
		})
	}

	var segmentRows []projectTimelineSegmentRow
	err = r.db.WithContext(ctx).
		Select(
			"timeline_id",
			"media_file_id",
			"segment_index",
			"source_start_ms",
			"source_end_ms",
			"output_start_ms",
			"output_end_ms",
		).
		Where("timeline_id IN ?", timelineIDs).
		Order("timeline_id ASC, segment_index ASC").
		Find(&segmentRows).Error
	if err != nil {
		return nil, err
	}
	for _, seg := range segmentRows {
		idx, ok := indexByID[seg.TimelineID]
		if !ok {
			continue
		}
		timelines[idx].Segments = append(timelines[idx].Segments, domainproject.ProjectTimelineSegmentView{
			Index:         seg.SegmentIndex,
			MediaFileID:   seg.MediaFileID,
			SourceStartMs: seg.SourceStartMs,
			SourceEndMs:   seg.SourceEndMs,
			OutputStartMs: seg.OutputStartMs,
			OutputEndMs:   seg.OutputEndMs,
		})
	}

	var decisionRows []projectEditDecisionRow
	err = r.db.WithContext(ctx).
		Select(
			"id",
			"timeline_id",
			"media_file_id",
			"decision_type",
			"source_start_ms",
			"source_end_ms",
			"action",
			"source",
			"confidence",
			"reasons",
		).
		Where("timeline_id IN ?", timelineIDs).
		Order("timeline_id ASC, source_start_ms ASC").
		Find(&decisionRows).Error
	if err != nil {
		return nil, err
	}
	for _, dec := range decisionRows {
		if dec.TimelineID == nil {
			continue
		}
		idx, ok := indexByID[*dec.TimelineID]
		if !ok {
			continue
		}
		reasons := make([]string, 0)
		if len(dec.Reasons) > 0 {
			_ = json.Unmarshal(dec.Reasons, &reasons)
		}
		timelines[idx].Decisions = append(timelines[idx].Decisions, domainproject.ProjectEditDecisionView{
			ID:            dec.ID,
			MediaFileID:   dec.MediaFileID,
			Type:          dec.DecisionType,
			SourceStartMs: dec.SourceStartMs,
			SourceEndMs:   dec.SourceEndMs,
			Action:        dec.Action,
			Source:        dec.Source,
			Confidence:    dec.Confidence,
			Reasons:       reasons,
		})
	}

	return timelines, nil
}

func (r *projectReadRepository) List(
	ctx context.Context,
	userID uuid.UUID,
	query paginate.PaginateQuery,
) ([]domainproject.ProjectListView, int64, error) {
	query.SortBy = normalizeProjectListSort(query.SortBy)

	db := r.db.WithContext(ctx).Table("project").
		Select(
			"project.id",
			"project.name",
			"project.status",
			"project.created_at",
			"thumb.media_file_id",
			"thumb.thumbnail_key",
		).
		Joins(`LEFT JOIN LATERAL (
			SELECT id AS media_file_id, thumbnail_key
			FROM media_file
			WHERE media_file.project_id = project.id
			  AND media_file.thumbnail_key <> ''
			ORDER BY media_file.created_at ASC
			LIMIT 1
		) AS thumb ON TRUE`).
		Where("project.user_id = ?", userID)
	if search := strings.TrimSpace(query.Search); search != "" {
		db = db.Where("project.name ILIKE ? ESCAPE '\\'", "%"+escapeLike(search)+"%")
	}

	scoped, total, err := Paginate(db, query)
	if err != nil {
		return nil, 0, err
	}

	var rows []projectListRow
	if err := scoped.Find(&rows).Error; err != nil {
		return nil, 0, err
	}

	views := make([]domainproject.ProjectListView, 0, len(rows))
	for _, row := range rows {
		view := domainproject.ProjectListView{
			ID:           row.ID,
			Name:         row.Name,
			Status:       row.Status,
			CreatedAt:    row.CreatedAt,
			ThumbnailKey: row.ThumbnailKey,
			Jobs:         []domainproject.ProjectJobView{},
		}
		if row.MediaFileID != nil {
			view.MediaFileID = *row.MediaFileID
		}
		views = append(views, view)
	}

	if err := r.attachJobs(ctx, views); err != nil {
		return nil, 0, err
	}
	return views, total, nil
}

func (r *projectReadRepository) attachJobs(ctx context.Context, views []domainproject.ProjectListView) error {
	if len(views) == 0 {
		return nil
	}

	projectIDs := make([]uuid.UUID, len(views))
	indexByID := make(map[uuid.UUID]int, len(views))
	for i, view := range views {
		projectIDs[i] = view.ID
		indexByID[view.ID] = i
	}

	var jobRows []projectJobRow
	err := r.db.WithContext(ctx).
		Select(
			"id",
			"project_id",
			"media_file_id",
			"name",
			"status",
			"error_message",
			"created_at",
			"updated_at",
			"started_at",
			"completed_at",
		).
		Where("project_id IN ?", projectIDs).
		Order("created_at ASC").
		Find(&jobRows).Error
	if err != nil {
		return err
	}

	for _, job := range jobRows {
		idx, ok := indexByID[job.ProjectID]
		if !ok {
			continue
		}
		views[idx].Jobs = append(views[idx].Jobs, projectJobViewFromRow(job))
	}
	return nil
}

func projectJobViewFromRow(job projectJobRow) domainproject.ProjectJobView {
	view := domainproject.ProjectJobView{
		ID:          job.ID,
		MediaFileID: job.MediaFileID,
		Name:        job.Name,
		Status:      job.Status,
		CreatedAt:   job.CreatedAt,
		UpdatedAt:   job.UpdatedAt,
		StartedAt:   job.StartedAt,
		CompletedAt: job.CompletedAt,
	}
	if job.ErrorMessage != nil {
		view.ErrorMessage = *job.ErrorMessage
	}
	return view
}

func normalizeProjectListSort(sortBy string) string {
	if column, ok := projectListSortColumns[sortBy]; ok {
		return column
	}
	return "project.created_at"
}

func escapeLike(value string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(value)
}
