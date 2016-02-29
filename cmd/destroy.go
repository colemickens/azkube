// +build ignore

package cmd

import (
	"log"
	"reflect"

	"github.com/colemickens/azkube/util"
	"github.com/spf13/cobra"
)

const (
	destroyDeploymentLongDescription = "long desc"
)

func NewDestroyDeploymentCmd() *cobra.Command {
	var statePath string

	var destroyDeploymentCmd = &cobra.Command{
		Use:   "destroy-deployment",
		Short: "destroy a kubernetes deployment",
		Long:  destroyDeploymentLongDescription,
		Run: func(cmd *cobra.Command, args []string) {
			log.Println("starting destroy-deployment command")

			state := &util.State{}
			var err error
			state, err = ReadAndValidateState(
				statePath,
				[]reflect.Type{
					reflect.TypeOf(state.Common),
					reflect.TypeOf(state.App),
					reflect.TypeOf(state.Ssh),
					reflect.TypeOf(state.Pki),
					reflect.TypeOf(state.Vault),
					reflect.TypeOf(state.Secrets),
				},
				[]reflect.Type{
					reflect.TypeOf(state.Myriad),
				},
			)
			if err != nil {
				panic(err)
			}

			state = RunDestroyDeploymentCmd(state)

			err = WriteState(statePath, state)
			if err != nil {
				panic(err)
			}

			log.Println("finished destroy-deployment command")
		},
	}

	destroyDeploymentCmd.Flags().StringVarP(&statePath, "state", "s", "./state.json", "path to load state from, and to persist state into")

	return destroyDeploymentCmd
}

func RunDestroyDeploymentCmd(stateIn *util.State) (stateOut *util.State) {
	*stateOut = *stateIn

	return stateOut
}
