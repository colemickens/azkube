package util

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/arm/authorization"
	"github.com/Azure/go-autorest/autorest"
	log "github.com/Sirupsen/logrus"
	"github.com/pborman/uuid"
	"github.com/spf13/viper"
)

const (
	AzureAdApiVersion = "1.6"
	AzureAdBaseURL    = "https://graph.windows.net/%s"

	AzureRoleManagementApiVersion = "2015-07-01"
	AzureManagementBaseURL        = "https://management.azure.com/{tenant-id}"

	AzureAdRoleReferenceTemplate = "/subscriptions/{subscription-id}/providers/Microsoft.Authorization/roleDefinitions/{role-definition-id}"
	AzureAdContributorRoleId     = "b24988ac-6180-42a0-ab88-20f7382dd24c"
	AzureAdOwnerRoleId           = "8e3af657-a8ff-443c-a75c-2fe8c4bcb635"
	AzureAdAssignedRoleId        = AzureAdOwnerRoleId

	ServicePrincipalKeySize = 4096
)

var (
	spMissingMessageRegexp *regexp.Regexp
)

func init() {
	spMissingMessageRegexp = regexp.MustCompile(`Principal (.+) does not exist in the directory (.+)\.`)
}

type AdClient struct {
	autorest.Client
	TenantID string
}

type AdApplication struct {
	ApplicationID string `json:"appId,omitempty"`    // readonly
	ObjectID      string `json:"objectId,omitempty"` // readonly

	AvailableToOtherTenants bool                   `json:"availableToOtherTenants"`
	DisplayName             string                 `json:"displayName,omitempty"`
	Homepage                string                 `json:"homepage,omitempty"`
	IdentifierURIs          []string               `json:"identifierUris,omitempty"`
	PasswordCredentials     []AdPasswordCredential `json:"passwordCredentials,omitempty"`
}

type AdPasswordCredential struct {
	KeyId     string `json:"keyId,omitempty"`
	StartDate string `json:"startDate,omitempty"`
	EndDate   string `json:"endDate,omitempty"`
	Value     string `json:"value,omitempty"`
}

type AdServicePrincipal struct {
	ObjectID string `json:"objectId,omitempty"` // readonly

	ApplicationID  string `json:"appId,omitempty"`
	AccountEnabled bool   `json:"accountEnabled,omitempty"`
	// ServicePrincipalNames []string `json:"servicePrincipalNames,omitempty"`
}

type AdRoleAssignment struct {
	RoleDefinitionID string `json:"roleDefinitionId,omitempty"`
	PrincipalID      string `json:"principalId,omitempty"`
}

