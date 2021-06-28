package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sphera-erp/sphera/pkg/nalogSoap/platformRegistration"
)

var joinFns = &cobra.Command{
	Use:   "join",
	Short: "join in fns",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initializeConfig(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := join(); err != nil {
			return err
		}
		return nil
	},
}

func join() error {

	App.Logger = App.NewLogger()

	App.Logger.Info().Msgf("Starting server")

	err := platformRegistration.PlatformRegistration(App)

	fmt.Println(err)

	return nil
}
