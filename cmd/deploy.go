package cmd

import (
	"crypto/rsa"
	"fmt"
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

	kubernetesStableReleaseURL = "https://github.com/kubernetes/kubernetes/releases/download/v1.1.8/kubernetes.tar.gz"
	//kubernetesHyperkubeSpec    = "gcr.io/google_containers/hyperkube-amd64:v1.2.0-alpha.8"
	kubernetesHyperkubeSpec = "gcr.io/google_containers/hyperkube:v1.1.8"
)

var (
	deployArgNames = util.DeployArgNames
)

func NewDeployCmd() *cobra.Command {
	var deployCmd = &cobra.Command{
		Use:   "deploy",
		Short: "creates a new kubernetes cluster in Azure",
		Long:  deployLongDescription,
		Run: func(cmd *cobra.Command, args []string) {
			err := validateDeployArgs()
			if err != nil {
				log.Fatalf("Failed to validate arguments for `deploy` command: %q", err)
			}
			err = deployRun(cmd, args)
			if err != nil {
				log.Fatalf("Error occurred during deployment: %q", err)
			}
		},
	}

	flags := deployCmd.Flags()

	flags.String(deployArgNames.OutputDirectory, "", "output directory (this is derived from --deployment-name if omitted)")
	flags.String(deployArgNames.DeploymentName, "", "deployment identifier (used to name output, resource group, and other resources)")
	flags.String(deployArgNames.ResourceGroup, "", "resource group to deploy to (this is derived from --deployment-name if omitted)")
	flags.String(deployArgNames.Location, "brazilsouth", "location to deploy Azure resource (these can be found by running `azure location list` with azure-xplat-cli)")
	flags.String(deployArgNames.MasterSize, "Standard_A1", "size of the master virtual machine")
	flags.String(deployArgNames.NodeSize, "Standard_A1", "size of the node virtual machines")
	flags.Int(deployArgNames.NodeCount, 3, "initial number of node virtual machines")
	flags.String(deployArgNames.KubernetesReleaseURL, kubernetesStableReleaseURL, "url to kubernetes release tarball (not used yet)")
	flags.String(deployArgNames.KubernetesHyperkubeSpec, kubernetesHyperkubeSpec, "docker spec for hyperkube container to use")
	flags.String(deployArgNames.Username, "kube", "username to virtual machines")
	flags.String(deployArgNames.MasterFQDN, "", "fqdn for master (used for PKI). calculated from cloudapp dns for master's public ip") // tODO is this wired up?
	flags.StringSlice(deployArgNames.MasterExtraFQDNs, []string{}, "comma delimited list of SANs for the master")                      // tODO is this wired up?

	viper.BindPFlag(deployArgNames.OutputDirectory, flags.Lookup(deployArgNames.OutputDirectory))
	viper.BindPFlag(deployArgNames.DeploymentName, flags.Lookup(deployArgNames.DeploymentName))
	viper.BindPFlag(deployArgNames.ResourceGroup, flags.Lookup(deployArgNames.ResourceGroup))
	viper.BindPFlag(deployArgNames.Location, flags.Lookup(deployArgNames.Location))
	viper.BindPFlag(deployArgNames.MasterSize, flags.Lookup(deployArgNames.MasterSize))
	viper.BindPFlag(deployArgNames.NodeSize, flags.Lookup(deployArgNames.NodeSize))
	viper.BindPFlag(deployArgNames.NodeCount, flags.Lookup(deployArgNames.NodeCount))
	viper.BindPFlag(deployArgNames.KubernetesReleaseURL, flags.Lookup(deployArgNames.KubernetesReleaseURL))
	viper.BindPFlag(deployArgNames.KubernetesHyperkubeSpec, flags.Lookup(deployArgNames.KubernetesHyperkubeSpec))
	viper.BindPFlag(deployArgNames.Username, flags.Lookup(deployArgNames.Username))
	viper.BindPFlag(deployArgNames.MasterFQDN, flags.Lookup(deployArgNames.MasterFQDN))
	viper.BindPFlag(deployArgNames.MasterExtraFQDNs, flags.Lookup(deployArgNames.MasterExtraFQDNs))

	return deployCmd
}