func (a *AdClient) CreateApp(appName, appURL string) (applicationID, servicePrincipalObjectID, servicePrincipalClientSecret string, err error) {
	notBefore := time.Now()
	notAfter := time.Now().Add(5 * 365 * 24 * time.Hour)
	notAfter = time.Now().Add(10000 * 24 * time.Hour)

	startDate := notBefore.Format(time.RFC3339)
	endDate := notAfter.Format(time.RFC3339)

	servicePrincipalClientSecret = uuid.New()

	log.Infof("ad: creating application with name=%q identifierURL=%q", appName, appURL)

	applicationReq := AdApplication{
		AvailableToOtherTenants: false,
		DisplayName:             appName,
		Homepage:                appURL,
		IdentifierURIs:          []string{appURL},
		PasswordCredentials: []AdPasswordCredential{
			AdPasswordCredential{
				KeyId:     uuid.New(),
				StartDate: startDate,
				EndDate:   endDate,
				Value:     servicePrincipalClientSecret,
			},
		},
	}

	q := map[string]interface{}{"api-version": AzureAdApiVersion}

	azureAdURL := fmt.Sprintf(AzureAdBaseURL, a.TenantID)

	req, err := autorest.Prepare(&http.Request{},
		autorest.AsJSON(),
		autorest.AsPost(),
		autorest.WithBaseURL(azureAdURL),
		autorest.WithPath("applications"),
		autorest.WithQueryParameters(q),
		autorest.WithJSON(applicationReq))
	if err != nil {
		log.Errorf("ad: failed to prepare the application creation request")
		return "", "", "", err
	}

	resp, err := a.Send(req)
	if err != nil {
		log.Errorf("ad: failed to send the application creation request")
		return "", "", "", err
	}

	var applicationResp AdApplication
	err = autorest.Respond(
		resp,
		autorest.WithErrorUnlessStatusCode(http.StatusCreated),
		autorest.ByUnmarshallingJSON(&applicationResp))
	if err != nil {
		log.Errorf("ad: failed to respond to application creation response")
		return "", "", "", err
	}

	applicationID = applicationResp.ApplicationID

	servicePrincipalReq := AdServicePrincipal{
		ApplicationID:  applicationID,
		AccountEnabled: true,
		// ServicePrincipalNames: []string{appURL},
	}

	log.Infof("ad: creating servicePrincipal for applicationID: %q", applicationID)

	req, err = autorest.Prepare(&http.Request{},
		autorest.AsJSON(),
		autorest.AsPost(),
		autorest.WithBaseURL(azureAdURL),
		autorest.WithPath("servicePrincipals"),
		autorest.WithQueryParameters(q),
		autorest.WithJSON(servicePrincipalReq))
	if err != nil {
		log.Errorf("ad: failed to prepare the servicePrincipal creation request")
		return "", "", "", err
	}

	resp, err = a.Send(req)
	if err != nil {
		log.Errorf("ad: failed to send the servicePrincipal creation request")
		return "", "", "", err
	}

	var servicePrincipalResp AdServicePrincipal
	err = autorest.Respond(
		resp,
		autorest.WithErrorUnlessStatusCode(http.StatusCreated),
		autorest.ByUnmarshallingJSON(&servicePrincipalResp))
	if err != nil {
		log.Errorf("ad: failed to respond to the servicePrincipal creation request")
		return "", "", "", err
	}

	servicePrincipalObjectID = servicePrincipalResp.ObjectID

	return applicationID, servicePrincipalObjectID, servicePrincipalClientSecret, nil
}

func (d *Deployer) CreateRoleAssignment(resourceGroup string, servicePrincipalObjectID string) error {
	roleAssignmentName := uuid.New()

	roleDefinitionId := strings.Replace(AzureAdRoleReferenceTemplate, "{subscription-id}", viper.GetString(rootArgNames.SubscriptionID), -1)
	roleDefinitionId = strings.Replace(roleDefinitionId, "{role-definition-id}", AzureAdAssignedRoleId, -1)

	scope := fmt.Sprintf("subscriptions/%s/resourceGroups/%s", viper.GetString(rootArgNames.SubscriptionID), resourceGroup)

	log.Infof("ad: creating role assignment for servicePrincipal (objectId=%q)", servicePrincipalObjectID)

	roleAssignmentParameters := authorization.RoleAssignmentCreateParameters{
		Properties: &authorization.RoleAssignmentProperties{
			RoleDefinitionID: &roleDefinitionId,
			PrincipalID:      &servicePrincipalObjectID,
		},
	}

	for {
		_, err := d.RoleAssignmentsClient.Create(
			scope,
			roleAssignmentName,
			roleAssignmentParameters,
		)
		if err != nil {
			/*
				TODO: fix when azure-sdk-for-go is regenerated
				azureErr, ok := err.(*azure.DetailedError)
				if ok && spMissingMessageRegexp.MatchString(azureErr.ServiceError.Message) {
					log.Warnf("failed to create role assignment (will retry): %q", err)
					continue
				} else {
					log.Errorf("failed to create role assignment: %q", err)
					return err
				}
			*/
			log.Warnf("ad: failed to create role assignment (will retry): %q", err)
			time.Sleep(3 * time.Second)
			continue
		}
		break
	}

	return nil
}
