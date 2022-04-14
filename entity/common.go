package entity

const (
	ModuleName = "lobby"
)

const (
	BucketAvatar   = "avatar"
	AvatarFileName = "%s_image"
)

func InterfaceToString(inf interface{}) string {
	if inf == nil {
		return ""
	}
	str, ok := inf.(string)
	if !ok {
		return ""
	}
	return str
}
