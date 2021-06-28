package cli

import (
	"fmt"
	"os"
	"strings"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/sphera-erp/sphera/app"
)

var (
	cfgFile         string
	consulLastIndex uint64 = 0

	rootCmd = &cobra.Command{
		Use:     "rest-service-example",
		Version: "v1.0",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {

			// заполним переменные
			return initializeConfig(cmd)
		},
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
	App *app.App
)

func init() {
	App = app.NewApp()
	rootCmd.AddCommand(startServer)
	// rootCmd.PersistentFlags().StringVarP(&Cfg.DB.Host, "db-host", "h", "localhost", "Cockroach database host, default: localhost")
}

// Execute is
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func initializeConfig(cmd *cobra.Command) error {

	v := viper.New()
	if cfgFile != "" { // enable ability to specify config file via flag
		v.SetConfigFile(cfgFile)
	}

	v.SetConfigName(".env") // name of config file (without extension)
	v.AddConfigPath(".")    // adding home directory as first search path
	v.AutomaticEnv()        // read in environment variables that match

	// If a config file is found, read it in.
	if err := v.ReadInConfig(); err == nil {
		v.WatchConfig()
		fmt.Println("Using config file:", v.ConfigFileUsed())
	}

	v.SetEnvKeyReplacer(strings.NewReplacer("_", "."))

	// Bind the current command's persistent flags and flags  to viper
	bindPersistentFlags(cmd, v)
	bindFlags(cmd, v)

	//if App.Cfg.Consul.Endpoint != "" {
	//	loadConsulConfig()
	//}

	fmt.Println(App.Cfg)

	return nil
}

func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Environment variables can't have dashes in them, so bind them to their equivalent
		// keys with underscores, e.g. --favorite.color to STING_FAVORITE_COLOR
		if strings.Contains(f.Name, "-") {
			envVarSuffix := strings.ToUpper(strings.ReplaceAll(f.Name, ".", "_"))
			v.BindEnv(f.Name, envVarSuffix)
		}

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}

func bindPersistentFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		// Environment variables can't have dashes in them, so bind them to their equivalent
		// keys with underscores, e.g. --favorite.color to STING_FAVORITE_COLOR
		if strings.Contains(f.Name, "-") {
			envVarSuffix := strings.ToUpper(strings.ReplaceAll(f.Name, ".", "_"))
			v.BindEnv(f.Name, envVarSuffix)
		}

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}

func loadConsulConfig() {
	var err error

	App.Consul, err = App.InitConsul()

	qo := &consulapi.QueryOptions{
		WaitIndex: consulLastIndex,
	}

	kvPairs, qm, err := App.Consul.KV().List("/api", qo)
	if err != nil {

	}

	fmt.Println("remote ConsulLastIndex", qm.LastIndex)

	if consulLastIndex == qm.LastIndex {
		fmt.Println("ConsulLastIndex not changed")
	}

	// newConfig := make(map[string]string)

	for idx, item := range kvPairs {
		fmt.Printf("item[%d] %#v\n", idx, item)
	}
}
