package secret

type GeneratedSecretDefinition struct {
	Key       string `json:"key"`
	Generator string `json:"generator"`

	BaseDomain  string `json:"-"`
	ClusterName string `json:"-"`

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

type TaylorbotTemplateInputs struct {
	GitHubCredentialsSecretRef LastpassSecretRef `json:"gitHubCredentialsSecretRef"`
}

type LastpassTemplateInputs struct {
	Format    string            `json:"format"`
	SecretRef LastpassSecretRef `json:"secretRef"`
}
