package util

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"

	"github.com/Azure/azure-sdk-for-go/arm/authorization"
	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/spf13/viper"
)

const (
	AzkubeClientID            = "a87032a7-203c-4bf7-913c-44c50d23409a"
	AzureActiveDirectoryScope = "https://graph.windows.net/"
	AzureResourceManagerScope = "https://management.core.windows.net/"
)

func NewDeployer() (*Deployer, error) {
	oauthConfig, err := azure.PublicCloud.OAuthConfigForTenant(viper.GetString(rootArgNames.TenantID))
	if err != nil {
		return nil, err
	}

	client := &autorest.Client{}
	resource := AzureResourceManagerScope
	//resource := AzureActiveDirectoryScope

	var spt *azure.ServicePrincipalToken
	switch viper.GetString(rootArgNames.AuthMethod) {
	case "device":
		deviceCode, err := azure.InitiateDeviceAuth(client, *oauthConfig, AzkubeClientID, resource)
		if err != nil {
			return nil, err
		}
		fmt.Println(*deviceCode.Message)
		deviceToken, err := azure.WaitForUserCompletion(client, deviceCode)
		if err != nil {
			return nil, err
		}
		spt, err = azure.NewServicePrincipalTokenFromManualToken(*oauthConfig, AzkubeClientID, resource, *deviceToken)
		if err != nil {
			return nil, err
		}
		spt.Token = *deviceToken

	case "client_secret":
		spt, err = azure.NewServicePrincipalToken(*oauthConfig, viper.GetString(rootArgNames.ClientID), viper.GetString(rootArgNames.ClientSecret), resource)
		if err != nil {
			return nil, err
		}

	case "client_certificate":
		spt, err = newDeployerFromCertificate(*oauthConfig, viper.GetString(rootArgNames.SubscriptionID), viper.GetString(rootArgNames.ClientID), viper.GetString(rootArgNames.CertificatePath), viper.GetString(rootArgNames.PrivateKeyPath), resource)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("The authentication method is unknown: %q", viper.GetString(rootArgNames.AuthMethod))
	}

	var resourcesScopeSpt azure.ServicePrincipalToken = *spt
	var adScopeSpt azure.ServicePrincipalToken = *spt

	alternativeResource := AzureActiveDirectoryScope
	//alternativeResource := AzureResourceManagerScope

	err = adScopeSpt.RefreshExchange(alternativeResource)
	//err = resourcesScopeSpt.RefreshExchange(alternativeResource)
	if err != nil {
		return nil, err
	}

	deployer := &Deployer{
		DeploymentsClient:     resources.NewDeploymentsClient(viper.GetString(rootArgNames.SubscriptionID)),
		GroupsClient:          resources.NewGroupsClient(viper.GetString(rootArgNames.SubscriptionID)),
		RoleAssignmentsClient: authorization.NewRoleAssignmentsClient(viper.GetString(rootArgNames.SubscriptionID)),
		AdClient:              AdClient{Client: autorest.Client{}, TenantID: viper.GetString(rootArgNames.TenantID)},
	}

	deployer.DeploymentsClient.Authorizer = &resourcesScopeSpt
	deployer.GroupsClient.Authorizer = &resourcesScopeSpt
	deployer.RoleAssignmentsClient.Authorizer = &resourcesScopeSpt
	deployer.AdClient.Authorizer = &adScopeSpt

	return deployer, nil
}

func newDeployerFromCertificate(oauthConfig azure.OAuthConfig, subscriptionID, clientID, certificatePath, privateKeyPath, resource string) (*azure.ServicePrincipalToken, error) {
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

	return azure.NewServicePrincipalTokenFromCertificate(oauthConfig, clientID, certificate, privateKey, resource)
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
