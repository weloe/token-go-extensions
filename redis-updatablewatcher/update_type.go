package redis_updatablewatcher

type UpdateType string

const (
	UpdateForSetStr    UpdateType = "UpdateForSetStr"
	UpdateForUpdateStr UpdateType = "UpdateForUpdateStr"

	UpdateForSetSession    UpdateType = "UpdateForSetSession"
	UpdateForUpdateSession UpdateType = "UpdateForUpdateSession"

	UpdateForSetQRCode    UpdateType = "UpdateForSetQRCode"
	UpdateForUpdateQRCode UpdateType = "UpdateForUpdateQRCode"

	UpdateForDelete        UpdateType = "UpdateForDelete"
	UpdateForUpdateTimeout UpdateType = "UpdateForUpdateTimeout"
)
