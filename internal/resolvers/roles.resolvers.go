package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/sphera-erp/sphera/internal/models"
)

func (r *mutationResolver) RoleMutation(ctx context.Context, role *models.Role) (*models.Role, error) {
	return r.roles.RoleMutation(ctx, role)
}

func (r *queryResolver) Role(ctx context.Context, role *models.Role) (*models.Role, error) {
	return r.roles.Role(ctx, role)
}

func (r *queryResolver) Roles(ctx context.Context, role *models.Role, offset *int, limit *int) ([]*models.Role, error) {
	return r.roles.Roles(ctx, role, offset, limit)
}
