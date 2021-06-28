package jobs

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/sphera-erp/sphera/internal/models"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func (r *Resolver) LocalityJobCostMutate(ctx context.Context, localityJobCost *models.LocalityJobCost) (*models.LocalityJobCost, error) {
	tx, err := r.env.Cockroach.BeginX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBookMutation").Err(err).Msg("Error run transaction")
		return nil, gqlerror.Errorf("Error run transaction")
	}
	defer tx.Rollback(ctx)
	columns := make(map[string]interface{})
	rows, _, err := localityJobCost.Mutation(ctx, tx, r.env, nil, columns)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBookMutation").Err(err).Msg("Error mutation medicalBook")
		return nil, err
	}
	localityJobCost, err = localityJobCost.ParseRow(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, tx)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBookMutation").Err(err).Msg("Error parse row in medicalBook")
		return nil, gqlerror.Errorf("Error parse row in medicalBook")
	}
	err = tx.Commit(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBookMutation").Err(err).Msg("Error commit transaction")
		return nil, gqlerror.Errorf("Error commit transaction")
	}
	return localityJobCost, err
}

func (r *Resolver) LocalityJobCost(ctx context.Context, localityJobCost *models.LocalityJobCost) (*models.LocalityJobCost, error) {
	var err error
	sql := pglxqb.Select("locality_job_costs.*").From("locality_job_costs")
	result, sql, err := models.SqlGenSelectKeys(localityJobCost, sql, "locality_job_costs", 1)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBook").Err(err).Msg("Error generate select relations")
		return nil, gqlerror.Errorf("Error generate select relations")
	}
	rows, err := sql.Where(pglxqb.Eq(result)).RunWith(r.env.Cockroach).QueryX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBook").Err(err).Msg("Error select medicalBook")
		return nil, gqlerror.Errorf("Error select medicalBooks")
	}
	return localityJobCost.ParseRow(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, r.env.Cockroach)
}

func (r *Resolver) LocalityJobCosts(ctx context.Context, localityJobCost *models.LocalityJobCost, offset *int, limit *int) ([]*models.LocalityJobCost, error) {
	var err error
	sql := pglxqb.Select("locality_job_costs.*").From("locality_job_costs")
	result, sql, err := models.SqlGenSelectKeys(localityJobCost, sql, "locality_job_costs", 1)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBooks").Err(err).Msg("Error generate select relations")
		return nil, gqlerror.Errorf("Error generate select relations")
	}
	sql = sql.Where(pglxqb.Eq(result))
	if limit != nil {
		sql = sql.Limit(uint64(*limit))
	}
	if offset != nil {
		sql = sql.Offset(uint64(*offset))
	}
	rows, err := sql.RunWith(r.env.Cockroach).QueryX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBooks").Err(err).Msg("Error select medicalBooks")
		return nil, gqlerror.Errorf("Error select medicalBooks")
	}
	return localityJobCost.ParseRows(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, r.env.Cockroach)
}
