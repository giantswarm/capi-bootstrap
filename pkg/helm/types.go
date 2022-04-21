package helm

type Client struct {
	reposInitialized bool

	KubeconfigPath string
}

type helmRepo struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}
