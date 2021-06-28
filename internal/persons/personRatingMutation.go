package persons

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/sphera-erp/sphera/internal/models"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func (r *Resolver) PersonRatingMutation(ctx context.Context, personRating *models.PersonRating) (*models.PersonRating, error) {
	tx, err := r.env.Cockroach.BeginX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "persons").Str("func", "PersonaMutation").Err(err).Msg("Error run transaction")
		return nil, gqlerror.Errorf("Error run transaction")
	}
	defer tx.Rollback(ctx)

	columns := make(map[string]interface{})
	rows, _, err := personRating.Mutation(ctx, tx, r.env, nil, columns)
	if err != nil {
		_ = tx.Rollback(ctx)
		r.env.Logger.Error().Str("module", "persons").Str("func", "PersonaMutation").Err(err).Msg("Error mutation person")
		return nil, err
	}
	personRating, err = personRating.ParseRow(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, tx)
	if err != nil {
		r.env.Logger.Error().Str("module", "persons").Str("func", "PersonaMutation").Err(err).Msg("Error commit transaction")
		return nil, gqlerror.Errorf("Error commit transaction")
	}
	err = tx.Commit(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "persons").Str("func", "PersonaMutation").Err(err).Msg("Error commit transaction")
		return nil, gqlerror.Errorf("Error commit transaction")
	}
	return personRating, nil
}
