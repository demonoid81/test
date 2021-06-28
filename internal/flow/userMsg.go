package flow

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/internal/middleware"
	"github.com/sphera-erp/sphera/internal/models"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func (r *Resolver) UserMsg(ctx context.Context, status *models.Status, offset *int, limit *int) ([]*models.Status, error) {
	var err error
	userUUID, err := middleware.ExtractUserInTokenMetadata(ctx, r.env)
	if err != nil {
		return nil, gqlerror.Errorf("Error get user uuid from context")
	}
	// достанем персону из пользователя
	var personUUID uuid.UUID
	if err = pglxqb.Select("uuid").
		From("persons").
		Where(pglxqb.Eq{"uuid_user": userUUID}).
		RunWith(r.env.Cockroach).
		QueryRow(ctx).Scan(&personUUID); err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error Select person from user ")
		return nil, gqlerror.Errorf("Error Select person from user")
	}

	sql := pglxqb.Select("statuses.*").From("statuses")
	result, sql, err := models.SqlGenSelectKeys(status, sql, "statuses", 1)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBooks").Err(err).Msg("Error generate select relations")
		return nil, gqlerror.Errorf("Error generate select relations")
	}
	sql = sql.Where(pglxqb.Eq(result))
	sql = sql.Where(pglxqb.Eq{"uuid_executor": personUUID})
	if limit != nil {
		sql = sql.Limit(uint64(*limit))
	}
	if offset != nil {
		sql = sql.Offset(uint64(*offset))
	}
	rows, err := sql.RunWith(r.env.Cockroach).QueryX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "UserMsg").Err(err).Msg("Error select statuses")
		return nil, gqlerror.Errorf("Error select statuses")
	}
	return status.ParseRows(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, r.env.Cockroach)
}
