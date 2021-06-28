package cli

import (
	"context"
	"fmt"
	"github.com/cskr/pubsub"
	"github.com/spf13/cobra"
	"github.com/sphera-erp/sphera/api"
	"time"
)

var startServer = &cobra.Command{
	Use:   "server",
	Short: "Run Server",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initializeConfig(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := server(); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	fmt.Println(&App.Cfg)
	// App config
	startServer.Flags().StringVar(&App.Cfg.NodeName, "nodeName", "localhost", "Application instance name.")
	startServer.Flags().BoolVar(&App.Cfg.UseTracer, "useTracer", false, "User trace to watch.")
	// CockroachDB
	startServer.Flags().StringVar(&App.Cfg.DB.Host, "db.host", "localhost", "The host name or address of a CockroachDB node or load balancer.")
	startServer.Flags().IntVar(&App.Cfg.DB.Port, "db.port", 26257, "The port number of the SQL interface of the CockroachDB node or load balancer.")
	startServer.Flags().StringVar(&App.Cfg.DB.Username, "db.username", "", "The SQL user that will own the client session.")
	startServer.Flags().StringVar(&App.Cfg.DB.Password, "db.password", "", "The user's password. ")
	startServer.Flags().StringVar(&App.Cfg.DB.Database, "db.database", "", "A database name to use as current database.")
	startServer.Flags().StringVar(&App.Cfg.DB.DirectoryPath, "db.unixPath", "", "The directory path to the client listening for a socket connection.")
	startServer.Flags().StringVar(&App.Cfg.DB.ApplicationName, "db.applicationName", "", "The current application name for statistics collection.")
	startServer.Flags().StringVar(&App.Cfg.DB.SslMode, "db.sslmode", "", "Which type of secure connection to use: disable, allow, prefer, require, verify-ca or verify-full")
	startServer.Flags().StringVar(&App.Cfg.DB.SslRootCert, "db.sslrootcert", "", "Path to the CA certificate, when sslmode is not disable.")
	startServer.Flags().StringVar(&App.Cfg.DB.SslCert, "db.sslcert", "", "Path to the client certificate, when sslmode is not disable.")
	startServer.Flags().StringVar(&App.Cfg.DB.SslKey, "db.sslkey", "", "Path to the client private key, when sslmode is not disable.")
	// Tarantool
	startServer.Flags().StringVar(&App.Cfg.TarantoolURL, "tarantoolUrl", "tarantool", "The address of a Tarantool node or load balancer.")
	//
	startServer.Flags().StringVar(&App.Cfg.Jaeger.Endpoint, "jaeger.endpoint", "http://localhost:14268/api/traces", "The HTTP endpoint for sending spans directly to a collector.")
	startServer.Flags().StringVar(&App.Cfg.Jaeger.Username, "jaeger.username", "", "Username to send as part of \"Basic\" authentication to the collector endpoint")
	startServer.Flags().StringVar(&App.Cfg.Jaeger.Password, "jaeger.password", "", "Password to send as part of \"Basic\" authentication to the collector endpoin.")

	startServer.Flags().StringVar(&App.Cfg.Consul.Endpoint, "consul.endpoint", "", "Cockroach database host")
}

func server() error {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	App.Logger = App.NewLogger()

	App.Logger.Info().Msgf("Starting server")

	//  инициализируем базу данных
	db, err := App.InitCockroachDB(ctx)
	if err != nil {
		cancel()
		return err
	}
	//defer App.CloseCockroachDB()
	App.Cockroach = db
	// err = App.MigrateDatabase(ctx, App.Cockroach)
	// if err != nil {
	// 	cancel()
	// 	return err
	// }

	App.PubSub = pubsub.New(0)

	App.S3 = App.InitS3()

	api.Api(ctx, cancel, App)
	return nil
}
