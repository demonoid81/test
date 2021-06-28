package accounting

import (
	"context"
	"time"

	"github.com/sphera-erp/sphera/app"
	"github.com/sphera-erp/sphera/internal/models"
)

type Resolver struct {
	env *app.App
	Resolvers
}

type Resolvers interface {
	MovementMutation(ctx context.Context, movement *models.Movement) (*models.Movement, error)
	AddToBalance(ctx context.Context, organization models.Organization, amount float64) (bool, error)
	Movement(ctx context.Context, movement *models.Movement) (*models.Movement, error)
	Movements(ctx context.Context, movement *models.MovementInput, filter *models.MovementFilter, offset *int, limit *int) ([]*models.Movement, error)
	FlowBalance(ctx context.Context, organization *models.Organization, from *time.Time, to *time.Time) ([]*models.Balance, error)
	GetBalance(ctx context.Context, organization models.Organization, until *time.Time) (*float64, error)
	Statistics(ctx context.Context, organization *models.Organization) (*models.Stat, error)
}

func NewAccountingResolvers(app *app.App) (*Resolver, error) {
	return &Resolver{
		env: app,
	}, nil
}
