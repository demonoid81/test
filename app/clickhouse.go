package app

import (
	"context"
	"crypto/tls"
	"fmt"
	"strconv"

	"github.com/ClickHouse/clickhouse-go"
	_ "github.com/ClickHouse/clickhouse-go"
	"github.com/jmoiron/sqlx"
)

// InitClickhouse is
func (app *App) InitClickhouse(ctx context.Context) (*sqlx.DB, error) {

	clickhouse.RegisterTLSConfig("tls", &tls.Config{
		//      RootCAs: rootCertPool,
		//      Certificates: clientCert,
	})

	db, err := sqlx.Open("clickhouse", clickhouseURL(app.Cfg.Clickhouse))
	if err != nil {
		app.Logger.Fatal().Err(err).Msg("Unable connect to clickhouse")
	}

	return db, nil
}

func clickhouseURL(cfg ClickhouseCfg) string {
	// tcp://host1:9000?username=user&password=qwerty&database=clicks&read_timeout=10&write_timeout=20&alt_hosts=host2:9000,host3:9000
	var options = ""
	// Username string
	if cfg.Username != "" {
		options += "username=" + cfg.Username
	}
	// Password string
	if cfg.Password != "" {
		if options != "" || options[len(options)-1:] != "&" {
			options += "&"
		}
		options += "password=" + cfg.Password
	}
	// Debug bool
	if cfg.Debug {
		if options != "" || options[len(options)-1:] != "&" {
			options += "&"
		}
		options += "debug=" + strconv.FormatBool(cfg.Debug)
	}
	// ReadTimeout int
	if cfg.ReadTimeout != 0 {
		if options != "" || options[len(options)-1:] != "&" {
			options += "&"
		}
		options += fmt.Sprintf("read_timeout=", cfg.ReadTimeout)
	}
	// WriteTimeout int
	if cfg.WriteTimeout != 0 {
		if options != "" || options[len(options)-1:] != "&" {
			options += "&"
		}
		options += fmt.Sprintf("write_timeout=%d", cfg.WriteTimeout)
	}
	// NoDelay bool
	if cfg.NoDelay {
		if options != "" || options[len(options)-1:] != "&" {
			options += "&"
		}
		options += "no_delay=" + strconv.FormatBool(cfg.NoDelay)
	}
	// ConnectionOpenStrategy string
	if cfg.ConnectionOpenStrategy != "" {
		if options != "" || options[len(options)-1:] != "&" {
			options += "&"
		}
		options += "connection_open_strategy=" + cfg.ConnectionOpenStrategy
	}
	// BlockSize              int64
	if cfg.BlockSize != 0 {
		if options != "" || options[len(options)-1:] != "&" {
			options += "&"
		}
		options += fmt.Sprintf("block_size=%d", cfg.BlockSize)
	}
	// PoolSize               int
	if cfg.PoolSize != 0 {
		if options != "" || options[len(options)-1:] != "&" {
			options += "&"
		}
		options += fmt.Sprintf("pool_size=%d", cfg.PoolSize)
	}
	// Compress               int
	if cfg.Compress != 0 {
		if options != "" || options[len(options)-1:] != "&" {
			options += "&"
		}
		options += fmt.Sprintf("compress=%d", cfg.Compress)
	}
	// Secure                 bool
	if cfg.Secure {
		if options != "" || options[len(options)-1:] != "&" {
			options += "&"
		}
		options += "secure=" + strconv.FormatBool(cfg.Secure)
	}
	// SkipVerify             bool
	if cfg.SkipVerify {
		if options != "" || options[len(options)-1:] != "&" {
			options += "&"
		}
		options += "skip_verify=" + strconv.FormatBool(cfg.Debug)
	}
	// TlsConfig              string
	// AltHosts               []string
	if len(cfg.AltHosts) > 0 {
		if options != "" || options[len(options)-1:] != "&" {
			options += "&"
		}
		var hosts string
		for i, item := range cfg.AltHosts {
			if i > 0 {
				hosts += ","
			}
			hosts += fmt.Sprintf("%s:%d", item.Host, item.Port)
		}
		options += "alt_hosts=" + hosts
	}
	url := fmt.Sprintf("tcp://%s:%d", cfg.Host, cfg.Port)

	if len(options) > 0 {
		url += "?" + options
	}
	return url
}
