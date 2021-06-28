package app

import consulapi "github.com/hashicorp/consul/api"

func (app *App) InitConsul() (*consulapi.Client, error) {

	config := consulapi.DefaultConfig()
	config.Address = app.Cfg.Consul.Endpoint

	consul, err := consulapi.NewClient(config)
	if err != nil {

	}
	return consul, nil
}
