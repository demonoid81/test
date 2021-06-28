package jobs

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/sphera-erp/sphera/internal/models"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func (r *Resolver) JobTemplateMutate(ctx context.Context, jobTemplate *models.JobTemplate) (*models.JobTemplate, error) {
	logger := r.env.Logger.Error().Str("module", "jobTemplate").Str("func", "JobTemplateMutate")
	tx, err := r.env.Cockroach.BeginX(ctx)
	if err != nil {
		logger.Err(err).Msg("Error run transaction")
		return nil, gqlerror.Errorf("Error run transaction")
	}
	defer tx.Rollback(ctx)
	columns := make(map[string]interface{})
	rows, _, err := jobTemplate.Mutation(ctx, tx, r.env, nil, columns)
	if err != nil {
		logger.Err(err).Msg("Error mutation JobTemplate")
		return nil, gqlerror.Errorf("Error run transaction")
	}
	jobTemplate, err = jobTemplate.ParseRow(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, tx)
	if err != nil {
		logger.Err(err).Msg("Error parse row in jobTemplate")
		return nil, gqlerror.Errorf("Error parse row in jobTemplate")
	}
	err = tx.Commit(ctx)
	if err != nil {
		logger.Err(err).Msg("Error commit transaction")
		return nil, gqlerror.Errorf("Error commit transaction")
	}
	return jobTemplate, err
}

func (r *Resolver) JobTemplate(ctx context.Context, jobTemplate *models.JobTemplate) (*models.JobTemplate, error) {
	var err error
	sql := pglxqb.Select("job_templates.*").From("job_templates")
	result, sql, err := models.SqlGenSelectKeys(jobTemplate, sql, "job_templates", 1)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBook").Err(err).Msg("Error generate select relations")
		return nil, gqlerror.Errorf("Error generate select relations")
	}
	rows, err := sql.Where(pglxqb.Eq(result)).RunWith(r.env.Cockroach).QueryX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBook").Err(err).Msg("Error select medicalBook")
		return nil, gqlerror.Errorf("Error select medicalBooks")
	}
	return jobTemplate.ParseRow(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, r.env.Cockroach)
}

func (r *Resolver) JobTemplates(ctx context.Context, jobTemplate *models.JobTemplate, offset *int, limit *int) ([]*models.JobTemplate, error) {
	var err error
	sql := pglxqb.Select("job_templates.*").From("job_templates")
	result, sql, err := models.SqlGenSelectKeys(jobTemplate, sql, "job_templates", 1)
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
	return jobTemplate.ParseRows(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, r.env.Cockroach)
}
