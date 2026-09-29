package read

import (
	"context"
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

	return &domainproject.ProjectDetailView{
		ID:         row.ID,
		Name:       row.Name,
		Status:     row.Status,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
		MediaFiles: mediaFiles,
	}, nil
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
		}
		if row.MediaFileID != nil {
			view.MediaFileID = *row.MediaFileID
		}
		views = append(views, view)
	}
	return views, total, nil
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
