package roles

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
	RoleMutation(ctx context.Context, role *models.Role) (*models.Role, error)
	Role(ctx context.Context, role *models.Role) (*models.Role, error)
	Roles(ctx context.Context, role *models.Role, offset *int, limit *int) ([]*models.Role, error)
}

func NewRolesResolvers(app *app.App) (*Resolver, error) {
	return &Resolver{
		env: app,
	}, nil
}

func (r *Resolver) RoleMutation(ctx context.Context, role *models.Role) (*models.Role, error) {
	tx, err := r.env.Cockroach.BeginX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBookMutation").Err(err).Msg("Error run transaction")
		return nil, gqlerror.Errorf("Error run transaction")
	}
	defer tx.Rollback(ctx)

	columns := make(map[string]interface{})
	rows, _, err := role.Mutation(ctx, tx, r.env, nil, columns)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBookMutation").Err(err).Msg("Error mutation medicalBook")
		return nil, err
	}
	role, err = role.ParseRow(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, tx)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBookMutation").Err(err).Msg("Error parse row in medicalBook")
		return nil, gqlerror.Errorf("Error parse row in medicalBook")
	}
	err = tx.Commit(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBookMutation").Err(err).Msg("Error commit transaction")
		return nil, gqlerror.Errorf("Error commit transaction")
	}
	return role, err
}

func (r *Resolver) Role(ctx context.Context, role *models.Role) (*models.Role, error) {
	var err error
	sql := pglxqb.Select("roles.*").From("roles")
	result, sql, err := models.SqlGenSelectKeys(role, sql, "roles", 1)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBook").Err(err).Msg("Error generate select relations")
		return nil, gqlerror.Errorf("Error generate select relations")
	}
	rows, err := sql.Where(pglxqb.Eq(result)).RunWith(r.env.Cockroach).QueryX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBook").Err(err).Msg("Error select medicalBook")
		return nil, gqlerror.Errorf("Error select medicalBooks")
	}
	return role.ParseRow(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, r.env.Cockroach)
}

func (r *Resolver) Roles(ctx context.Context, role *models.Role, offset *int, limit *int) ([]*models.Role, error) {
	var err error
	sql := pglxqb.Select("roles.*").From("roles")
	result, sql, err := models.SqlGenSelectKeys(role, sql, "roles", 1)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBooks").Err(err).Msg("Error generate select relations")
		return nil, gqlerror.Errorf("Error generate select relations")
	}
	rows, err := sql.Where(pglxqb.Eq(result)).RunWith(r.env.Cockroach).QueryX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBooks").Err(err).Msg("Error select medicalBooks")
		return nil, gqlerror.Errorf("Error select medicalBooks")
	}
	return role.ParseRows(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, r.env.Cockroach)
}
