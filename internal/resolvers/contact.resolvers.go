package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/sphera-erp/sphera/internal/models"
)

func (r *mutationResolver) ContactMutation(ctx context.Context, contact *models.Contact) (*models.Contact, error) {
	return r.contacts.ContactMutation(ctx, contact)
}

func (r *mutationResolver) ContactTypeMutation(ctx context.Context, contactType *models.ContactType) (*models.ContactType, error) {
	return r.contacts.ContactTypeMutation(ctx, contactType)
}

func (r *queryResolver) Contact(ctx context.Context, contact models.Contact) (*models.Contact, error) {
	return r.contacts.Contact(ctx, contact)
}

func (r *queryResolver) Contacts(ctx context.Context, contact *models.Contact, filter *models.ContactFilter, offset *int, limit *int) ([]*models.Contact, error) {
	return r.contacts.Contacts(ctx, contact)
}

func (r *queryResolver) ContactType(ctx context.Context, contactType *models.ContactType) (*models.ContactType, error) {
	return r.contacts.ContactType(ctx, contactType)
}

func (r *queryResolver) ContactTypes(ctx context.Context, contactType *models.ContactType, filter *models.ContactTypeFilter, offset *int, limit *int) ([]*models.ContactType, error) {
	return r.contacts.ContactTypes(ctx, contactType)
}
