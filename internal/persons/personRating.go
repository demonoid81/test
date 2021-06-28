package persons

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/sphera-erp/sphera/internal/models"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func (r *Resolver) PersonRating(ctx context.Context, personRating *models.PersonRating) (*models.PersonRating, error) {
	var err error
	sql := pglxqb.Select("person_ratings.*").From("person_ratings")
	result, sql, err := models.SqlGenSelectKeys(personRating, sql, "persons", 1)
	if err != nil {
		r.env.Logger.Error().Str("module", "persons").Str("func", "PersonRating").Err(err).Msg("Error generate select relations")
		return nil, gqlerror.Errorf("Error generate select relations")
	}
	rows, err := sql.Where(pglxqb.Eq(result)).RunWith(r.env.Cockroach).QueryX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "persons").Str("func", "PersonRating").Err(err).Msg("Error select person rating")
		return nil, gqlerror.Errorf("Error select person rating")
	}
	return personRating.ParseRow(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, r.env.Cockroach)
}
