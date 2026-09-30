package realtime

const (
	ActionCreated        = "created"
	ActionUpdated        = "updated"
	ActionDeleted        = "deleted"
	ActionUploaded       = "uploaded"
	ActionThumbnailReady = "thumbnail_ready"
	ActionTimelineUpdated = "timeline_updated"
	ActionTimelineFailed  = "timeline_failed"

	EntityUser      = "user"
	EntityMediaFile = "media_file"
	EntityProject   = "project"
	EntityJob       = "job"
)

func EventType(entity, action string) string {
	return entity + "." + action
}
