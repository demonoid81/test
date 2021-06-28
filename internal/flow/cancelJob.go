package flow

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/sphera-erp/sphera/internal/jobs"
	"github.com/sphera-erp/sphera/internal/middleware"

	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/internal/models"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func (r *Resolver) CancelJob(ctx context.Context, job *models.Job, reason string) (bool, error) {
	if job == nil || job.UUID == nil {
		return false, gqlerror.Errorf("Error Job  is nil")
	}
	userUUID, err := middleware.ExtractUserInTokenMetadata(ctx, r.env)
	if err != nil {
		return false, gqlerror.Errorf("Error get user uuid from context")
	}
	tx, err := r.env.Cockroach.BeginX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "RefuseJob").Err(err).Msg("Error run transaction")
		return false, gqlerror.Errorf("Error run transaction")
	}
	defer tx.Rollback(ctx)
	// найдем персону по uuid
	var personUUID uuid.UUID
	if err = pglxqb.Select("uuid").
		From("persons").
		Where(pglxqb.Eq{"uuid_user": userUUID}).
		RunWith(tx).QueryRow(ctx).Scan(&personUUID); err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "RefuseJob").Err(err).Msg("Error Select person from user ")
		return false, gqlerror.Errorf("Error run transaction")
	}

	jRows, err := pglxqb.SelectAll().
		From("jobs").
		Where(pglxqb.Eq{"uuid": job.UUID}).
		RunWith(tx).QueryX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "AddMsg").Err(err).Msg("Error Select job")
		return false, gqlerror.Errorf("Error Select job")
	}
	for jRows.Next() {
		if err := jRows.StructScan(&job); err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "AddMsg").Err(err).Msg("Error parse job")
			return false, gqlerror.Errorf("Error parse job")
		}
	}

	var statusesUUID []uuid.UUID
	err = pglxqb.Select("uuid_statuses").From("jobs").Where(pglxqb.Eq{"uuid": job.UUID}).RunWith(tx).QueryRow(ctx).Scan(&statusesUUID)
	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error select status from jobs")
		return false, gqlerror.Errorf("Error select status from jobs")
	}
	// зафиксируем изменений статуса
	statusUUID := uuid.New()
	if _, err = pglxqb.Insert("statuses").
		Columns("uuid", "uuid_job", "status", "description", "uuid_person", "uuid_executor").
		Values(statusUUID, job.UUID, models.JobStatusCancel, reason, personUUID, job.UUIDExecutor).
		RunWith(tx).Exec(ctx); err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error set job status")
		return false, gqlerror.Errorf("Error run transaction")
	}
	// зафиксируем изменений статуса в работе
	if _, err = pglxqb.Update("jobs").
		Set("status", models.JobStatusCancel).
		Set("uuid_statuses", append(statusesUUID, statusUUID)).
		Where(pglxqb.Eq{"uuid": job.UUID}).
		RunWith(tx).Exec(ctx); err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "RunJob").Err(err).Msg("Error update job status")
		return false, gqlerror.Errorf("Error run transaction")
	}
	err = tx.Commit(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "RefuseJob").Err(err).Msg("Error commit transaction")
		return false, gqlerror.Errorf("Error commit transaction")
	}

	for _, c := range jobs.SubscriptionsMutateJobResults.MutateJobResults[uuid.Nil] {

		rows, err := pglxqb.SelectAll().From("jobs").Where(pglxqb.Eq{"uuid": job.UUID}).RunWith(r.env.Cockroach).QueryX(ctx)
		if err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error get candidates from db")
			return false, gqlerror.Errorf("Error Select person from user")
		}
		for rows.Next() {
			if err := rows.StructScan(&job); err != nil {
				r.env.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct candidates")
				return false, gqlerror.Errorf("Error scan response to struct Person")
			}
		}
		if err := job.ParseRequestedFields(ctx, graphql.CollectFieldsCtx(c.SubContext, nil), r.env, r.env.Cockroach); err != nil {
			r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBookMutation").Err(err).Msg("Error parse row in medicalBook")
			// return nil, gqlerror.Errorf("Error parse row in medicalBook")
		}
		c.Chanel <- job
	}

	return true, nil
}
