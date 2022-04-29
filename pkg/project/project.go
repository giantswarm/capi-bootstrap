package project

var (
	description = "Tool for automating the configuration of CAPI management clusters"
	gitSHA      = "n/a"
	name        = "capi-bootstrap"
	source      = "https://github.com/giantswarm/capi-bootstrap"
	version     = "0.1.0-dev"
)

func Description() string {
	return description
}

func GitSHA() string {
	return gitSHA
}

func Name() string {
	return name
}

func Source() string {
	return source
}

func Version() string {
	return version
}
