package sops

import "github.com/giantswarm/capi-bootstrap/pkg/lastpass"

type EncryptionKey struct {
	PrivateKey string `json:"privateKey"`
	PublicKey  string `json:"publicKey"`
	Type       string `json:"type"`
}

type Config struct {
	LastpassClient *lastpass.Client
	ClusterName    string
}

type Client struct {
	lastpassClient *lastpass.Client
	clusterName    string
	encryptionKey  *EncryptionKey
}

type CreationRule struct {
	PathRegex      string `json:"path_regex"`
	EncryptedRegex string `json:"encrypted_regex"`
	Age            string `json:"age"`
}

type SopsConfig struct {
	CreationRules []CreationRule `json:"creation_rules"`
}
