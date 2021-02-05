package profiles

type KubernetesOriginatingIdentityHeader struct {
	Username string                 `json:"username" binding:"required"`
	Uid      string                 `json:"uid" binding:"required"`
	Groups   []string               `json:"groups" binding:"required"`
	Extra    map[string]interface{} `json:"extra" binding:"required"`
}
