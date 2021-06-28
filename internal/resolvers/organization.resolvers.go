package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/internal/models"
)

func (r *mutationResolver) OrganizationPositionMutation(ctx context.Context, organizationPosition *models.OrganizationPosition) (*models.OrganizationPosition, error) {
	return r.organizations.OrganizationPositionMutation(ctx, organizationPosition)
}

func (r *mutationResolver) OrganizationContactMutation(ctx context.Context, organizationContact *models.OrganizationContact) (*models.OrganizationContact, error) {
	return r.organizations.OrganizationContactMutation(ctx, organizationContact)
}

func (r *mutationResolver) OrganizationMutation(ctx context.Context, organization *models.Organization) (*models.Organization, error) {
	return r.organizations.OrganizationMutation(ctx, organization)
}

func (r *mutationResolver) ExcludePerson(ctx context.Context, organization uuid.UUID, person uuid.UUID) (bool, error) {
	return r.organizations.ExcludePerson(ctx, organization, person)
}

func (r *mutationResolver) ExcludePersonInObject(ctx context.Context, organization uuid.UUID, person uuid.UUID) (bool, error) {
	return r.organizations.ExcludePersonInObject(ctx, organization, person)
}

func (r *mutationResolver) DropOrganization(ctx context.Context, organization *models.Organization) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) RemoveParent(ctx context.Context, organization *models.Organization) (bool, error) {
	return r.organizations.RemoveParent(ctx, organization)
}

func (r *queryResolver) OrganizationPosition(ctx context.Context, organizationPosition *models.OrganizationPosition) (*models.OrganizationPosition, error) {
	return r.organizations.OrganizationPosition(ctx, organizationPosition)
}

func (r *queryResolver) OrganizationPositions(ctx context.Context, organizationPosition *models.OrganizationPosition, offset *int, limit *int) ([]*models.OrganizationPosition, error) {
	return r.organizations.OrganizationPositions(ctx, organizationPosition)
}

func (r *queryResolver) OrganizationContact(ctx context.Context, organizationContact *models.OrganizationContact) (*models.OrganizationContact, error) {
	return r.organizations.OrganizationContact(ctx, organizationContact)
}

func (r *queryResolver) OrganizationContacts(ctx context.Context, organizationContact *models.OrganizationContact, offset *int, limit *int) ([]*models.OrganizationContact, error) {
	return r.organizations.OrganizationContacts(ctx, organizationContact)
}

func (r *queryResolver) Organization(ctx context.Context, organization *models.Organization) (*models.Organization, error) {
	return r.organizations.Organization(ctx, organization)
}

func (r *queryResolver) Organizations(ctx context.Context, organization *models.Organization, offset *int, limit *int) ([]*models.Organization, error) {
	return r.organizations.Organizations(ctx, organization)
}

func (r *queryResolver) GetOrganizationRating(ctx context.Context, organization *models.Organization) (*float64, error) {
	return r.organizations.GetOrganizationRating(ctx, organization)
}

func (r *subscriptionResolver) OrganizationSub(ctx context.Context) (<-chan *models.Organization, error) {
	return r.organizations.OrganizationSub(ctx)
}
