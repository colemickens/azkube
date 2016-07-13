package cmd

import (
	"crypto/rsa"
	"fmt"
	"net"
	"os"
	"path"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/colemickens/azkube/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	deployLongDescription = "creates a new kubernetes cluster in Azure"

	kubernetesHyperkubeSpec = "gcr.io/google_containers/hyperkube-amd64:v1.3.0"
)

type DeployArguments struct {
	OutputDirectory             string
	DeploymentName              string
	ResourceGroup               string
	Location                    string
	MasterSize                  string
	NodeSize                    string
	NodeCount                   int
	KubernetesHyperkubeSpec     string
	Username                    string
	MasterFQDN                  string
	MasterPrivateIP             net.IP
	ClusterDomain               string
	MasterExtraFQDNs            []string
	ServicePrincipalPassthrough bool
	NoCloudProvider             bool
}

func NewDeployCmd() *cobra.Command {
	deployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "creates a new kubernetes cluster in Azure",
		Long:  deployLongDescription,
		Run:   runDeploy,
	}

	flags := deployCmd.Flags()
	flags.String("output-directory", "", "output directory (this is derived from --deployment-name if omitted)")
	flags.String("deployment-name", "", "deployment identifier (used to name output, resource group, and other resources)")
	flags.String("resource-group", "", "resource group to deploy to (this is derived from --deployment-name if omitted)")
	flags.String("location", "westus", "location to deploy Azure resource (these can be found by running `azure location list` with azure-xplat-cli)")
	flags.String("master-size", "Standard_A1", "size of the master virtual machine")
	flags.String("node-size", "Standard_A1", "size of the node virtual machines")
	flags.Int("node-count", 3, "initial number of node virtual machines")
	flags.String("kubernetes-hyperkube-spec", kubernetesHyperkubeSpec, "docker spec for hyperkube container to use")
	flags.String("username", "kube", "username to virtual machines")
	flags.String("master-fqdn", "", "fqdn for master (used for PKI). calculated from cloudapp dns for master's public ip")
	flags.String("master-private-ip", "10.0.1.4", "the internal vnet ip address to use for the master (used as a SAN in the PKI generation)")
	flags.String("cluster-domain", "cluster.local", "the dns suffix used in the cluster (used as a SAN in the PKI generation)")
	flags.StringSlice("master-extra-fqdns", []string{}, "comma delimited list of SANs for the master")
	flags.Bool("service-principal-passthrough", false, "bypass service principal creation and use deployers credentials for cluster's service principal")
	flags.Bool("no-cloud-provider", false, "skip service principal steps entirely. this suppresses creation of a new service principal and prevents passthrough of client_secret credentials")

	return deployCmd
}

func parseDeployArgs(cmd *cobra.Command, args []string) (RootArguments, DeployArguments) {
	rootArgs := parseRootArgs(cmd, args)

	flags := cmd.Flags()
	viper.BindPFlag("output-directory", flags.Lookup("output-directory"))
	viper.BindPFlag("deployment-name", flags.Lookup("deployment-name"))
	viper.BindPFlag("resource-group", flags.Lookup("resource-group"))
	viper.BindPFlag("location", flags.Lookup("location"))
	viper.BindPFlag("master-size", flags.Lookup("master-size"))
	viper.BindPFlag("node-size", flags.Lookup("node-size"))
	viper.BindPFlag("node-count", flags.Lookup("node-count"))
	viper.BindPFlag("kubernetes-release-url", flags.Lookup("kubernetes-release-url"))
	viper.BindPFlag("kubernetes-hyperkube-spec", flags.Lookup("kubernetes-hyperkube-spec"))
	viper.BindPFlag("username", flags.Lookup("username"))
	viper.BindPFlag("master-fqdn", flags.Lookup("master-fqdn"))
	viper.BindPFlag("master-private-ip", flags.Lookup("master-private-ip"))
	viper.BindPFlag("cluster-domain", flags.Lookup("cluster-domain"))
	viper.BindPFlag("master-extra-fqdns", flags.Lookup("master-extra-fqdns"))
	viper.BindPFlag("service-principal-passthrough", flags.Lookup("service-principal-passthrough"))
	viper.BindPFlag("no-cloud-provider", flags.Lookup("no-cloud-provider"))

	parsedMasterPrivateIP := net.ParseIP(viper.GetString("master-private-ip"))
	if parsedMasterPrivateIP == nil {
		log.Fatalf("Failed to parse --master-private-ip as an ip address")
	}

	deployArgs := DeployArguments{
		OutputDirectory:             viper.GetString("output-directory"),
		DeploymentName:              viper.GetString("deployment-name"),
		ResourceGroup:               viper.GetString("resource-group"),
		Location:                    viper.GetString("location"),
		MasterSize:                  viper.GetString("master-size"),
		NodeSize:                    viper.GetString("node-size"),
		NodeCount:                   viper.GetInt("node-count"),
		KubernetesHyperkubeSpec:     viper.GetString("kubernetes-hyperkube-spec"),
		Username:                    viper.GetString("username"),
		MasterFQDN:                  viper.GetString("master-fqdn"),
		MasterPrivateIP:             parsedMasterPrivateIP,
		ClusterDomain:               viper.GetString("cluster-domain"),
		MasterExtraFQDNs:            viper.GetStringSlice("master-extra-fqdns"),
		ServicePrincipalPassthrough: viper.GetBool("service-principal-passthrough"),
		NoCloudProvider:             viper.GetBool("no-cloud-provider"),
	}

	if deployArgs.DeploymentName == "" {
		deployArgs.DeploymentName = fmt.Sprintf("kube-%s", time.Now().Format("20060102-150405"))
		log.Warnf("--deployment-name is unset. Generated one: %q", deployArgs.DeploymentName)
	}

	if deployArgs.ResourceGroup == "" {
		deployArgs.ResourceGroup = deployArgs.DeploymentName
		log.Warnf("--resource-group is unset. Derived one from --deployment-name: %q", deployArgs.ResourceGroup)
	}

	if deployArgs.MasterFQDN == "" {
		deployArgs.MasterFQDN = fmt.Sprintf("%s.%s.cloudapp.azure.com", deployArgs.DeploymentName, deployArgs.Location)
		log.Warnf("--master-fqdn is unset. Derived one from input: %q.", deployArgs.MasterFQDN)
	}

	if deployArgs.ServicePrincipalPassthrough == true {
		if rootArgs.AuthMethod != "client_secret" {
			log.Fatalf("--service-principal-passthrough is only allowed when --auth-method=client_secret.")
		}
	}

	if deployArgs.OutputDirectory == "" {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatalf("--output-directory: ERROR: unable to get working directory for output")
		}

		deployArgs.OutputDirectory = path.Join(wd, "_deployments", deployArgs.DeploymentName)
		log.Warnf("--output-directory is unset. Using this location: %q.", deployArgs.OutputDirectory)

		err = os.MkdirAll(deployArgs.OutputDirectory, 0700)
		if err != nil {
			log.Fatalf("--output-directory: unable to create output directory for deployment: %q.", err)
		}
	}

	return rootArgs, deployArgs
}

