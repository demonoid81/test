package app

import (
	consulapi "github.com/hashicorp/consul/api"
	"github.com/jmoiron/sqlx"
	"github.com/minio/minio-go/v7"
	"github.com/rs/zerolog"
	"github.com/sphera-erp/sphera/pkg/pglx"
	"github.com/cskr/pubsub"
)

type App struct {
	Cfg        *Config
	Logger     *zerolog.Logger
	Cockroach  *pglx.DB
	S3         *minio.Client
	Clickhouse *sqlx.DB
	Consul     *consulapi.Client
	PubSub     *pubsub.PubSub
}
