package app

import (
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"log"
)

func (app *App) InitS3() *minio.Client {
	var err error
	s3, err := minio.New(app.Cfg.S3.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(app.Cfg.S3.AccessKeyID, app.Cfg.S3.SecretAccessKey, ""),
		Secure: app.Cfg.S3.UseSSL,
	})
	if err != nil {
		log.Fatalln(err)
	}
	return s3
}
