package lastpass

type Credentials struct {
	Username   string
	Password   string
	TOTPSecret string
}

type Client struct {
	authenticated bool
}

type Secret struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Fullname        string `json:"fullname"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	LastModifiedGmt string `json:"last_modified_gmt"`
	LastTouch       string `json:"last_touch"`
	Group           string `json:"group"`
	URL             string `json:"url"`
	Note            string `json:"note"`
}
