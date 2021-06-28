package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/sphera-erp/sphera/internal/models"
)

func (r *mutationResolver) Validate(ctx context.Context, pincode string) (*string, error) {
	return r.users.Validate(ctx, pincode)
}

func (r *mutationResolver) UserMutation(ctx context.Context, user *models.User) (*models.User, error) {
	return r.users.UserMutation(ctx, user)
}

func (r *mutationResolver) ResetUser(ctx context.Context, phone *string) (bool, error) {
	return r.users.ResetUser(ctx, phone)
}

func (r *mutationResolver) UpdateToken(ctx context.Context, token string) (bool, error) {
	return r.users.UpdateToken(ctx, token)
}

func (r *queryResolver) AuthUserByPhone(ctx context.Context, phone string, client *models.ClientType) (*string, error) {
	return r.users.AuthUserByPhone(ctx, phone, client)
}

func (r *queryResolver) RegUserByPhone(ctx context.Context, phone string) (*string, error) {
	return r.users.RegUserByPhone(ctx, phone)
}

func (r *queryResolver) GetCurrentUser(ctx context.Context) (*models.User, error) {
	return r.users.GetCurrentUser(ctx)
}

func (r *queryResolver) User(ctx context.Context, user *models.User) (*models.User, error) {
	return r.users.User(ctx, user)
}

func (r *queryResolver) Users(ctx context.Context, user *models.User, filter *models.UserFilter, sort []models.UserSort, offset *int, limit *int) ([]*models.User, error) {
	return r.users.Users(ctx, user, filter, sort, offset, limit)
}

func (r *queryResolver) UsersByObject(ctx context.Context, object *models.Organization, user *models.User, filter *models.UserFilter, sort []models.UserSort, offset *int, limit *int) ([]*models.User, error) {
	return r.users.UsersByObject(ctx, object, user, filter, sort, offset, limit)
}

func (r *queryResolver) UserLocation(ctx context.Context, lat *float64, lon *float64) (*bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *subscriptionResolver) UserSub(ctx context.Context) (<-chan *models.User, error) {
	return r.users.UserSub(ctx)
}
