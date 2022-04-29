package cli

type Client struct{}

type jsonSecret struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	LastModifiedGMT string `json:"last_modified_gmt"`
	LastTouch       string `json:"last_touch"`
	Group           string `json:"group"`
	Share           string `json:"share"`
	URL             string `json:"url"`
	Notes           string `json:"note"`
}
