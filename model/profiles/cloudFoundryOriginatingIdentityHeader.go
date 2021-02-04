package profiles

type CloudFoundryOriginatingIdentityHeader struct {
	UserID string `json:"user_id" binding:"required"`
}
