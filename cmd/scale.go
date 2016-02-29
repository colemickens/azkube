// +build ignore

package cmd

import (
	"log"
	"reflect"

	"github.com/colemickens/azkube/util"
	"github.com/spf13/cobra"
)

const (
	scaleDeploymentLongDescription = "long desc"
)

func NewScaleDeploymentCmd() *cobra.Command {
	var statePath string

	var scaleDeploymentCmd = &cobra.Command{
		Use:   "scale-deployment",
		Short: "scale a kubernetes deployment",
		Long:  scaleDeploymentLongDescription,
		Run: func(cmd *cobra.Command, args []string) {
			log.Println("starting scale-deployment command")

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

			state = RunScaleDeploymentCmd(state)
			err = WriteState(statePath, state)
			if err != nil {
				panic(err)
			}

			log.Println("finished scale-deployment command")
		},
	}

	scaleDeploymentCmd.Flags().StringVarP(&statePath, "state", "s", "./state.json", "path to load state from, and to persist state into")

	return scaleDeploymentCmd
}

func RunScaleDeploymentCmd(stateIn *util.State) (stateOut *util.State) {
	*stateOut = *stateIn

	return stateOut
}
