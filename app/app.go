package app

func NewApp() *App {
	app := &App{
		Cfg: &DefaultCfg,
	}

	return app
}
