package realtime

const (
	ActionCreated  = "created"
	ActionUpdated  = "updated"
	ActionDeleted  = "deleted"
	ActionUploaded = "uploaded"

	EntityUser      = "user"
	EntityMediaFile = "media_file"
)


func EventType(entity, action string) string {
	return entity + "." + action
}
