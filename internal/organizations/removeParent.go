package organizations

import (
	"context"

	"github.com/sphera-erp/sphera/internal/models"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func (r *Resolver) RemoveParent(ctx context.Context, organization *models.Organization) (bool, error) {
	if organization == nil || organization.UUID == nil {
		return false, gqlerror.Errorf("Error organization  is nil")
	}

	tx, err := r.env.Cockroach.BeginX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "RunJob").Err(err).Msg("Error run transaction")
		return false, gqlerror.Errorf("Error run transaction")
	}
	defer tx.Rollback(ctx)

	if _, err = pglxqb.Update("organizations").
		Set("uuid_parent", nil).
		Where(pglxqb.Eq{"uuid": organization.UUID}).
		RunWith(tx).Exec(ctx); err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "RunJob").Err(err).Msg("Error update job status")
		return false, gqlerror.Errorf("Error run transaction")
	}
	if err = tx.Commit(ctx); err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "RunJob").Err(err).Msg("Error commit transaction")
		return false, gqlerror.Errorf("Error commit transaction")
	}
	return true, nil
}
