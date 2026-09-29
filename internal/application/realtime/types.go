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

	EntityUser      = "user"
	EntityMediaFile = "media_file"
	EntityProject   = "project"
)


func EventType(entity, action string) string {
	return entity + "." + action
}
