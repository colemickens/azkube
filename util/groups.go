package util

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	log "github.com/Sirupsen/logrus"
)

func (azureClient *AzureClient) EnsureResourceGroup(name, location string) (resourceGroup *resources.ResourceGroup, err error) {
	log.Debugf("groups: ensuring resource group %q exists", name)
	response, err := azureClient.GroupsClient.CreateOrUpdate(name, resources.ResourceGroup{
		Name:     &name,
		Location: &location,
	})
	if err != nil {
		return &response, err
	}

	return &response, nil
}

func (azureClient *AzureClient) ListResources(resourceGroup string) (*[]resources.GenericResource, error) {
	var allResources []resources.GenericResource

	resourceList, err := azureClient.GroupsClient.ListResources(resourceGroup, "", nil)
	if err != nil {
		return nil, err
	}
	if resourceList.Value == nil {
		return nil, fmt.Errorf("resource list was nil")
	}
	allResources = append(allResources, *resourceList.Value...)
	for {
		moreResources, err := azureClient.GroupsClient.ListResourcesNextResults(resourceList)
		if err != nil {
			return nil, err
		}
		if moreResources.Value == nil || len(*moreResources.Value) == 0 {
			break
		}

		allResources = append(allResources, *moreResources.Value...)
	}

	return &allResources, nil
}
