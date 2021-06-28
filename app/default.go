package app

import "time"

var DefaultCfg = Config{
	MigrationsPath: "./migrations",
	DB: CockroachCfg{
		Host:     "192.168.10.244",
		Port:     26257,
		Database: "wfmt",
	},
	Logs: Logs{
		Level: "debug",
	},

	Api: Api{
		AllowedOrigins: []string{"*"},
		Address:        "0.0.0.0",
		Port:           8989,
		AccessSecret:   "go_very_goog_lang",
		S3Key:          "123456",
		DadataSecret:   "456d34e1358095a2a23e11b3f4ac3542ba13e5da",
		DadataApi:      "e55d0d43fb4f455b8e61ba333285f29cf617231f",
		MasterToken:    "bem3lUAJDfSpmERuyNDjuwL0wNcH25pDn0cG5zEBFlrwUwphmTR5x1XpQEDhL7LosjFNJxItLj97zwYUMNc2NvLBIWheWVKNJFd53INWZo32i5C59KusQFPhcQcZ250b",
		FnsTempToken: FnsTempToken{
			Expire: time.Now(),
		},
	},
	Consul: Consul{
		Endpoint: "",
	},
	S3: S3Cfg{
		Endpoint:        "minio:9000",
		AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
		SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		Location:        "us-west-1",
		UseSSL:          false,
	},
}
