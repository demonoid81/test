package flow

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/internal/jobs"
	"github.com/sphera-erp/sphera/internal/middleware"
	"github.com/sphera-erp/sphera/internal/models"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func (r *Resolver) OnPlace(ctx context.Context, job *models.Job, lat *float64, lon *float64) (bool, error) {
	if job == nil || job.UUID == nil {
		return false, gqlerror.Errorf("Error Job  is nil")
	}
	userUUID, err := middleware.ExtractUserInTokenMetadata(ctx, r.env)
	if err != nil {
		return false, gqlerror.Errorf("Error get user uuid from context")
	}

	tx, err := r.env.Cockroach.BeginX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "OnPlace").Err(err).Msg("Error run transaction")
		return false, gqlerror.Errorf("Error run transaction")
	}
	defer tx.Rollback(ctx)
	// найдем персону по uuid
	var personUUID uuid.UUID
	if err = pglxqb.Select("uuid").
		From("persons").
		Where(pglxqb.Eq{"uuid_user": userUUID}).
		RunWith(tx).QueryRow(ctx).Scan(&personUUID); err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "OnPlace").Err(err).Msg("Error Select person from user ")
		return false, gqlerror.Errorf("Error run transaction")
	}

	var statusesUUID []uuid.UUID
	var object uuid.UUID
	if err = pglxqb.Select("uuid_statuses", "uuid_object").
		From("jobs").
		Where(pglxqb.Eq{"uuid": job.UUID}).
		RunWith(tx).QueryRow(ctx).Scan(&statusesUUID, &object); err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "OnPlace").Err(err).Msg("Error select status from jobs")
		return false, gqlerror.Errorf("Error select status from jobs")
	}

	var notificationTokens []*string
	rows, err := pglxqb.Select("notification_token").
		From("users").
		Where(pglxqb.Expr("?::uuid = ANY (uuid_objects)", object)).
		RunWith(tx).QueryX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "OnPlace").Err(err).Msg("Error select person from user")
		return false, gqlerror.Errorf("Error run transaction")
	}

	for rows.Next() {
		var notificationToken *string
		err := rows.Scan(&notificationToken)
		if err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "OnPlace").Err(err).Msg("Error scan response")
			return false, gqlerror.Errorf("Error scan response")
		}
		notificationTokens = append(notificationTokens, notificationToken)
	}

	// зафиксируем изменений статуса
	statusUUID := uuid.New()
	if _, err = pglxqb.Insert("statuses").
		Columns("uuid", "uuid_job", "status", "lat", "lon", "uuid_executor").
		Values(statusUUID, job.UUID, models.JobStatusOnObject, lat, lon, personUUID).
		RunWith(tx).Exec(ctx); err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "OnPlace").Err(err).Msg("Error set job status")
		return false, gqlerror.Errorf("Error run transaction")
	}
	// зафиксируем изменений статуса в работе
	if _, err = pglxqb.Update("jobs").
		Set("status", models.JobStatusOnObject).
		Set("uuid_statuses", append(statusesUUID, statusUUID)).
		Where(pglxqb.Eq{"uuid": job.UUID}).
		RunWith(tx).Exec(ctx); err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "OnPlace").Err(err).Msg("Error update job status")
		return false, gqlerror.Errorf("Error run transaction")
	}
	err = tx.Commit(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "OnPlace").Err(err).Msg("Error commit transaction")
		return false, gqlerror.Errorf("Error commit transaction")
	}

	for _, token := range notificationTokens {
		if token != nil {
			text := "Исполнитель на месте"
			r.env.SendPush("192.168.10.244:9999", []string{*token}, text)
		}
	}

	for _, c := range jobs.SubscriptionsMutateJobResults.MutateJobResults[uuid.Nil] {

		rows, err := pglxqb.SelectAll().From("jobs").Where(pglxqb.Eq{"uuid": job.UUID}).RunWith(r.env.Cockroach).QueryX(ctx)
		if err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "OnPlace").Err(err).Msg("Error get candidates from db")
			return false, gqlerror.Errorf("Error Select person from user")
		}
		for rows.Next() {
			if err := rows.StructScan(&job); err != nil {
				r.env.Logger.Error().Str("module", "persons").Str("func", "OnPlace").Err(err).Msg("Error scan response to struct candidates")
				return false, gqlerror.Errorf("Error scan response to struct Person")
			}
		}
		if err := job.ParseRequestedFields(ctx, graphql.CollectFieldsCtx(c.SubContext, nil), r.env, r.env.Cockroach); err != nil {
			r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "OnPlace").Err(err).Msg("Error parse row in medicalBook")
			// return nil, gqlerror.Errorf("Error parse row in medicalBook")
		}
		c.Chanel <- job
	}

	return true, nil
}
