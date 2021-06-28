package passports

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/sphera-erp/sphera/app"
	"github.com/sphera-erp/sphera/internal/models"
	"github.com/sphera-erp/sphera/internal/utils"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type Resolver struct {
	env *app.App
	Resolvers
}

type Resolvers interface {
	PassportMutation(ctx context.Context, passport *models.Passport) (*models.Passport, error)
	Passport(ctx context.Context, passport models.Passport) (*models.Passport, error)
	Passports(ctx context.Context, passport *models.Passport, filter models.PassportFilter, sort []models.PassportSort, offset *int, limit *int) ([]*models.Passport, error)
}

func NewPassportsResolvers(app *app.App) (*Resolver, error) {
	return &Resolver{
		env: app,
	}, nil
}

func (r *Resolver) PassportMutation(ctx context.Context, passport *models.Passport) (*models.Passport, error) {
	tx, err := r.env.Cockroach.BeginX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "passports").Str("func", "PassportMutation").Err(err).Msg("Error run transaction")
		return nil, gqlerror.Errorf("Error run transaction")
	}
	defer tx.Rollback(ctx)
	columns := make(map[string]interface{})
	rows, _, err := passport.Mutation(ctx, tx, r.env, nil, columns)
	if err != nil {
		r.env.Logger.Error().Str("module", "passports").Str("func", "PassportMutation").Err(err).Msg("Error mutation passport")
		return nil, err
	}
	passport, err = passport.ParseRow(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, tx)
	if err != nil {
		r.env.Logger.Error().Str("module", "passports").Str("func", "PassportMutation").Err(err).Msg("Error commit transaction")
		return nil, gqlerror.Errorf("Error commit transaction")
	}
	err = tx.Commit(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "passports").Str("func", "PassportMutation").Err(err).Msg("Error commit transaction")
		return nil, gqlerror.Errorf("Error commit transaction")
	}
	return passport, err
}

func (r *Resolver) Passport(ctx context.Context, passport models.Passport) (*models.Passport, error) {
	var err error
	sql := pglxqb.Select("passports.*").From("passports")
	result, sql, err := models.SqlGenSelectKeys(passport, sql, "passports", 1)
	if err != nil {
		r.env.Logger.Error().Str("module", "passports").Str("func", "Passport").Err(err).Msg("Error generate select relations")
		return nil, gqlerror.Errorf("Error generate select relations")
	}
	rows, err := sql.Where(pglxqb.Eq(result)).RunWith(r.env.Cockroach).QueryX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "passports").Str("func", "Passport").Err(err).Msg("Error select passport")
		return nil, gqlerror.Errorf("Error select person")
	}
	return passport.ParseRow(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, r.env.Cockroach)
}

func (r *Resolver) Passports(ctx context.Context, passport *models.Passport, filter *models.PassportFilter, sort []models.PassportSort, offset *int, limit *int) ([]*models.Passport, error) {
	var err error
	table := "passports"
	logger := r.env.Logger.Error().Str("module", "passports").Str("func", "Passports")
	sql := pglxqb.Select(fmt.Sprintf("%s.*", table)).From(table)
	if filter != nil {
		sql = utils.ReflectFilter(table, sql, filter)
	} else if passport != nil {
		result, sql, err := models.SqlGenSelectKeys(passport, sql, "passports", 1)
		if err != nil {
			logger.Err(err).Msg("Error generate select relations")
			return nil, gqlerror.Errorf("Error generate select relations")
		}
		if len(result) > 0 {
			sql = sql.Where(pglxqb.Eq(result))
		}
	}
	if sort != nil {
		for _, sortItem := range sort {
			sql = sql.OrderBy(fmt.Sprintf("%s %s", sortItem.Field, sortItem.Order))
		}
	}
	if limit != nil {
		sql = sql.Limit(uint64(*limit))
	}
	if offset != nil {
		sql = sql.Offset(uint64(*offset))
	}
	rows, err := sql.RunWith(r.env.Cockroach).QueryX(ctx)
	if err != nil {
		logger.Err(err).Msg("Error select passports")
		return nil, gqlerror.Errorf("Error select passports")
	}
	return passport.ParseRows(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, r.env.Cockroach)
}
