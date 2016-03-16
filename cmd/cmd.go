package cmd

import (
	"strings"

	"github.com/colemickens/azkube/util"

	"github.com/Azure/go-autorest/autorest/azure"
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	rootName             = "azkube"
	rootShortDescription = "A Kubernetes deployment helper for Azure"
)

type RootArguments struct {
	Debug           bool
	SubscriptionID  string
	AuthMethod      string
	ClientID        string
	ClientSecret    string
	CertificatePath string
	PrivateKeyPath  string
}

func NewRootCmd() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   rootName,
		Short: rootShortDescription,
	}

	pflags := rootCmd.PersistentFlags()
	pflags.Bool("debug", false, "debug mode, outputs more logging")
	pflags.String("subscription-id", "", "azure subscription id")
	pflags.String("auth-method", "device", "auth method (default:`device`, `client_secret`, `client_certificate`)")
	pflags.String("client-id", "", "client id (used with --auth-method=[client_secret|client_certificate])")
	pflags.String("client-secret", "", "client secret (used with --auth-mode=client_secret)")
	pflags.String("certificate-path", "", "path to client certificate (used with --auth-method=client_certificate)")
	pflags.String("private-key-path", "", "path to private key (used with --auth-method=client_certificate)")

	viper.SetEnvPrefix("azkube")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.BindPFlag("debug", pflags.Lookup("debug"))
	viper.BindPFlag("subscription-id", pflags.Lookup("subscription-id"))
	viper.BindPFlag("auth-method", pflags.Lookup("auth-method"))
	viper.BindPFlag("client-id", pflags.Lookup("client-id"))
	viper.BindPFlag("client-secret", pflags.Lookup("client-secret"))
	viper.BindPFlag("certificate-path", pflags.Lookup("certificate-path"))
	viper.BindPFlag("private-key-path", pflags.Lookup("private-key-path"))

	rootCmd.AddCommand(NewDeployCmd())
	rootCmd.AddCommand(NewScaleDeploymentCmd())
	rootCmd.AddCommand(NewDestroyDeploymentCmd())

	return rootCmd
}

func parseRootArgs(cmd *cobra.Command, args []string) RootArguments {
	rootArgs := RootArguments{
		Debug:           viper.GetBool("debug"),
		SubscriptionID:  viper.GetString("subscription-id"),
		AuthMethod:      viper.GetString("auth-method"),
		ClientID:        viper.GetString("client-id"),
		ClientSecret:    viper.GetString("client-secret"),
		CertificatePath: viper.GetString("certificate-path"),
		PrivateKeyPath:  viper.GetString("private-key-path"),
	}

	if rootArgs.SubscriptionID == "" {
		log.Fatal("--subscription-id must be specified")
	}

	if rootArgs.AuthMethod == "client_secret" {
		if rootArgs.ClientID == "" || rootArgs.ClientSecret == "" {
			log.Fatal("--client-id and --client-secret must be specified when --auth-method=\"client_secret\".")
		}
	} else if rootArgs.AuthMethod == "client_certificate" {
		if rootArgs.ClientID == "" || rootArgs.CertificatePath == "" || rootArgs.PrivateKeyPath == "" {
			log.Fatal("--client-id and --certificate-path, and --private-key-path must be specified when --auth-method=\"client_certificate\".")
		}
	}

	if rootArgs.Debug {
		log.SetLevel(log.DebugLevel)
	}

	return rootArgs
}

func getClient(rootArgs RootArguments) (*util.AzureClient, error) {
	azureEnvironment := azure.PublicCloud
	tenantID, err := util.GetTenantID(azureEnvironment, rootArgs.SubscriptionID)
	if err != nil {
		return nil, err
	}

	switch rootArgs.AuthMethod {
	case "device":
		return util.NewClientWithDeviceAuth(azureEnvironment, rootArgs.SubscriptionID, tenantID)
	case "client_secret":
		return util.NewClientWithClientSecret(azureEnvironment, rootArgs.SubscriptionID, tenantID, rootArgs.ClientID, rootArgs.ClientSecret)
	case "client_certificate":
		return util.NewClientWithClientCertificate(azureEnvironment, rootArgs.SubscriptionID, tenantID, rootArgs.ClientID, rootArgs.CertificatePath, rootArgs.PrivateKeyPath)
	default:
		log.Fatalf("--auth-method: ERROR: method unsupported. method=%q.", rootArgs.AuthMethod)
	}

	return nil, nil // unreachable
}
