package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"time"

	"github.com/sphera-erp/sphera/internal/models"
)

func (r *mutationResolver) MovementMutation(ctx context.Context, movement *models.Movement) (*models.Movement, error) {
	return r.accounting.MovementMutation(ctx, movement)
}

func (r *mutationResolver) AddToBalance(ctx context.Context, organization models.Organization, amount float64) (bool, error) {
	return r.accounting.AddToBalance(ctx, organization, amount)
}

func (r *queryResolver) Movement(ctx context.Context, movement *models.Movement) (*models.Movement, error) {
	return r.accounting.Movement(ctx, movement)
}

func (r *queryResolver) Movements(ctx context.Context, movement *models.Movement, filter *models.MovementFilter, offset *int, limit *int) ([]*models.Movement, error) {
	return r.accounting.Movements(ctx, movement, filter, offset, limit)
}

func (r *queryResolver) FlowBalance(ctx context.Context, organization *models.Organization, from *time.Time, to *time.Time) ([]*models.Balance, error) {
	return r.accounting.FlowBalance(ctx, organization, from, to)
}

func (r *queryResolver) GetBalance(ctx context.Context, organization models.Organization, until *time.Time) (*float64, error) {
	return r.accounting.GetBalance(ctx, organization, until)
}

func (r *queryResolver) Statistics(ctx context.Context, organization *models.Organization) (*models.Stat, error) {
	return r.accounting.Statistics(ctx, organization)
}
