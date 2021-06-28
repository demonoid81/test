package jobs

import (
	"context"

	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func (r *Resolver) SetJobRating(ctx context.Context, job uuid.UUID, rating float64, description *string) (bool, error) {
	tx, err := r.env.Cockroach.BeginX(ctx)
	if err != nil {
		r.env.Logger.Err(err).Msg("Error run transaction")
		return false, gqlerror.Errorf("Error run transaction")
	}
	defer tx.Rollback(ctx)

	if _, err = pglxqb.Update("jobs").
		Set("rating", rating).
		Set("rating_description", description).
		Where(pglxqb.Eq{"uuid": job}).
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
