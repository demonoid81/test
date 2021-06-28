package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/sphera-erp/sphera/internal/models"
)

func (r *mutationResolver) NationalityMutation(ctx context.Context, nationality *models.Nationality) (*models.Nationality, error) {
	return r.nationalities.NationalityMutation(ctx, nationality)
}

func (r *queryResolver) Nationality(ctx context.Context, nationality *models.Nationality) (*models.Nationality, error) {
	return r.nationalities.Nationality(ctx, nationality)
}

func (r *queryResolver) Nationalities(ctx context.Context, nationality *models.Nationality, offset *int, limit *int) ([]*models.Nationality, error) {
	return r.nationalities.Nationalities(ctx, nationality)
}
