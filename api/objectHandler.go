package api

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v7"
	"github.com/sphera-erp/sphera/app"
	"log"
	"mime"
	"net/http"
	"time"
)

func getObjectHandler(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		bucket, _ := vars["bucket"]
		objectName, _ := vars["object"]

		// New SSE-C where the cryptographic key is derived from a password and the objectname + bucketname as salt
		//encryption := encrypt.DefaultPBKDF([]byte(app.Cfg.Api.S3Key), []byte(bucket+objectName))

		// Get the encrypted object
		reader, err := app.S3.GetObject(context.Background(), bucket, objectName, minio.GetObjectOptions{})
		if err != nil {
			log.Fatalln(err)
		}
		defer reader.Close()
		cd := mime.FormatMediaType("attachment", map[string]string{"filename": objectName})
		w.Header().Set("Content-Disposition", cd)
		w.Header().Set("Content-Type", "application/octet-stream")
		http.ServeContent(w, r, objectName, time.Now(), reader)
	}
}
