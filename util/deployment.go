package util

import (
	//	"fmt"
	//	"time"

	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	log "github.com/Sirupsen/logrus"
)

func (azureClient *AzureClient) DeployTemplate(resourceGroupName, deploymentName string, template map[string]interface{}, parameters map[string]interface{}) (response *resources.DeploymentExtended, err error) {
	deployment := resources.Deployment{
		Properties: &resources.DeploymentProperties{
			Template:   &template,
			Parameters: &parameters,
			Mode:       resources.Incremental,
		},
	}

	log.Infof("Starting ARM Deployment. This will take some time. deployment=%q", deploymentName)
	_, err = azureClient.DeploymentsClient.CreateOrUpdate(
		resourceGroupName,
		deploymentName,
		deployment)
	if err != nil {
		return nil, err
	}
	log.Infof("Finished ARM Deployment. deployment=%q", deploymentName)

	return nil, nil
}

func (azureClient *AzureClient) DeployFlavor(flavor string, flavorArgs FlavorArguments, outputDirectory string) error {
	masterScript, err := InterpolateArmPlaceholders(flavor, "master-cloudconfig.in.yml")
	if err != nil {
		return err
	}

	nodeScript, err := InterpolateArmPlaceholders(flavor, "node-cloudconfig.in.yml")
	if err != nil {
		return err
	}

	template, err := PopulateTemplateMap(flavor, "cluster-deploy.in.json",
		struct{ MasterScript, NodeScript string }{masterScript, nodeScript})
	if err != nil {
		return err
	}

	parameters, err := PopulateTemplateMap(flavor, "cluster-parameters.in.json", flavorArgs)
	if err != nil {
		return err
	}

	utilScript, err := PopulateTemplate(flavor, "util.in.sh", flavorArgs)
	if err != nil {
		return err
	}

	err = SaveDeploymentMap(outputDirectory, "cluster-deploy.json", template, 0600)
	if err != nil {
		return err
	}
	err = SaveDeploymentMap(outputDirectory, "cluster-parameters.json", parameters, 0600)
	if err != nil {
		return err
	}

	err = SaveDeploymentFile(outputDirectory, "util.sh", utilScript, 0700)
	if err != nil {
		return err
	}

	_, err = azureClient.DeployTemplate(
		flavorArgs.ResourceGroup,
		flavorArgs.DeploymentName,
		template,
		parameters)
	if err != nil {
		return err
	}

	return nil
}
