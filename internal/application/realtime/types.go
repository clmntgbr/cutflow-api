package realtime

const (
	ActionCreated        = "created"
	ActionUpdated        = "updated"
	ActionDeleted        = "deleted"
	ActionUploaded       = "uploaded"
	ActionThumbnailReady = "thumbnail_ready"
	ActionProbing        = "probing"
	ActionReady          = "ready"
	ActionFailed         = "failed"
	ActionAudioReady       = "audio_ready"
	ActionSilenceDetected  = "silence_detected"
	ActionTranscriptReady  = "transcript_ready"

	EntityUser      = "user"
	EntityMediaFile = "media_file"
	EntityProject   = "project"
	EntityJob       = "job"
)


func EventType(entity, action string) string {
	return entity + "." + action
}