func validateDeployArgs() error {
	validateRootArgs()

	// TODO: validate location + vmsizes, esp since used for masterfqdn

	if viper.GetString(deployArgNames.DeploymentName) == "" {
		viper.Set(deployArgNames.DeploymentName, fmt.Sprintf("kube-%s", time.Now().Format("20060102-150405")))
		log.Warnf("deployargs: --deployment-name is unset, generated a random deployment name: %q", viper.GetString(deployArgNames.DeploymentName))
	}

	if viper.GetString(deployArgNames.OutputDirectory) == "" {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatalf("unable to get working directory for output")
		}

		viper.Set(deployArgNames.OutputDirectory, path.Join(wd, "_deployments", viper.GetString(deployArgNames.DeploymentName)))
		log.Warnf("deployargs: --output-directory is unset, using this location: %q", viper.GetString(deployArgNames.OutputDirectory))

		err = os.MkdirAll(viper.GetString(deployArgNames.OutputDirectory), 0700)
		if err != nil {
			log.Fatalf("unable to create output directory for deployment: %q", err)
		}
	}

	if viper.GetString(deployArgNames.ResourceGroup) == "" {
		viper.Set(deployArgNames.ResourceGroup, viper.GetString(deployArgNames.DeploymentName))
		log.Warnf("deployargs: --resource-group is unset, derived one from --deployment-name: %q", viper.GetString(deployArgNames.ResourceGroup))
	}

	if viper.GetString(deployArgNames.MasterFQDN) == "" {
		viper.Set(deployArgNames.MasterFQDN, fmt.Sprintf("%s.%s.cloudapp.azure.com", viper.GetString(deployArgNames.DeploymentName), viper.GetString(deployArgNames.Location)))
		log.Warnf("deployargs: --master-fqdn is unset, derived from input: %q", viper.GetString(deployArgNames.MasterFQDN))
	}

	return nil
}

func deployRun(cmd *cobra.Command, args []string) error {
	d, err := util.NewDeployer()
	if err != nil {
		return err
	}

	var (
		appName, appURL, applicationID, servicePrincipalObjectID, servicePrincipalClientSecret string

		sshPrivateKey      *rsa.PrivateKey
		sshPublicKeyString string

		ca, apiserver, client *util.PkiKeyCertPair
	)

	pkiLock := stepPki(d, &ca, &apiserver, &client)
	sshLock := stepSsh(d, &sshPrivateKey, &sshPublicKeyString)

	rgLock := stepRg(d)
	if err = <-rgLock; err != nil {
		return err
	}

	adLock := stepAd(d, &applicationID, &servicePrincipalObjectID, &servicePrincipalClientSecret)
	if err = <-adLock; err != nil {
		return err
	}
	if err = <-pkiLock; err != nil {
		return err
	}
	if err = <-sshLock; err != nil {
		return err
	}

	_, _ = appName, appURL

	flavorArgs := deployToFlavorArgs(
		applicationID, servicePrincipalObjectID, servicePrincipalClientSecret,
		sshPrivateKey, sshPublicKeyString,
		ca, apiserver, client)

	deployLock := stepDeploy(d, flavorArgs)
	if err = <-deployLock; err != nil {
		return err
	}

	validateLock := stepValidate(flavorArgs)
	if err = <-validateLock; err != nil {
		return err
	}

	log.Infof("Deployment Complete!")
	log.Infof("master: %q", "https://"+viper.GetString(deployArgNames.MasterFQDN)+":6443")
	log.Infof("scripts: %q", viper.GetString(deployArgNames.OutputDirectory))

	return nil
}

func stepValidate(flavorArgs util.FlavorArguments) chan error {
	var c chan error = make(chan error)

	go func() {
		var err error
		defer func() {
			c <- err
		}()

		util.ValidateDeployment(flavorArgs)
	}()

	return c
}

func stepRg(d *util.Deployer) chan error {
	var c chan error = make(chan error)

	go func() {
		var err error
		defer func() {
			c <- err
		}()

		_, err = d.EnsureResourceGroup(
			viper.GetString(deployArgNames.ResourceGroup),
			viper.GetString(deployArgNames.Location),
			true)
		if err != nil {
			return
		}
	}()

	return c
}

func stepAd(d *util.Deployer,
	applicationID, servicePrincipalObjectID, servicePrincipalClientSecret *string) chan error {
	var c chan error = make(chan error)

	go func() {
		var err error
		defer func() {
			c <- err
		}()

		appName := viper.GetString(deployArgNames.DeploymentName)
		appURL := fmt.Sprintf("https://%s/", viper.GetString(deployArgNames.DeploymentName))
		*applicationID, *servicePrincipalObjectID, *servicePrincipalClientSecret, err =
			d.AdClient.CreateApp(appName, appURL)
		if err != nil {
			return
		}

		err = d.CreateRoleAssignment(viper.GetString(deployArgNames.ResourceGroup), *servicePrincipalObjectID)
		if err != nil {
			return
		}
	}()

	return c
}

