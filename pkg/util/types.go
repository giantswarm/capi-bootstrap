package util

type CloudConfigCloudAuth struct {
	AuthURL        string `json:"auth_url"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	UserDomainName string `json:"user_domain_name"`
	ProjectID      string `json:"project_id"`
}

type CloudConfigCloud struct {
	Auth               CloudConfigCloudAuth `json:"auth"`
	Verify             bool                 `json:"verify"`
	RegionName         string               `json:"region_name"`
	Interface          string               `json:"interface"`
	IdentityAPIVersion int                  `json:"identity_api_version"`
}

type CloudConfig struct {
	Clouds map[string]CloudConfigCloud `json:"clouds"`
}
