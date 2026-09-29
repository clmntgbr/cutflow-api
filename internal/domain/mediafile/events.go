package mediafile

import "time"

const (
	EventTypeMediaFileCreated        = "media_file.created.v1"
	EventTypeMediaFileUploaded       = "media_file.uploaded.v1"
	EventTypeMediaFileUploadExpired  = "media_file.upload_expired.v1"
	EventTypeMediaFileThumbnailReady = "media_file.thumbnail_ready.v1"
	EventTypeMediaFileProbing        = "media_file.probing.v1"
	EventTypeMediaFileReady          = "media_file.ready.v1"
	EventTypeMediaFileProbeFailed         = "media_file.probe_failed.v1"
	EventTypeMediaFileSilenceRequested    = "media_file.silence_requested.v1"
	EventTypeMediaFileTranscriptRequested = "media_file.transcript_requested.v1"
	EventTypeMediaFileAnalysisRequested   = "media_file.analysis_requested.v1"
	EventTypeMediaFileViralRequested      = "media_file.viral_requested.v1"
)

type MediaFileCreated struct {
	ID          string    `json:"eventId"`
	MediaFileID string    `json:"mediaFileId"`
	ProjectID   string    `json:"projectId"`
	UserID      string    `json:"userId"`
	Filename    string    `json:"filename"`
	StorageKey  string    `json:"storageKey"`
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e MediaFileCreated) EventID() string       { return e.ID }
func (e MediaFileCreated) EventType() string     { return EventTypeMediaFileCreated }
func (e MediaFileCreated) AggregateID() string   { return e.MediaFileID }
func (e MediaFileCreated) OccurredAt() time.Time { return e.Timestamp }

type MediaFileUploaded struct {
	ID          string    `json:"eventId"`
	MediaFileID string    `json:"mediaFileId"`
	ProjectID   string    `json:"projectId"`
	UserID      string    `json:"userId"`
	StorageKey  string    `json:"storageKey"`
	ContentType string    `json:"contentType"`
	SizeBytes   int64     `json:"sizeBytes"`
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e MediaFileUploaded) EventID() string       { return e.ID }
func (e MediaFileUploaded) EventType() string     { return EventTypeMediaFileUploaded }
func (e MediaFileUploaded) AggregateID() string   { return e.MediaFileID }
func (e MediaFileUploaded) OccurredAt() time.Time { return e.Timestamp }

type MediaFileUploadExpired struct {
	ID          string    `json:"eventId"`
	MediaFileID string    `json:"mediaFileId"`
	ProjectID   string    `json:"projectId"`
	UserID      string    `json:"userId"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e MediaFileUploadExpired) EventID() string       { return e.ID }
func (e MediaFileUploadExpired) EventType() string     { return EventTypeMediaFileUploadExpired }
func (e MediaFileUploadExpired) AggregateID() string   { return e.MediaFileID }
func (e MediaFileUploadExpired) OccurredAt() time.Time { return e.Timestamp }

type MediaFileThumbnailReady struct {
	ID           string    `json:"eventId"`
	MediaFileID  string    `json:"mediaFileId"`
	ProjectID    string    `json:"projectId"`
	UserID       string    `json:"userId"`
	ThumbnailKey string    `json:"thumbnailKey"`
	Status       string    `json:"status"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e MediaFileThumbnailReady) EventID() string       { return e.ID }
func (e MediaFileThumbnailReady) EventType() string     { return EventTypeMediaFileThumbnailReady }
func (e MediaFileThumbnailReady) AggregateID() string   { return e.MediaFileID }
func (e MediaFileThumbnailReady) OccurredAt() time.Time { return e.Timestamp }

type MediaFileProbing struct {
	ID          string    `json:"eventId"`
	MediaFileID string    `json:"mediaFileId"`
	ProjectID   string    `json:"projectId"`
	UserID      string    `json:"userId"`
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e MediaFileProbing) EventID() string       { return e.ID }
func (e MediaFileProbing) EventType() string     { return EventTypeMediaFileProbing }
func (e MediaFileProbing) AggregateID() string   { return e.MediaFileID }
func (e MediaFileProbing) OccurredAt() time.Time { return e.Timestamp }

