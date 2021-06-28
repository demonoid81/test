package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/sphera-erp/sphera/internal/models"
)

func (r *mutationResolver) PassportMutation(ctx context.Context, passport *models.Passport) (*models.Passport, error) {
	return r.passports.PassportMutation(ctx, passport)
}

func (r *queryResolver) Passport(ctx context.Context, passport models.Passport) (*models.Passport, error) {
	return r.passports.Passport(ctx, passport)
}

func (r *queryResolver) Passports(ctx context.Context, passport *models.Passport, filter *models.PassportFilter, sort []models.PassportSort, offset *int, limit *int) ([]*models.Passport, error) {
	return r.passports.Passports(ctx, passport, filter, sort, offset, limit)
}
