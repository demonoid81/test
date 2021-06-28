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

type candidate struct {
	uuid         uuid.UUID `db:"uuid"`
	uuidPerson   uuid.UUID `db:"uuid_person"`
	candidateTag string    `db:"candidate_tag"`
}

func (r *Resolver) RefuseJob(ctx context.Context, job *models.Job, reason string) (bool, error) {
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
	if err := pglxqb.Select("uuid").
		From("persons").
		Where(pglxqb.Eq{"uuid_user": userUUID}).
		RunWith(tx).QueryRow(ctx).Scan(&personUUID); err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "RefuseJob").Err(err).Msg("Error Select person from user ")
		tx.Rollback(ctx)
		return false, gqlerror.Errorf("Error run transaction")
	}

	var candidates []candidate
	rows, err := pglxqb.Select("uuid_person", "candidate_tag").From("candidates").Where(pglxqb.Eq{"uuid_job": job.UUID}).OrderBy("candidate_tag, created").RunWith(tx).Query(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error get candidates from db")
		tx.Rollback(ctx)
		return false, gqlerror.Errorf("Error Select person from user")
	}
	for rows.Next() {
		var candidate candidate
		if err := rows.Scan(&candidate.uuidPerson, &candidate.candidateTag); err != nil {
			r.env.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct candidates")
			tx.Rollback(ctx)
			return false, gqlerror.Errorf("Error scan response to struct Person")
		}
		// нам нудны только основные и запасные работники
		if candidate.candidateTag == models.Primary.String() || candidate.candidateTag == models.Secondary.String() {
			candidates = append(candidates, candidate)
		}
	}
	count := len(candidates)
	setPrimary := false
	for _, candidate := range candidates {
		if candidate.uuidPerson == personUUID {
			// Если отказался основной то заменим
			if candidate.candidateTag == models.Primary.String() {
				setPrimary = true
			}
			// исключим кадидата
			if _, err := pglxqb.Update("candidates").
				Set("candidate_tag", models.Refused.Point()).
				Where(pglxqb.And{pglxqb.Eq{"uuid_person": personUUID}, pglxqb.Eq{"uuid_job": job.UUID}}).
				RunWith(tx).Exec(ctx); err != nil {
				r.env.Logger.Error().Str("module", "flow").Str("func", "RefuseJob").Err(err).Msg("Error Select person from user ")
				tx.Rollback(ctx)
				return false, gqlerror.Errorf("Error run transaction")
			}

			if _, err = pglxqb.Insert("statuses").
				Columns("uuid_job", "status", "description", "uuid_person", "uuid_executor").
				Values(job.UUID, models.JobStatusRefuse, reason, personUUID, personUUID).
				RunWith(tx).Exec(ctx); err != nil {
				r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error set job status")
				return false, gqlerror.Errorf("Error run transaction")
			}

			var status string
			var object uuid.UUID
			if err = pglxqb.Select("status", "uuid_object").
				From("jobs").
				Where(pglxqb.Eq{"uuid": job.UUID}).
				RunWith(tx).QueryRow(ctx).Scan(&status, &object); err != nil {
				r.env.Logger.Error().Str("module", "flow").Str("func", "OnPlace").Err(err).Msg("Error select status from jobs")
				return false, gqlerror.Errorf("Error select status from jobs")
			}
			if status == models.JobStatusOnObject.String() || status == models.JobStatusReady.String() {

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

				for _, token := range notificationTokens {
					if token != nil {
						r.env.SendPush("192.168.10.244:9999", []string{*token}, "Испольнитель отказался от смены, выполняем поиск замены")
					}
				}
			}

		} else if candidate.candidateTag == models.Secondary.String() && setPrimary {
			if _, err = pglxqb.Update("candidates").
				Set("candidate_tag", models.Primary.Point()).
				Where(pglxqb.And{pglxqb.Eq{"uuid_person": candidate.uuidPerson}, pglxqb.Eq{"uuid_job": job.UUID}}).
				RunWith(tx).Exec(ctx); err != nil {
				r.env.Logger.Error().Str("module", "flow").Str("func", "RefuseJob").Err(err).Msg("Error Select person from user ")
				tx.Rollback(ctx)
				return false, gqlerror.Errorf("Error run transaction")
			}
			var token *string
			err = pglxqb.Select("users.notification_token").
				From("users").
				LeftJoin("persons p on p.uuid_user = users.uuid").
				Where(pglxqb.Eq{"p.uuid": candidate.uuidPerson}).RunWith(tx).QueryRow(ctx).Scan(&token)
			if err != nil {
				r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error select status from jobs")
				tx.Rollback(ctx)
				return false, gqlerror.Errorf("Error select status from jobs")
			}
			if token != nil {
				r.env.SendPush("192.168.10.244:9999", []string{*token}, "Ваш статус из резерва переведен на Основной - пожалуйста подтвердите в течении 15 минут вашу готовность выйти на смену")
			}
			setPrimary = false
		}
	}

	sql := pglxqb.Update("jobs").
		Set("status", models.JobStatusPublish).
		Set("uuid_executor", nil)
	// у нас нет кандидатов делаем работу горячей
	if count == 1 {
		sql = sql.Set("is_hot", true)
	}
	// зафиксируем изменений статуса в работе
	if _, err = sql.Where(pglxqb.Eq{"uuid": job.UUID}).
		RunWith(tx).Exec(ctx); err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "RefuseJob").Err(err).Msg("Error update job status")
		tx.Rollback(ctx)
		return false, gqlerror.Errorf("Error run transaction")
	}
	// зафиксируем изменений статуса
	if _, err = pglxqb.Insert("statuses").
		Columns("uuid_job", "status", "description", "uuid_person").
		Values(job.UUID, models.JobStatusPublish, reason, personUUID).
		RunWith(tx).Exec(ctx); err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error set job status")
		tx.Rollback(ctx)
		return false, gqlerror.Errorf("Error run transaction")
	}
	err = tx.Commit(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "RefuseJob").Err(err).Msg("Error commit transaction")
		tx.Rollback(ctx)
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
			return false, gqlerror.Errorf("Error parse row in medicalBook")
		}
		c.Chanel <- job
	}

	return true, nil
}
