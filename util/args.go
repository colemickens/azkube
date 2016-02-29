package util

import ()

type RootArguments struct {
	TenantID        string
	SubscriptionID  string
	AuthMethod      string
	ClientID        string
	ClientSecret    string
	CertificatePath string
	PrivateKeyPath  string
}

var RootArgNames = RootArguments{
	TenantID:        "tenant-id",
	SubscriptionID:  "subscription-id",
	AuthMethod:      "auth-method",
	ClientID:        "client-id",
	ClientSecret:    "client-secret",
	CertificatePath: "certificate-path",
	PrivateKeyPath:  "private-key-path",
}
var rootArgNames = RootArgNames

type DeployArguments struct {
	OutputDirectory         string
	DeploymentName          string
	ResourceGroup           string
	Location                string
	MasterSize              string
	NodeSize                string
	NodeCount               string
	Username                string
	MasterFQDN              string
	MasterExtraFQDNs        string
	KubernetesReleaseURL    string
	KubernetesHyperkubeSpec string
}

var DeployArgNames = DeployArguments{
	OutputDirectory:         "output-directory",
	DeploymentName:          "deployment-name",
	ResourceGroup:           "resource-group",
	Location:                "location",
	MasterSize:              "master-size",
	NodeSize:                "node-size",
	NodeCount:               "node-count",
	Username:                "username",
	MasterFQDN:              "master-fqdn",
	MasterExtraFQDNs:        "master-fqdns",
	KubernetesReleaseURL:    "kubernetes-release-url",
	KubernetesHyperkubeSpec: "kubernetes-hyperkube-spec",
}
var deployArgNames = DeployArgNames

// part of interface to flavors.
// this is what flavors get.
// make a way to opt-in/opt-out?
type FlavorArguments struct {
	DeploymentName string

	TenantID string

	MasterSize       string
	NodeSize         string
	NodeCount        int
	Username         string
	SshPublicKeyData string

	ServicePrincipalClientID     string
	ServicePrincipalClientSecret string

	MasterFQDN string

	KubernetesReleaseURL    string
	KubernetesHyperkubeSpec string

	CAKeyPair        *PkiKeyCertPair
	ApiserverKeyPair *PkiKeyCertPair
	ClientKeyPair    *PkiKeyCertPair
}