func stepSsh(d *util.Deployer,
	sshPrivateKey **rsa.PrivateKey, sshPublicKeyString *string) chan error {
	var c chan error = make(chan error)

	go func() {
		var err error
		defer func() {
			c <- err
		}()
		*sshPrivateKey, *sshPublicKeyString, err = util.GenerateSsh(path.Join(viper.GetString(deployArgNames.OutputDirectory), "private.key"))
		if err != nil {
			return
		}

		privateKeyPem := util.PrivateKeyToPem(*sshPrivateKey)
		err = util.SaveDeploymentFile(fmt.Sprintf("%s_rsa", viper.GetString(deployArgNames.Username)), string(privateKeyPem), 0600)
		if err != nil {
			return
		}
	}()

	return c
}

func stepPki(d *util.Deployer,
	ca, apiserver, client **util.PkiKeyCertPair) chan error {
	var c chan error = make(chan error)

	go func() {
		var err error
		defer func() {
			c <- err
		}()
		*ca, *apiserver, *client, err =
			util.CreateKubeCertificates(viper.GetString(deployArgNames.MasterFQDN), viper.GetStringSlice(deployArgNames.MasterExtraFQDNs))
		if err != nil {
			err = fmt.Errorf("error occurred while creating kube certificates")
			return
		}

		err = util.SaveDeploymentFile("ca.key", (*ca).PrivateKeyPem, 0600)
		if err != nil {
			return
		}
		err = util.SaveDeploymentFile("ca.crt", (*ca).CertificatePem, 0600)
		if err != nil {
			return
		}
		err = util.SaveDeploymentFile("apiserver.key", (*apiserver).PrivateKeyPem, 0600)
		if err != nil {
			return
		}
		err = util.SaveDeploymentFile("apiserver.crt", (*apiserver).CertificatePem, 0600)
		if err != nil {
			return
		}
		err = util.SaveDeploymentFile("client.key", (*client).PrivateKeyPem, 0600)
		if err != nil {
			return
		}
		err = util.SaveDeploymentFile("client.crt", (*client).CertificatePem, 0600)
		if err != nil {
			return
		}
	}()
	return c
}

func deployToFlavorArgs(
	applicationID, servicePrincipalObjectID, servicePrincipalClientSecret string,
	sshPrivateKey *rsa.PrivateKey, sshPublicKeyString string,
	ca, apiserver, client *util.PkiKeyCertPair) util.FlavorArguments {
	flavorArgs := util.FlavorArguments{
		DeploymentName: viper.GetString(deployArgNames.DeploymentName),

		TenantID: viper.GetString(rootArgNames.TenantID),

		MasterSize:       viper.GetString(deployArgNames.MasterSize),
		NodeSize:         viper.GetString(deployArgNames.NodeSize),
		NodeCount:        viper.GetInt(deployArgNames.NodeCount),
		Username:         viper.GetString(deployArgNames.Username),
		SshPublicKeyData: sshPublicKeyString,

		KubernetesReleaseURL:    viper.GetString(deployArgNames.KubernetesReleaseURL),
		KubernetesHyperkubeSpec: viper.GetString(deployArgNames.KubernetesHyperkubeSpec),

		ServicePrincipalClientID:     applicationID,
		ServicePrincipalClientSecret: servicePrincipalClientSecret,

		MasterFQDN: viper.GetString(deployArgNames.MasterFQDN),

		CAKeyPair:        ca,
		ApiserverKeyPair: apiserver,
		ClientKeyPair:    client,
	}
	return flavorArgs
}

func stepDeploy(d *util.Deployer, flavorArgs util.FlavorArguments) chan error {
	var c chan error = make(chan error)

	go func() {
		var err error
		defer func() {
			c <- err
		}()

		template, parameters, err := util.ProduceTemplateAndParameters(flavorArgs)
		if err != nil {
			return
		}

		err = util.SaveDeploymentMap("azdeploy.json", template, 0600)
		if err != nil {
			return
		}
		err = util.SaveDeploymentMap("parameters.json", parameters, 0600)
		if err != nil {
			return
		}

		utilScript, err := util.ProduceUtilScript(flavorArgs)
		if err != nil {
			return
		}
		err = util.SaveDeploymentFile("util.sh", utilScript, 0700)
		if err != nil {
			return
		}

		_, err = d.DoDeployment(
			viper.GetString(deployArgNames.ResourceGroup),
			"myriad",
			template,
			parameters,
			true)
		if err != nil {
			return
		}

		return
	}()
	return c
}
