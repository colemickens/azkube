package util

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/Azure/azure-sdk-for-go/arm/authorization"
	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	"github.com/Azure/azure-sdk-for-go/arm/resources/subscriptions"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
	log "github.com/Sirupsen/logrus"
)

const (
	AzkubeClientID = "a87032a7-203c-4bf7-913c-44c50d23409a"
	//AzureActiveDirectoryScope = "https://graph.windows.net/"
	//AzureResourceManagerScope = "https://management.core.windows.net/"
)

var (
	RequiredResourceProviders = []string{"Microsoft.Compute", "Microsoft.Storage", "Microsoft.Network"}
)

type AzureClient struct {
	SubscriptionID string
	TenantID       string
	Environment    azure.Environment

	DeploymentsClient     resources.DeploymentsClient
	GroupsClient          resources.GroupsClient
	RoleAssignmentsClient authorization.RoleAssignmentsClient
	ResourcesClient       resources.Client
	ProvidersClient       resources.ProvidersClient
	SubscriptionsClient   subscriptions.Client
	AdClient              AdClient
}

func NewClientWithDeviceAuth(azureEnvironment azure.Environment, subscriptionID, tenantID string) (*AzureClient, error) {
	oauthConfig, err := azureEnvironment.OAuthConfigForTenant(tenantID)
	if err != nil {
		return nil, err
	}

	client := &autorest.Client{}

	deviceCode, err := azure.InitiateDeviceAuth(client, *oauthConfig, AzkubeClientID, azureEnvironment.ServiceManagementEndpoint)
	if err != nil {
		return nil, err
	}
	fmt.Println(*deviceCode.Message)
	deviceToken, err := azure.WaitForUserCompletion(client, deviceCode)
	if err != nil {
		return nil, err
	}
	spt, err := azure.NewServicePrincipalTokenFromManualToken(*oauthConfig, AzkubeClientID, azureEnvironment.ServiceManagementEndpoint, *deviceToken)
	if err != nil {
		return nil, err
	}
	spt.Token = *deviceToken

	return makeAzureClient(azureEnvironment, subscriptionID, tenantID, spt)
}

func NewClientWithClientSecret(azureEnvironment azure.Environment, subscriptionID, tenantID, clientID, clientSecret string) (*AzureClient, error) {
	oauthConfig, err := azureEnvironment.OAuthConfigForTenant(tenantID)
	if err != nil {
		return nil, err
	}

	spt, err := azure.NewServicePrincipalToken(*oauthConfig, clientID, clientSecret, azureEnvironment.ServiceManagementEndpoint)
	if err != nil {
		return nil, err
	}

	return makeAzureClient(azureEnvironment, tenantID, subscriptionID, spt)
}

func NewClientWithClientCertificate(azureEnvironment azure.Environment, subscriptionID, tenantID, clientID, certificatePath, privateKeyPath string) (*AzureClient, error) {
	oauthConfig, err := azureEnvironment.OAuthConfigForTenant(tenantID)
	if err != nil {
		return nil, err
	}

	certificateData, err := ioutil.ReadFile(certificatePath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read certificate: %q", err)
	}

	block, _ := pem.Decode(certificateData)
	if block == nil {
		return nil, fmt.Errorf("Failed to decode pem block from certificate")
	}

	certificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse certificate: %q", err)
	}

	privateKey, err := parseRsaPrivateKey(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse rsa private key: %q", err)
	}

	spt, err := azure.NewServicePrincipalTokenFromCertificate(*oauthConfig, clientID, certificate, privateKey, azureEnvironment.ServiceManagementEndpoint)

	return makeAzureClient(azureEnvironment, tenantID, subscriptionID, spt)
}

func makeAzureClient(azureEnvironment azure.Environment, subscriptionID, tenantID string, armSpt *azure.ServicePrincipalToken) (*AzureClient, error) {
	adSpt := *armSpt

	err := adSpt.RefreshExchange(azureEnvironment.GraphEndpoint)
	if err != nil {
		return nil, err
	}

	azureClient := &AzureClient{
		SubscriptionID: subscriptionID,
		TenantID:       tenantID,
		Environment:    azureEnvironment,

		DeploymentsClient:     resources.NewDeploymentsClient(subscriptionID),
		GroupsClient:          resources.NewGroupsClient(subscriptionID),
		RoleAssignmentsClient: authorization.NewRoleAssignmentsClient(subscriptionID),
		ResourcesClient:       resources.NewClient(subscriptionID),
		ProvidersClient:       resources.NewProvidersClient(subscriptionID),
		AdClient:              AdClient{Client: autorest.Client{}, TenantID: tenantID},
	}

	azureClient.DeploymentsClient.Authorizer = armSpt
	azureClient.GroupsClient.Authorizer = armSpt
	azureClient.RoleAssignmentsClient.Authorizer = armSpt
	azureClient.ResourcesClient.Authorizer = armSpt
	azureClient.ProvidersClient.Authorizer = armSpt
	azureClient.AdClient.Authorizer = &adSpt

	err = azureClient.ensureProvidersRegistered(subscriptionID)
	if err != nil {
		return nil, err
	}

	return azureClient, nil
}

func (azureClient *AzureClient) ensureProvidersRegistered(subscriptionID string) error {
	registeredProviders, err := azureClient.ProvidersClient.List(nil)
	if err != nil {
		return err
	}
	if registeredProviders.Value == nil {
		return fmt.Errorf("Providers list was nil. subscription=%q", subscriptionID)
	}

	m := make(map[string]bool)
	for _, provider := range *registeredProviders.Value {
		m[strings.ToLower(to.String(provider.Namespace))] = to.String(provider.RegistrationState) == "Registered"
	}

	for _, provider := range RequiredResourceProviders {
		registered, ok := m[strings.ToLower(provider)]
		if !ok {
			return fmt.Errorf("Unknown resource provider %q", provider)
		}
		if registered {
			log.Debugf("Already registered for %q", provider)
		} else {
			log.Info("Registering subscription to resource provider. provider=%q subscription=%q", provider, subscriptionID)
			if _, err := azureClient.ProvidersClient.Register(provider); err != nil {
				return err
			}
		}
	}
	return nil
}

func parseRsaPrivateKey(path string) (*rsa.PrivateKey, error) {
	privateKeyData, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(privateKeyData)
	if block == nil {
		return nil, fmt.Errorf("Failed to decode a pem block from private key")
	}

	privatePkcs1Key, errPkcs1 := x509.ParsePKCS1PrivateKey(block.Bytes)
	if errPkcs1 == nil {
		return privatePkcs1Key, nil
	}

	privatePkcs8Key, errPkcs8 := x509.ParsePKCS8PrivateKey(block.Bytes)
	if errPkcs8 == nil {
		privatePkcs8RsaKey, ok := privatePkcs8Key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("Pkcs8 contained non-RSA key. Expected RSA key.")
		}
		return privatePkcs8RsaKey, nil
	}

	return nil, fmt.Errorf("Failed to parse private key as Pkcs#1 or Pkcs#8. (%s). (%s).", errPkcs1, errPkcs8)
}
