package objectStorage

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/sphera-erp/sphera/app"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.opentelemetry.io/otel"
	"log"
)

type Resolver struct {
	env *app.App
	Resolvers
}

type Resolvers interface {
	SingleUpload(ctx context.Context, file graphql.Upload, bucket string) (uuid.UUID, error)
	MultipleUpload(ctx context.Context, files []graphql.Upload, bucket string) ([]uuid.UUID, error)
}

func NewObjectStorageResolvers(app *app.App) (*Resolver, error) {
	return &Resolver{
		env: app,
	}, nil
}

func (r *Resolver) SingleUpload(ctx context.Context, file graphql.Upload, bucket string) (uuid.UUID, error) {
	if r.env.Cfg.UseTracer {
		tr := otel.Tracer("SingleUpload")
		_, span := tr.Start(ctx, "SingleUpload")
		defer span.End()
	}
	uuidObject := uuid.New()
	//encryption := encrypt.DefaultPBKDF([]byte(r.env.Cfg.Api.S3Key), []byte(bucket + uuidObject.String()))
	//minio.PutObjectOptions{ServerSideEncryption: encryption}
	n, err := r.env.S3.PutObject(ctx, bucket, uuidObject.String(), file.File, file.Size, minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		r.env.Logger.Error().Str("module", "objectStorage").Str("function", "SingleUpload").Err(err).Msg("Error uploading file")
		return uuid.Nil, gqlerror.Errorf("Error uploading file: %s", err)
	}
	r.env.Logger.Debug().Str("module", "objectStorage").Str("function", "SingleUpload").Msgf("Uploaded %s of size: %d Successfully.", file.Filename, n.Size)
	tx, err := r.env.Cockroach.BeginX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "persons").Str("func", "PersonaMutation").Err(err).Msg("Error run transaction")
		return uuid.Nil, gqlerror.Errorf("Error run transaction")
	}
	defer tx.Rollback(ctx)
	err = PutObjectInDB(ctx, r.env, tx, n, uuidObject)
	if err != nil {
		// удалим мертвый объект так как нанего нет ссылки
		opts := minio.RemoveObjectOptions{
			GovernanceBypass: true,
		}

		err = r.env.S3.RemoveObject(ctx, bucket, file.Filename, opts)
		if err != nil {
			log.Fatalln(err)
		}
	}
	err = tx.Commit(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "persons").Str("func", "PersonaMutation").Err(err).Msg("Error commit transaction")
		return uuid.Nil, gqlerror.Errorf("Error commit transaction")
	}
	return uuidObject, nil
}

func (r *Resolver) MultipleUpload(ctx context.Context, files []graphql.Upload, bucket string) ([]uuid.UUID, error) {
	if r.env.Cfg.UseTracer {
		tr := otel.Tracer("SingleUpload")
		_, span := tr.Start(ctx, "SingleUpload")
		defer span.End()
	}
	tx, err := r.env.Cockroach.BeginX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "persons").Str("func", "PersonaMutation").Err(err).Msg("Error run transaction")
		return nil, gqlerror.Errorf("Error run transaction")
	}
	defer tx.Rollback(ctx)
	var uuidObjects []uuid.UUID
	for _, file := range files {
		uuidObject := uuid.New()
		// New SSE-C where the cryptographic key is derived from a password and the objectname + bucketname as salt
		//encryption := encrypt.DefaultPBKDF([]byte(r.env.Cfg.Api.S3Key), []byte(bucket + uuidObject.String()))

		// Encrypt file content and upload to the server
		n, err := r.env.S3.PutObject(ctx, bucket, uuidObject.String(), file.File, file.Size, minio.PutObjectOptions{ContentType: "application/octet-stream"})
		if err != nil {
			r.env.Logger.Error().Str("module", "objectStorage").Str("function", "SingleUpload").Err(err).Msg("Error uploading file")
			return nil, gqlerror.Errorf("Error uploading file: %s", err)
		}
		r.env.Logger.Debug().Str("module", "objectStorage").Str("function", "SingleUpload").Msgf("Uploaded %s of size: %d Successfully.", file.Filename, n.Size)

		err = PutObjectInDB(ctx, r.env, tx, n, uuidObject)
		if err != nil {
			// удалим мертвый объект так как нанего нет ссылки
			opts := minio.RemoveObjectOptions{
				GovernanceBypass: true,
			}

			err = r.env.S3.RemoveObject(ctx, bucket, file.Filename, opts)
			if err != nil {
				log.Fatalln(err)
			}
		}
		uuidObjects = append(uuidObjects, uuidObject)
	}
	err = tx.Commit(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "persons").Str("func", "PersonaMutation").Err(err).Msg("Error commit transaction")
		return nil, gqlerror.Errorf("Error commit transaction")
	}
	return uuidObjects, nil
}
