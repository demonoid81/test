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

func (r *Resolver) AgreeToJob(ctx context.Context, job *models.Job, user *models.User) (*models.InfoAboutJob, error) {
	if job == nil || job.UUID == nil {
		return nil, gqlerror.Errorf("Error Job  is nil")
	}
	userUUID, err := middleware.ExtractUserInTokenMetadata(ctx, r.env)
	if err != nil {
		return nil, gqlerror.Errorf("Error get user uuid from context")
	}
	if user != nil && user.UUID != nil {
		userUUID = *user.UUID
	}

	tx, err := r.env.Cockroach.BeginX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error run transaction")
		return nil, gqlerror.Errorf("Error run transaction")
	}
	defer tx.Rollback(ctx)

	// проверим что не заполнены кандидаты
	var candidates []candidate
	cRows, err := pglxqb.Select("uuid", "uuid_person", "candidate_tag").
		From("candidates").
		Where(pglxqb.Eq{"uuid_job": job.UUID}).
		RunWith(tx).Query(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error get candidates from db")
		return nil, gqlerror.Errorf("Error Select person from user")
	}
	for cRows.Next() {
		var candidate candidate
		if err := cRows.Scan(&candidate.uuid, &candidate.uuidPerson, &candidate.candidateTag); err != nil {
			r.env.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct candidates")
			return nil, gqlerror.Errorf("Error scan response to struct Person")
		}
		candidates = append(candidates, candidate)
	}

	// посчитаем закрытые ставки
	count := 0
	for _, candidate := range candidates {
		if candidate.candidateTag == models.Primary.String() || candidate.candidateTag == models.Secondary.String() {
			count++
		}
		// если больше двух то все ставки заняты
		if count > 2 {
			r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error job is full")
			return nil, gqlerror.Errorf("Error job is full")
		}
	}

	// достанем персону из пользователя
	var personUUID uuid.UUID
	err = pglxqb.Select("uuid").From("persons").Where(pglxqb.Eq{"uuid_user": userUUID}).RunWith(tx).QueryRow(ctx).Scan(&personUUID)
	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error Select person from user ")
		tx.Rollback(ctx)
		return nil, gqlerror.Errorf("Error Select person from user")
	}

	// защита от повторного добавления
	for _, candidate := range candidates {
		if candidate.uuidPerson == personUUID {
			r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error candidate has already been added")
			return nil, gqlerror.Errorf("Error candidate has already been added")
		}
	}

	// добавим пользователя
	uuidCandidate := uuid.New()
	result := new(models.InfoAboutJob)
	candidateTag := models.Primary.Point()
	// если никого нет то будет основным
	if count == 0 {
		result.WorkerOrder = models.WorkerOrderPrimary.Point()
	} else {
		result.WorkerOrder = models.WorkerOrderSecondary.Point()
		candidateTag = models.Secondary.Point()
	}
	// добавим кандидата
	_, err = pglxqb.Insert("candidates").
		Columns("uuid", "uuid_person", "uuid_job", "candidate_tag").
		Values(uuidCandidate, personUUID, job.UUID, candidateTag).
		RunWith(tx).Exec(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error Select person from user ")
		tx.Rollback(ctx)
		return nil, gqlerror.Errorf("Error person from user")
	}
	// создадим массив кандидатов
	var uuidCandidates []uuid.UUID
	for _, candidate := range candidates {
		uuidCandidates = append(uuidCandidates, candidate.uuid)
	}
	// добавим нового кандидата
	uuidCandidates = append(uuidCandidates, uuidCandidate)

	var statusesUUID []uuid.UUID
	if err = pglxqb.Select("uuid_statuses").
		From("jobs").
		Where(pglxqb.Eq{"uuid": job.UUID}).OrderBy("created ASC").
		RunWith(tx).QueryRow(ctx).Scan(&statusesUUID); err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error select status from jobs")
		tx.Rollback(ctx)
		return nil, gqlerror.Errorf("Error select status from jobs")
	}

	var LastUUIDExecutor *uuid.UUID
	if err = pglxqb.Select("uuid_statuses").
		From("jobs").
		Where(pglxqb.Eq{"uuid": job.UUID}).OrderBy("created DESC").Limit(1).
		RunWith(tx).QueryRow(ctx).Scan(&LastUUIDExecutor); err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error select status from jobs")
		tx.Rollback(ctx)
		return nil, gqlerror.Errorf("Error select status from jobs")
	}

	sql := pglxqb.Update("jobs").
		Set("uuid_candidates", uuidCandidates)
	// Уберем работу из валидных
	if count == 2 {
		// зафиксируем изменений статуса
		statusUUID := uuid.New()
		if _, err = pglxqb.Insert("statuses").
			Columns("uuid", "uuid_job", "status", "uuid_executor").
			Values(statusUUID, job.UUID, models.JobStatusFull, personUUID).
			RunWith(tx).Exec(ctx); err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error set job status")
			tx.Rollback(ctx)
			return nil, gqlerror.Errorf("Error set job status")
		}
		sql = sql.Set("uuid_statuses", append(statusesUUID, statusUUID)).
			Set("status", models.JobStatusFull)
	}
	if _, err = sql.Where(pglxqb.Eq{"uuid": job.UUID}).
		RunWith(tx).Exec(ctx); err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error update job status")
		tx.Rollback(ctx)
		return nil, gqlerror.Errorf("Error run transaction")
	}
	err = tx.Commit(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error commit transaction")
		tx.Rollback(ctx)
		return nil, gqlerror.Errorf("Error commit transaction")
	}
	var token *string
	err = pglxqb.Select("notification_token").
		From("users").
		Where(pglxqb.Eq{"uuid": userUUID}).RunWith(r.env.Cockroach).QueryRow(ctx).Scan(&token)
	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error select status from jobs")
		tx.Rollback(ctx)
		return nil, gqlerror.Errorf("Error select status from jobs")
	}
	if token != nil {
		text := "Вы откликнулись в качестве резервного исполнителя"
		if candidateTag.String() == "primary" {
			text = "Вы откликнулись в качестве основного исполнителя"
		}
		r.env.SendPush("192.168.10.244:9999", []string{*token}, text)
	}

	for _, c := range jobs.SubscriptionsMutateJobResults.MutateJobResults[uuid.Nil] {

		rows, err := pglxqb.SelectAll().From("jobs").Where(pglxqb.Eq{"uuid": job.UUID}).RunWith(r.env.Cockroach).QueryX(ctx)
		if err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error get jobs from db")
			return nil, gqlerror.Errorf("Error get jobs from db")
		}
		for rows.Next() {
			if err := rows.StructScan(&job); err != nil {
				r.env.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct candidates")
				return nil, gqlerror.Errorf("Error scan response to struct Person")
			}
		}
		if err := job.ParseRequestedFields(ctx, graphql.CollectFieldsCtx(c.SubContext, nil), r.env, r.env.Cockroach); err != nil {
			r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBookMutation").Err(err).Msg("Error parse row in medicalBook")
			return nil, gqlerror.Errorf("Error parse row in medicalBook")
		}
		c.Chanel <- job
	}

	return result, nil
}
