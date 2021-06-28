package flow

import (
	"context"
	"github.com/sphera-erp/sphera/internal/middleware"

	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/internal/models"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func (r *Resolver) RejectPerson(ctx context.Context, job *models.Job, person *models.Person, reason string) (bool, error) {
	if job == nil || job.UUID == nil || person == nil || person.UUID == nil {
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
	err = pglxqb.Select("uuid").From("persons").Where(pglxqb.Eq{"uuid_user": userUUID}).RunWith(tx).QueryRow(ctx).Scan(&personUUID)
	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "RefuseJob").Err(err).Msg("Error Select person from user ")
		return false, gqlerror.Errorf("Error run transaction")
	}
	var candidates []candidate
	rows, err := pglxqb.Select("uuid_person", "candidate_tag").
		From("candidates").
		Where(pglxqb.Eq{"uuid_job": job.UUID}).
		OrderBy("candidate_tag, created").
		RunWith(tx).QueryX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error get candidates from db")
		return false, gqlerror.Errorf("Error Select person from user")
	}
	for rows.Next() {
		var c candidate
		if err := rows.Scan(&c.uuidPerson, &c.candidateTag); err != nil {
			r.env.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct Person")
			return false, gqlerror.Errorf("Error scan response to struct Person")
		}
		// нам нудны только основные и запасные работники
		if c.candidateTag == models.Primary.String() || c.candidateTag == models.Secondary.String() {
			candidates = append(candidates, c)
		}
	}

	count := len(candidates)
	setPrimary := false
	for _, vCandidate := range candidates {
		if vCandidate.uuidPerson == *person.UUID {
			// исключим кадидата
			if _, err = pglxqb.Update("candidates").
				Set("candidate_tag", models.Rejected.Point()).
				Where(pglxqb.And{pglxqb.Eq{"uuid_person": person.UUID}, pglxqb.Eq{"uuid_job": job.UUID}}).
				RunWith(tx).Exec(ctx); err != nil {
				r.env.Logger.Error().Str("module", "flow").Str("func", "RefuseJob").Err(err).Msg("Error reject person from candidates")
				return false, gqlerror.Errorf("Error reject person from candidates")
			}

			if _, err = pglxqb.Insert("statuses").
				Columns("uuid_job", "status", "description", "uuid_person", "uuid_executor").
				Values(job.UUID, models.JobStatusReject, reason, personUUID, personUUID).
				RunWith(tx).Exec(ctx); err != nil {
				r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error set job status")
				return false, gqlerror.Errorf("Error run transaction")
			}

			var token *string
			err = pglxqb.Select("users.notification_token").
				From("users").
				LeftJoin("persons p on p.uuid_user = users.uuid").
				Where(pglxqb.Eq{"p.uuid": vCandidate.uuidPerson}).RunWith(tx).QueryRow(ctx).Scan(&token)
			if err != nil {
				r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error select status from jobs")
				tx.Rollback(ctx)
				return false, gqlerror.Errorf("Error select status from jobs")
			}
			if token != nil {
				r.env.SendPush("192.168.10.244:9999", []string{*token}, "Вы сняты с исполнения по инициативе Заказчика")
			}
		} else if vCandidate.candidateTag == models.Secondary.String() && !setPrimary {
			if _, err = pglxqb.Update("candidates").
				Set("candidate_tag", models.Primary.Point()).
				Where(pglxqb.And{pglxqb.Eq{"uuid_person": vCandidate.uuidPerson}, pglxqb.Eq{"uuid_job": job.UUID}}).
				RunWith(tx).Exec(ctx); err != nil {
				r.env.Logger.Error().Str("module", "flow").Str("func", "RefuseJob").Err(err).Msg("Error set person as primary from candidates")
				return false, gqlerror.Errorf("Error set person as primary from candidates")
			}

			var token *string
			err = pglxqb.Select("users.notification_token").
				From("users").
				LeftJoin("persons p on p.uuid_user = users.uuid").
				Where(pglxqb.Eq{"p.uuid": vCandidate.uuidPerson}).RunWith(tx).QueryRow(ctx).Scan(&token)
			if err != nil {
				r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error select status from jobs")
				tx.Rollback(ctx)
				return false, gqlerror.Errorf("Error select status from jobs")
			}
			if token != nil {
				r.env.SendPush("192.168.10.244:9999", []string{*token}, "Ваш статус из резерва переведен на Основной - пожалуйста подтвердите в течении 15 минут вашу готовность выйти на смену")
			}
			setPrimary = true
		}
	}
	sql := pglxqb.Update("jobs").
		Set("status", models.JobStatusPublish).
		Set("uuid_executor", nil)

	if count == 1 {
		sql = sql.Set("is_hot", true)
	}
	// зафиксируем изменений статуса в работе
	if _, err = sql.Where(pglxqb.Eq{"uuid": job.UUID}).
		RunWith(tx).Exec(ctx); err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "RefuseJob").Err(err).Msg("Error update job status")
		return false, gqlerror.Errorf("Error run transaction")
	}
	// зафиксируем изменений статуса
	if _, err = pglxqb.Insert("statuses").
		Columns("uuid_job", "status", "description", "uuid_person").
		Values(job.UUID, models.JobStatusPublish, reason, personUUID).
		RunWith(tx).Exec(ctx); err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error set job status")
		return false, gqlerror.Errorf("Error run transaction")
	}
	err = tx.Commit(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "RefuseJob").Err(err).Msg("Error commit transaction")
		return false, gqlerror.Errorf("Error commit transaction")
	}
	return true, nil
}
