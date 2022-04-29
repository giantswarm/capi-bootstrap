package key

const (
	EncryptionKeySecretShare = "Shared-Team Rocket"
	EncryptionKeySecretGroup = "Encryption Keys"
)

func EncryptionKeySecretName(clusterName string) string {
	return clusterName
}

const (
	GeneratorNameAWSIAM      = "awsiam"
	GeneratorNameCA          = "ca"
	GeneratorNameGitHubOAuth = "githuboauth"
	GeneratorNameLastpass    = "lastpass"
	GeneratorNameTaylorbot   = "taylorbot"
)
