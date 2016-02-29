package cmd

import (
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/colemickens/azkube/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	rootName             = "azkube"
	rootShortDescription = "A Kubernetes deployment helper for Azure"
)

var (
	rootArgNames util.RootArguments = util.RootArgNames
)

func NewRootCmd() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   rootName,
		Short: rootShortDescription,
	}

	pflags := rootCmd.PersistentFlags()
	pflags.String(rootArgNames.TenantID, "", "azure tenant id")
	pflags.String(rootArgNames.SubscriptionID, "", "azure subscription id")
	pflags.String(rootArgNames.AuthMethod, "device", "auth method (default:`device`, `client_secret`, `client_certificate`)")
	pflags.String(rootArgNames.ClientID, "", "client id (used with --auth-method=[client_secret|client_certificate])")
	pflags.String(rootArgNames.ClientSecret, "", "client secret (used with --auth-mode=clientsecret)")
	pflags.String(rootArgNames.CertificatePath, "", "path to client certificate (used with --auth-method=client_certificate)")
	pflags.String(rootArgNames.PrivateKeyPath, "", "path to private key (used with --auth-method=client_certificate)")

	viper.SetEnvPrefix("azkube")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.BindPFlag(rootArgNames.TenantID, pflags.Lookup(rootArgNames.TenantID))
	viper.BindPFlag(rootArgNames.SubscriptionID, pflags.Lookup(rootArgNames.SubscriptionID))
	viper.BindPFlag(rootArgNames.AuthMethod, pflags.Lookup(rootArgNames.AuthMethod))
	viper.BindPFlag(rootArgNames.ClientID, pflags.Lookup(rootArgNames.ClientID))
	viper.BindPFlag(rootArgNames.ClientSecret, pflags.Lookup(rootArgNames.ClientSecret))
	viper.BindPFlag(rootArgNames.CertificatePath, pflags.Lookup(rootArgNames.CertificatePath))
	viper.BindPFlag(rootArgNames.PrivateKeyPath, pflags.Lookup(rootArgNames.PrivateKeyPath))

	rootCmd.AddCommand(NewDeployCmd())
	//rootCmd.AddCommand(NewScaleDeploymentCmd())
	//rootCmd.AddCommand(NewDestroyDeploymentCmd())

	return rootCmd
}

func validateRootArgs() {
	if viper.GetString(util.RootArgNames.SubscriptionID) == "" {
		log.Fatal("--subscription-id must be specified")
	}

	if viper.GetString(rootArgNames.TenantID) == "" {
		log.Fatal("--tenant-id must be specified")
	}

	if rootArgNames.AuthMethod == "client_secret" {
		if viper.GetString(rootArgNames.ClientID) == "" || viper.GetString(rootArgNames.ClientSecret) == "" {
			log.Fatal("--client-id and --client-secret must be specified when --auth-method=\"client_secret\"")
		}
	} else if rootArgNames.AuthMethod == "client_certificate" {
		if viper.GetString(rootArgNames.ClientID) == "" || viper.GetString(rootArgNames.CertificatePath) == "" || viper.GetString(rootArgNames.PrivateKeyPath) == "" {
			log.Fatal("--client-id and --certificate-path, and --private-key-path must be specified when --auth-method=\"client_certificate\"")
		}
	}
}
