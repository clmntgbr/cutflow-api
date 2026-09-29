package write

import (
	"context"
	"errors"
	"time"

	domaintranscript "go-api/internal/domain/transcript"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type transcriptWriteRepository struct {
	db *gorm.DB
}

func NewTranscriptWriteRepository(db *gorm.DB) domaintranscript.TranscriptWriteRepository {
	return &transcriptWriteRepository{db: db}
}

func (r *transcriptWriteRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ContextWithTx(ctx, tx))
	})
}

func (r *transcriptWriteRepository) Save(ctx context.Context, transcript *domaintranscript.Transcript) error {
	return DBWithContext(ctx, r.db).Create(transcriptModelFromDomain(transcript)).Error
}

func (r *transcriptWriteRepository) Update(ctx context.Context, transcript *domaintranscript.Transcript) error {
	return DBWithContext(ctx, r.db).Save(transcriptModelFromDomain(transcript)).Error
}

func (r *transcriptWriteRepository) ReplaceWords(
	ctx context.Context,
	transcriptID uuid.UUID,
	words []domaintranscript.Word,
) error {
	db := DBWithContext(ctx, r.db)
	if err := db.Where("transcript_id = ?", transcriptID).Delete(&TranscriptWordModel{}).Error; err != nil {
		return err
	}
	if len(words) == 0 {
		return nil
	}
	now := time.Now().UTC()
	models := make([]*TranscriptWordModel, 0, len(words))
	for _, word := range words {
		kind := word.Kind
		if kind == "" {
			kind = domaintranscript.KindSpeech
		}
		models = append(models, &TranscriptWordModel{
			TranscriptID:  transcriptID,
			WordIndex:     word.WordIndex,
			Text:          word.Text,
			SourceStartMs: word.SourceStartMs,
			SourceEndMs:   word.SourceEndMs,
			Confidence:    word.Confidence,
			Kind:          kind,
			CreatedAt:     now,
		})
	}
	return db.Create(&models).Error
}

func (r *transcriptWriteRepository) ListWords(
	ctx context.Context,
	transcriptID uuid.UUID,
) ([]domaintranscript.Word, error) {
	var models []TranscriptWordModel
	err := DBWithContext(ctx, r.db).
		Where("transcript_id = ?", transcriptID).
		Order("word_index ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	out := make([]domaintranscript.Word, 0, len(models))
	for _, m := range models {
		out = append(out, domaintranscript.Word{
			WordIndex:     m.WordIndex,
			Text:          m.Text,
			SourceStartMs: m.SourceStartMs,
			SourceEndMs:   m.SourceEndMs,
			Confidence:    m.Confidence,
			Kind:          m.Kind,
		})
	}
	return out, nil
}

func (r *transcriptWriteRepository) GetByMediaFileID(
	ctx context.Context,
	mediaFileID uuid.UUID,
) (*domaintranscript.Transcript, error) {
	db := DBWithContext(ctx, r.db)
	if _, ok := ctx.Value(txCtxKey{}).(*gorm.DB); ok {
		db = db.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var model TranscriptModel
	err := db.Where("media_file_id = ?", mediaFileID).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return transcriptDomainFromModel(&model), nil
}
