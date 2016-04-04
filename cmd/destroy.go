package cmd

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	destroyLongDescription = "destroy a deployment (and its containing resource group)"
)

type DestroyArguments struct {
	DeploymentName string
	ResourceGroup  string
	SkipConfirm    bool
}

func NewDestroyDeploymentCmd() *cobra.Command {
	var destroyCmd = &cobra.Command{
		Use:   "destroy",
		Short: destroyLongDescription,
		Long:  destroyLongDescription,
		Run:   runDestroy,
	}
	flags := destroyCmd.Flags()

	flags.String("deployment-name", "", "deployment name to destroy (required)")
	flags.String("resource-group", "", "resource group to destroy (derived from --deployment-name if omitted)")
	flags.Bool("skip-confirm", false, "skip confimration of resource deletion")

	return destroyCmd
}

func parseDestroyArgs(cmd *cobra.Command, args []string) (RootArguments, DestroyArguments) {
	rootArgs := parseRootArgs(cmd, args)
	flags := cmd.Flags()

	viper.BindPFlag("deployment-name", flags.Lookup("deployment-name"))
	viper.BindPFlag("resource-group", flags.Lookup("resource-group"))
	viper.BindPFlag("skip-confirm", flags.Lookup("skip-confirm"))

	destroyArgs := DestroyArguments{
		DeploymentName: viper.GetString("deployment-name"),
		ResourceGroup:  viper.GetString("resource-group"),
		SkipConfirm:    viper.GetBool("skip-confirm"),
	}

	if destroyArgs.DeploymentName == "" {
		log.Fatalf("--deployment-name must be set.")
	}

	if destroyArgs.ResourceGroup == "" {
		destroyArgs.ResourceGroup = destroyArgs.DeploymentName
		log.Warnf("--resource-group is unset. deriving it from --deployment-name.")
	}

	if destroyArgs.SkipConfirm {
		log.Warnf("--skip-confirm is set. Will NOT confirm deletion!")
	}

	return rootArgs, destroyArgs
}

func runDestroy(cmd *cobra.Command, args []string) {
	rootArgs, destroyArgs := parseDestroyArgs(cmd, args)

	azureClient, err := getClient(rootArgs)
	if err != nil {
		log.Fatalf("Error occurred while creating the Azure client: %q", err)
	}

	resources, err := azureClient.ListResources(destroyArgs.ResourceGroup)
	if err != nil {
		log.Fatalf("Failed to list resources to destroy: %q", err)
	}

	for _, resource := range *resources {
		log.Warnf("Going to delete: %s (%s)", *resource.Name, *resource.Type)
	}

	log.Warnf("Going to delete a total of: %d item(s)", len(*resources))
	if !destroyArgs.SkipConfirm {
		for {
			var response string
			fmt.Printf("Enter 'y' to confirm deletion, or 'n' to abort: ")
			fmt.Scanln(&response)
			if response == "y" {
				break
			} else if response == "n" {
				log.Fatalf("Exit due to user abort")
			} else {
				log.Warnf("Unexpected choice: %q. Please enter 'y' or 'n'.", response)
			}
		}
	}

	log.Infof("Starting the deletion of resource group. resourceGroup=%q", destroyArgs.ResourceGroup)
	_, err = azureClient.GroupsClient.Delete(destroyArgs.ResourceGroup, nil)
	if err != nil {
		log.Fatalf("Failed to destroy the resource group: %q", err)
	}
	log.Infof("Finished the deletion of resource group. resourceGroup=%q", destroyArgs.ResourceGroup)
}
