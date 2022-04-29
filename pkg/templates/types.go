package templates

type TemplateData struct {
	BaseDomain  string
	ClusterName string
	Customer    string
	Pipeline    string
	Provider    string
	Secrets     map[string]interface{}
}

type TemplateFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type TemplateSecret struct {
	Key       string `json:"key"`
	Generator string `json:"generator"`

	AWSIAM      *AWSIAMTemplateInputs      `json:"awsiam"`
	GitHubOAuth *GitHubOAuthTemplateInputs `json:"githuboauth,omitempty"`
	Taylorbot   *TaylorbotTemplateInputs   `json:"taylorbot,omitempty"`
	Lastpass    *LastpassTemplateInputs    `json:"lastpass,omitempty"`
}

type LastpassSecretRef struct {
	Share string `json:"share,omitempty"`
	Group string `json:"group,omitempty"`
	Name  string `json:"name"`
}

type AWSIAMTemplateInputs struct {
}

type GitHubOAuthTemplateInputs struct {
}

// InstallationInputs defines information about the current cluster. Not tied to any specific generator.
type InstallationInputs struct {
	BaseDomain  string
	ClusterName string
}

type TaylorbotTemplateInputs struct {
	GitHubCredentialsSecretRef LastpassSecretRef `json:"gitHubCredentialsSecretRef"`
}

type LastpassTemplateInputs struct {
	Format    string            `json:"format"`
	SecretRef LastpassSecretRef `json:"secretRef"`
}
