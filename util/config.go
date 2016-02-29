package util

import (
	"net"

	"github.com/Azure/azure-sdk-for-go/arm/authorization"
	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
)

type CommonProperties struct {
	DeploymentName string
	ResourceGroup  string
	Location       string
	TenantID       string
	SubscriptionID string
	MasterFQDN     string
	MasterIP       net.IP
}

type AppProperties struct {
	ApplicationID                string
	Name                         string
	IdentifierURL                string
	ServicePrincipalObjectID     string
	ServicePrincipalClientSecret string
}

type PkiKeyCertPair struct {
	CertificatePem string
	PrivateKeyPem  string
}

type PkiProperties struct {
	CA                    *PkiKeyCertPair
	ApiServer             *PkiKeyCertPair
	Kubelet               *PkiKeyCertPair
	Kubeproxy             *PkiKeyCertPair
	Scheduler             *PkiKeyCertPair
	ReplicationController *PkiKeyCertPair
}

type SshProperties struct {
	PrivateKeyPem string
}

type VaultProperties struct {
	Name string
}

type SecretsProperties struct {
	ServicePrincipalSecretURL string
}

type MyriadProperties struct {
	PodCidr     string
	ServiceCidr string

	MasterPrivateIp string
	ApiServiceIp    string
	DnsServiceIp    string

	HyperkubeContainerSpec string
}

// appears in same order as in variable in ARM template
type State struct {
	Common  *CommonProperties
	App     *AppProperties
	Pki     *PkiProperties
	Ssh     *SshProperties
	Vault   *VaultProperties
	Secrets *SecretsProperties
	Myriad  *MyriadProperties
}

type Deployer struct {
	DeploymentsClient     resources.DeploymentsClient
	GroupsClient          resources.GroupsClient
	RoleAssignmentsClient authorization.RoleAssignmentsClient
	AdClient              AdClient
}

type VaultTemplateInput struct {
	VaultName                string
	TenantID                 string
	ServicePrincipalObjectID string
}

// this fills out the template for what's needed for Kubernetes at runtime to talk to the underlying Azure Service Fabric
// it's also needed by this tool, which will be automatically run on the boxes during deployment, using the SP cert to bootstrap the rest of the secrets
// for the Kubernetes components (and eventually the Etcd components as well)
type RuntimeConfigTemplateInput struct {
	PrivateKeyPath  string
	CertificatePath string
	SubscriptionID  string
	TenantID        string
	ApplicationID   string
	VaultName       string
}

type CloudConfigTemplateInput struct {
	DeploymentName string

	RuntimeConfigFile      string
	HyperkubeContainerSpec string

	ServiceCidr string
	PodCidr     string

	MasterIP   string
	MasterFQDN string

	DnsServiceIP string
}

type MyriadTemplateInput struct {
	DeploymentName string

	MasterVmSize         string
	NodeVmSize           string
	NodeVmssInitialCount int
	Username             string

	VaultName                 string
	ServicePrincipalSecretURL string

	PodCidr     string
	ServiceCidr string

	SshPublicKeyData string

	MasterCloudConfig string
	NodeCloudConfig   string
}
