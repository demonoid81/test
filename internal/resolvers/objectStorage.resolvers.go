package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
)

func (r *mutationResolver) SingleUpload(ctx context.Context, file graphql.Upload, bucket string) (uuid.UUID, error) {
	return r.Resolver.objectStorage.SingleUpload(ctx, file, bucket)
}

func (r *mutationResolver) MultipleUpload(ctx context.Context, files []graphql.Upload, bucket string) ([]uuid.UUID, error) {
	return r.Resolver.objectStorage.MultipleUpload(ctx, files, bucket)
}
