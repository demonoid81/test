package persons

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/sphera-erp/sphera/internal/models"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func (r *Resolver) PersonRatings(ctx context.Context, personRating *models.PersonRating, offset *int, limit *int) ([]*models.PersonRating, error) {
	var err error
	table := "person_ratings"
	logger := r.env.Logger.Error().Str("module", "persons").Str("func", "Persons")
	sql := pglxqb.Select(fmt.Sprintf("%s.*", table)).From(table)

	result, sql, err := models.SqlGenSelectKeys(personRating, sql, table, 1)
	if err != nil {
		logger.Err(err).Msg("Error generate select relations")
		return nil, gqlerror.Errorf("Error generate select relations")
	}
	if len(result) > 0 {
		sql = sql.Where(pglxqb.Eq(result))
	}
	if limit != nil {
		sql = sql.Limit(uint64(*limit))
	}
	if offset != nil {
		sql = sql.Offset(uint64(*offset))
	}
	rows, err := sql.RunWith(r.env.Cockroach).QueryX(ctx)
	if err != nil {
		logger.Err(err).Msg("Error select persons")
		return nil, gqlerror.Errorf("Error select persons")
	}
	return personRating.ParseRows(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, r.env.Cockroach)
}
