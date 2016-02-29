package util

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	log "github.com/Sirupsen/logrus"
)

func (d *Deployer) EnsureResourceGroup(name, location string, waitDeployment bool) (resourceGroup *resources.ResourceGroup, err error) {
	log.Infof("groups: ensuring resource group %q exists", name)
	response, err := d.GroupsClient.CreateOrUpdate(name, resources.ResourceGroup{
		Name:     &name,
		Location: &location,
	})
	if err != nil {
		return &response, err
	}

	if waitDeployment {
		return d.WaitResourceGroup(name)
	}

	return &response, nil
}

func (d *Deployer) WaitResourceGroup(name string) (resourceGroup *resources.ResourceGroup, err error) {
	var response resources.ResourceGroup
	log.Infof("groups: waiting for resource group %q to finish provisioning", name)
	for {
		response, err = d.GroupsClient.Get(name)
		if err != nil {
			return &response, err
		}

		state := response.Properties.ProvisioningState

		if state == nil {
			continue
		}

		if *state == "Succeeded" {
			log.Infof("groups: resource group %q is provisioned", name)
			return &response, nil
		} else if *state == "Failed" {
			errorMessage := "groups: resource group %q failed to provision (ProvisioningState == 'Failed')"
			log.Errorf(errorMessage, name)
			return &response, fmt.Errorf(errorMessage, name)
		} else {
			errorMessage := "groups: resource group %q in unknown provision state (ProvisioningState == %q)"
			log.Errorf(errorMessage, name, *state)
			return &response, fmt.Errorf(errorMessage, name, *state)
		}
	}

	return nil, nil
}
