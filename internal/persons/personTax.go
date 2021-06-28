package persons

import (
	"context"

	"github.com/sphera-erp/sphera/internal/models"
)

func (r *Resolver) PersonTax(ctx context.Context) (*models.Taxes, error) {
	tax := 0.0

	return &models.Taxes{
		Proceeds:    &tax,
		Preliminary: &tax,
		Tax:         &tax,
		Penalty:     &tax,
	}, nil
}
