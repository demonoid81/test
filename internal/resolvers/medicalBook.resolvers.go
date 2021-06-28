package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/sphera-erp/sphera/internal/models"
)

func (r *mutationResolver) MedicalBookMutation(ctx context.Context, medicalBook *models.MedicalBook) (*models.MedicalBook, error) {
	return r.medicalBooks.MedicalBookMutation(ctx, medicalBook)
}

func (r *queryResolver) MedicalBook(ctx context.Context, medicalBook models.MedicalBook) (*models.MedicalBook, error) {
	return r.medicalBooks.MedicalBook(ctx, medicalBook)
}

func (r *queryResolver) MedicalBooks(ctx context.Context, medicalBook *models.MedicalBook, offset *int, limit *int) ([]*models.MedicalBook, error) {
	return r.medicalBooks.MedicalBooks(ctx, medicalBook)
}
