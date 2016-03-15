package cmd

import (
	//"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/colemickens/azkube/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	scaleLongDescription = "scale a deployment's vm scale set"
)

type ScaleArguments struct {
	DeploymentName string
	ResourceGroup  string
	NodeCount      int
	NodeSize       string
}

func NewScaleDeploymentCmd() *cobra.Command {
	var scaleCmd = &cobra.Command{
		Use:   "scale",
		Short: scaleLongDescription,
		Long:  scaleLongDescription,
		Run:   runScale,
	}

	flags := scaleCmd.Flags()
	flags.String("deployment-name", "", "deployment name (required)")
	flags.String("resource-group", "", "resource group name (derived from --deployment-name if unset)")
	flags.Int("node-count", -1, "number of nodes to scale to (required)")
	flags.String("node-size", "", "size of nodes to scale (required")

	return scaleCmd
}

func parseScaleArgs(cmd *cobra.Command, args []string) (RootArguments, ScaleArguments) {
	rootArgs := parseRootArgs(cmd, args)

	flags := cmd.Flags()
	viper.BindPFlag("deployment-name", flags.Lookup("deployment-name"))
	viper.BindPFlag("resource-group", flags.Lookup("resource-group"))
	viper.BindPFlag("node-count", flags.Lookup("node-count"))
	viper.BindPFlag("node-size", flags.Lookup("node-size"))

	scaleArgs := ScaleArguments{
		DeploymentName: viper.GetString("deployment-name"),
		ResourceGroup:  viper.GetString("resource-group"),
		NodeCount:      viper.GetInt("node-count"),
		NodeSize:       viper.GetString("node-size"),
	}

	if scaleArgs.DeploymentName == "" {
		log.Errorf("scaleargs: --deployment-name must be set!")
	}

	if scaleArgs.ResourceGroup == "" {
		scaleArgs.ResourceGroup = scaleArgs.DeploymentName
		log.Warnf("scaleargs: --resource-group is unset. (inferring it from --deployment-name: %q)", scaleArgs.ResourceGroup)
	}

	if scaleArgs.NodeCount == -1 {
		log.Errorf("scaleargs: --node-count must be specified")
	}

	if scaleArgs.NodeSize == "" {
		log.Errorf("scaleargs: --node-size must be specified")
	}

	return rootArgs, scaleArgs
}

func runScale(cmd *cobra.Command, args []string) {
	rootArgs, scaleArgs := parseScaleArgs(cmd, args)
	azureClient, err := getClient(rootArgs)
	if err != nil {
		log.Fatalf("Error occurred while creating the Azure client: %q", err)
	}

	var flavor string = "coreos"

	flavorArgs := util.FlavorArguments{
		DeploymentName: scaleArgs.DeploymentName,
		NodeCount:      scaleArgs.NodeCount,
		NodeSize:       scaleArgs.NodeSize,
	}

	template, err := util.PopulateTemplateMap(flavor, "scale-deploy.in.json", struct{}{})
	if err != nil {
		log.Fatalf("Failed to populate scale deployment template")
	}
	parameters, err := util.PopulateTemplateMap(flavor, "scale-parameters.in.json", flavorArgs)
	if err != nil {
		log.Fatalf("Failed to populate scale deployment parameters")
	}

	_, err = azureClient.DeployTemplate(
		scaleArgs.ResourceGroup,
		scaleArgs.DeploymentName+"-scale",
		template,
		parameters)
	if err != nil {
		log.Fatalf("Failed to deploy the scale change: %q", err)
	}
}
