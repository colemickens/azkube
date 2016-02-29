package util

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	log "github.com/Sirupsen/logrus"
)

func (d *Deployer) DoDeployment(resourceGroupName, deploymentName string, template map[string]interface{}, parameters map[string]interface{}, waitDeployment bool) (response *resources.DeploymentExtended, err error) {
	deployment := resources.Deployment{
		Properties: &resources.DeploymentProperties{
			Template:   &template,
			Parameters: &parameters,
			Mode:       resources.Incremental,
		},
	}

	deploymentResponse, err := d.DeploymentsClient.CreateOrUpdate(
		resourceGroupName,
		resourceGroupName+"-"+deploymentName+"-deploy",
		deployment)
	if err != nil {
		panic(err)
	}

	if waitDeployment {
		deploymentName := *deploymentResponse.Name
		response, err = d.WaitDeployment(resourceGroupName, deploymentName)
		return response, err
	}

	return &deploymentResponse, err
}

func (d *Deployer) WaitDeployment(resourceGroup, deploymentName string) (*resources.DeploymentExtended, error) {
	var err error
	var response resources.DeploymentExtended
	var infoUpdateCount int = 0
	log.Infof("deployment: waiting for completion. deployment=%q", deploymentName)
	for {
		time.Sleep(5 * time.Second)

		response, err = d.DeploymentsClient.Get(resourceGroup, deploymentName)
		if err != nil {
			return &response, err
		}

		state := response.Properties.ProvisioningState

		if state == nil || *state == "Accepted" || *state == "Running" {
			infoUpdateCount++
			if infoUpdateCount >= 6 {
				infoUpdateCount = 0
				log.Infof("deployment: in progress. deployment=%q state=%q", deploymentName, *state)
			}
			continue
		}

		if *state == "Succeeded" {
			log.Infof("deployment: finished. deployment=%q", deploymentName)
			return &response, nil
		} else if *state == "Failed" {
			return &response, fmt.Errorf("deployment: failed! deployment=%q", deploymentName)
		} else {
			return &response, fmt.Errorf("deployment: {unknown state}. deployment=%s", deploymentName, *state)
		}
	}

	return &response, nil
}
