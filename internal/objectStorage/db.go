package objectStorage

import (
	"context"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/sphera-erp/sphera/app"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
)

func PutObjectInDB(ctx context.Context, app *app.App, db pglxqb.BaseRunner, file minio.UploadInfo, uuidObject uuid.UUID) error {
	_, err := pglxqb.Insert("content").
		Columns("uuid", "bucket").
		Values(uuidObject, file.Bucket).
		RunWith(db).
		Exec(ctx)
	return err
}
