package nationalities

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/sphera-erp/sphera/app"
	"github.com/sphera-erp/sphera/internal/models"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type Resolver struct {
	env *app.App
	Resolvers
}

type Resolvers interface {
	NationalityMutation(ctx context.Context, nationality *models.Nationality) (*models.Nationality, error)
	Nationality(ctx context.Context, nationality *models.Nationality) (*models.Nationality, error)
	Nationalities(ctx context.Context, nationality *models.Nationality) ([]*models.Nationality, error)
}

func NewNationalitiesResolvers(app *app.App) (*Resolver, error) {
	return &Resolver{
		env: app,
	}, nil
}

func (r *Resolver) NationalityMutation(ctx context.Context, nationality *models.Nationality) (*models.Nationality, error) {
	tx, err := r.env.Cockroach.BeginX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "nationalities").Str("func", "NationalityMutation").Err(err).Msg("Error run transaction")
		return nil, gqlerror.Errorf("Error run transaction")
	}
	defer tx.Rollback(ctx)
	columns := make(map[string]interface{})
	rows, _, err := nationality.Mutation(ctx, tx, r.env, nil, columns)
	if err != nil {
		r.env.Logger.Error().Str("module", "nationalities").Str("func", "NationalityMutation").Err(err).Msg("Error mutation nationality")
		return nil, err
	}
	nationality, err = nationality.ParseRow(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, tx)
	if err != nil {
		r.env.Logger.Error().Str("module", "nationalities").Str("func", "NationalityMutation").Err(err).Msg("Error commit transaction")
		return nil, gqlerror.Errorf("Error commit transaction")
	}
	err = tx.Commit(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "nationalities").Str("func", "NationalityMutation").Err(err).Msg("Error commit transaction")
		return nil, gqlerror.Errorf("Error commit transaction")
	}
	return nationality, err
}

func (r *Resolver) Nationality(ctx context.Context, nationality *models.Nationality) (*models.Nationality, error) {
	var err error
	sql := pglxqb.Select("nationalities.*").From("nationalities")
	result, sql, err := models.SqlGenSelectKeys(nationality, sql, "nationalities", 1)
	if err != nil {
		r.env.Logger.Error().Str("module", "nationalities").Str("func", "Nationality").Err(err).Msg("Error generate select relations")
		return nil, gqlerror.Errorf("Error generate select relations")
	}
	rows, err := sql.Where(pglxqb.Eq(result)).RunWith(r.env.Cockroach).QueryX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "nationalities").Str("func", "Nationality").Err(err).Msg("Error select nationality")
		return nil, gqlerror.Errorf("Error select person")
	}
	return nationality.ParseRow(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, r.env.Cockroach)
}

func (r *Resolver) Nationalities(ctx context.Context, nationality *models.Nationality) ([]*models.Nationality, error) {
	var err error
	sql := pglxqb.Select("nationalities.*").From("nationalities")
	result, sql, err := models.SqlGenSelectKeys(nationality, sql, "nationalities", 1)
	if err != nil {
		r.env.Logger.Error().Str("module", "nationalities").Str("func", "Nationalities").Err(err).Msg("Error generate select relations")
		return nil, gqlerror.Errorf("Error generate select relations")
	}
	rows, err := sql.Where(pglxqb.Eq(result)).RunWith(r.env.Cockroach).QueryX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "nationalities").Str("func", "Nationalities").Err(err).Msg("Error select nationalities")
		return nil, gqlerror.Errorf("Error select users")
	}
	return nationality.ParseRows(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, r.env.Cockroach)
}
