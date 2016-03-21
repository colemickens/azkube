package util

import (
	"net"
)

type FlavorArguments struct {
	DeploymentName string
	ResourceGroup  string

	TenantID string

	MasterSize       string
	NodeSize         string
	NodeCount        int
	Username         string
	SshPublicKeyData string

	ServicePrincipalClientID     string
	ServicePrincipalClientSecret string

	MasterFQDN      string
	MasterPrivateIP net.IP
	ClusterDomain   string

	KubernetesReleaseURL    string
	KubernetesHyperkubeSpec string

	CAKeyPair        *PkiKeyCertPair
	ApiserverKeyPair *PkiKeyCertPair
	ClientKeyPair    *PkiKeyCertPair
}