func runDeploy(cmd *cobra.Command, args []string) {
	rootArgs, deployArgs := parseDeployArgs(cmd, args)

	azureClient, err := getClient(rootArgs)
	if err != nil {
		log.Fatalf("Error occurred while creating the Azure client: %q", err)
	}

	_, err = azureClient.EnsureResourceGroup(deployArgs.ResourceGroup, deployArgs.Location)
	if err != nil {
		log.Fatalf("Error occurred while ensuring the resource group is available: %q", err)
	}

	spClientID, spClientSecret, err := getCloudProviderCredentials(azureClient, rootArgs, deployArgs)
	if err != nil {
		log.Fatalf("Error occurred while creating service pricinpial assets: %q", err)
	}

	sshPrivateKey, sshPublicKeyString, err := util.CreateSaveSsh(deployArgs.Username, deployArgs.OutputDirectory)
	if err != nil {
		log.Fatalf("Error occurred while creating SSH assets: %q", err)
	}

	ca, apiserver, client, err := util.CreateSavePki(deployArgs.MasterFQDN, deployArgs.MasterExtraFQDNs, deployArgs.ClusterDomain, []net.IP{deployArgs.MasterPrivateIP}, deployArgs.OutputDirectory)
	if err != nil {
		log.Fatalf("Error occurred while creating PKI assets: %q", err)
	}

	flavorArgs := convertDeployArgsToFlavorArgs(deployArgs, azureClient.TenantID, spClientID, spClientSecret, sshPrivateKey, sshPublicKeyString, ca, apiserver, client)

	err = azureClient.DeployFlavor("coreos", flavorArgs, deployArgs.OutputDirectory)
	if err != nil {
		log.Fatalf("Error occurred while performing the deployment: %q", err)
	}

	err = util.ValidateKubernetes(flavorArgs)
	if err != nil {
		log.Fatalf("Error occurred while validating the deployment.")
	}

	log.Infof("Deployment Complete!")
	log.Infof("master: %q", "https://"+deployArgs.MasterFQDN+":6443")
	log.Infof("output: %q", deployArgs.OutputDirectory)
}

func getCloudProviderCredentials(azureClient *util.AzureClient, rootArgs RootArguments, deployArgs DeployArguments) (spClientID, spClientSecret string, err error) {
	if deployArgs.NoCloudProvider {
		spClientID = ""
		spClientSecret = ""
		return "", "", nil
	} else if deployArgs.ServicePrincipalPassthrough {
		return rootArgs.ClientID, rootArgs.ClientSecret, nil
	} else {
		appName := deployArgs.DeploymentName
		appURL := fmt.Sprintf("https://%s/", deployArgs.DeploymentName)
		_, spClientID, spClientSecret, err = azureClient.CreateApp(appName, appURL)
		if err != nil {
			return "", "", err
		}

		err = azureClient.CreateRoleAssignment(deployArgs.ResourceGroup, spClientID)
		if err != nil {
			return "", "", err
		}

		return spClientID, spClientSecret, nil
	}
}

func convertDeployArgsToFlavorArgs(deployArgs DeployArguments, tenantID string,
	spObjectID, spClientSecret string,
	sshPrivateKey *rsa.PrivateKey, sshPublicKeyString string,
	ca, apiserver, client *util.PkiKeyCertPair) util.FlavorArguments {
	flavorArgs := util.FlavorArguments{
		DeploymentName: deployArgs.DeploymentName,
		ResourceGroup:  deployArgs.ResourceGroup,

		TenantID: tenantID,

		MasterSize:       deployArgs.MasterSize,
		NodeSize:         deployArgs.NodeSize,
		NodeCount:        deployArgs.NodeCount,
		Username:         deployArgs.Username,
		SshPublicKeyData: sshPublicKeyString,

		KubernetesHyperkubeSpec: deployArgs.KubernetesHyperkubeSpec,

		ServicePrincipalClientID:     spObjectID,
		ServicePrincipalClientSecret: spClientSecret,

		MasterFQDN:      deployArgs.MasterFQDN,
		MasterPrivateIP: deployArgs.MasterPrivateIP,
		ClusterDomain:   deployArgs.ClusterDomain,

		CAKeyPair:        ca,
		ApiserverKeyPair: apiserver,
		ClientKeyPair:    client,
	}
	return flavorArgs
}
