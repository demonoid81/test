package app

import (
	"sync"
	"time"
)

type Config struct {
	NodeName       string
	UseTracer      bool
	MigrationsPath string
	TarantoolURL   string
	DB             CockroachCfg
	S3             S3Cfg
	Clickhouse     ClickhouseCfg
	Jaeger         JaegerCfg
	Logs           Logs
	Api            Api
	Consul         Consul
}

type Url struct {
	Host string
	Port int
}

type Tags struct {
	Name  string
	Value string
}

type Logs struct {
	Level string
}

type Api struct {
	AllowedOrigins      []string
	Address             string
	Port                int
	AccessSecret        string
	S3Key               string
	S3PersonsBucketName string
	DadataSecret        string
	DadataApi           string
	Gorush              string
	MasterToken         string
	FnsTempToken        FnsTempToken
}

type FnsTempToken struct {
	sync.RWMutex
	Token  string
	Expire time.Time
}

type Consul struct {
	Endpoint string
}

type CockroachCfg struct {
	Username        string
	Password        string
	Host            string
	Port            int
	Database        string
	DirectoryPath   string
	ApplicationName string
	SslMode         string
	SslRootCert     string
	SslCert         string
	SslKey          string
}

type ClickhouseCfg struct {
	Host                   string
	Port                   int
	Username               string
	Password               string
	Debug                  bool
	ReadTimeout            int
	WriteTimeout           int
	NoDelay                bool
	AltHosts               []Url
	ConnectionOpenStrategy string
	BlockSize              int64
	PoolSize               int
	Compress               int
	Secure                 bool
	SkipVerify             bool
	TlsConfig              string
}

type JaegerCfg struct {
	Endpoint string
	Username string
	Password string
	Tags     []Tags
}

type S3Cfg struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Location        string
	UseSSL          bool
}