type MediaFileReady struct {
	ID              string    `json:"eventId"`
	MediaFileID     string    `json:"mediaFileId"`
	ProjectID       string    `json:"projectId"`
	UserID          string    `json:"userId"`
	Status          string    `json:"status"`
	DurationMs      int64     `json:"durationMs"`
	Width           *int      `json:"width,omitempty"`
	Height          *int      `json:"height,omitempty"`
	FPS             *float64  `json:"fps,omitempty"`
	VideoCodec      string    `json:"videoCodec,omitempty"`
	AudioCodec      string    `json:"audioCodec,omitempty"`
	AudioSampleRate *int      `json:"audioSampleRate,omitempty"`
	AudioChannels   *int      `json:"audioChannels,omitempty"`
	SizeBytes       int64     `json:"sizeBytes"`
	Timestamp       time.Time `json:"timestamp"`
}

func (e MediaFileReady) EventID() string       { return e.ID }
func (e MediaFileReady) EventType() string     { return EventTypeMediaFileReady }
func (e MediaFileReady) AggregateID() string   { return e.MediaFileID }
func (e MediaFileReady) OccurredAt() time.Time { return e.Timestamp }

type MediaFileProbeFailed struct {
	ID          string    `json:"eventId"`
	MediaFileID string    `json:"mediaFileId"`
	ProjectID   string    `json:"projectId"`
	UserID      string    `json:"userId"`
	Status      string    `json:"status"`
	Reason      string    `json:"reason"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e MediaFileProbeFailed) EventID() string       { return e.ID }
func (e MediaFileProbeFailed) EventType() string     { return EventTypeMediaFileProbeFailed }
func (e MediaFileProbeFailed) AggregateID() string   { return e.MediaFileID }
func (e MediaFileProbeFailed) OccurredAt() time.Time { return e.Timestamp }

type MediaFileSilenceRequested struct {
	ID          string    `json:"eventId"`
	MediaFileID string    `json:"mediaFileId"`
	ProjectID   string    `json:"projectId"`
	UserID      string    `json:"userId"`
	JobID       string    `json:"jobId"`
	AudioKey    string    `json:"audioKey"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e MediaFileSilenceRequested) EventID() string       { return e.ID }
func (e MediaFileSilenceRequested) EventType() string     { return EventTypeMediaFileSilenceRequested }
func (e MediaFileSilenceRequested) AggregateID() string   { return e.MediaFileID }
func (e MediaFileSilenceRequested) OccurredAt() time.Time { return e.Timestamp }

type MediaFileTranscriptRequested struct {
	ID          string    `json:"eventId"`
	MediaFileID string    `json:"mediaFileId"`
	ProjectID   string    `json:"projectId"`
	UserID      string    `json:"userId"`
	JobID       string    `json:"jobId"`
	AudioKey    string    `json:"audioKey"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e MediaFileTranscriptRequested) EventID() string       { return e.ID }
func (e MediaFileTranscriptRequested) EventType() string     { return EventTypeMediaFileTranscriptRequested }
func (e MediaFileTranscriptRequested) AggregateID() string   { return e.MediaFileID }
func (e MediaFileTranscriptRequested) OccurredAt() time.Time { return e.Timestamp }

// MediaFileAnalysisRequested is emitted after transcription completes.
// The analysis worker only needs TranscriptWord[]; it does not use audio/silence.
type MediaFileAnalysisRequested struct {
	ID           string    `json:"eventId"`
	MediaFileID  string    `json:"mediaFileId"`
	ProjectID    string    `json:"projectId"`
	UserID       string    `json:"userId"`
	JobID        string    `json:"jobId"`
	TranscriptID string    `json:"transcriptId"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e MediaFileAnalysisRequested) EventID() string       { return e.ID }
func (e MediaFileAnalysisRequested) EventType() string     { return EventTypeMediaFileAnalysisRequested }
func (e MediaFileAnalysisRequested) AggregateID() string   { return e.MediaFileID }
func (e MediaFileAnalysisRequested) OccurredAt() time.Time { return e.Timestamp }

// MediaFileViralRequested is emitted after transcription completes (parallel with text analysis).
type MediaFileViralRequested struct {
	ID           string    `json:"eventId"`
	MediaFileID  string    `json:"mediaFileId"`
	ProjectID    string    `json:"projectId"`
	UserID       string    `json:"userId"`
	JobID        string    `json:"jobId"`
	TranscriptID string    `json:"transcriptId"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e MediaFileViralRequested) EventID() string       { return e.ID }
func (e MediaFileViralRequested) EventType() string     { return EventTypeMediaFileViralRequested }
func (e MediaFileViralRequested) AggregateID() string   { return e.MediaFileID }
func (e MediaFileViralRequested) OccurredAt() time.Time { return e.Timestamp }
