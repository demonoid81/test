package flow

import (
	"context"
	"encoding/json"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/internal/jobs"
	"github.com/sphera-erp/sphera/internal/middleware"
	"github.com/sphera-erp/sphera/internal/models"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func (r *Resolver) RunJob(ctx context.Context, code *string) (bool, error) {
	if code == nil {
		return false, gqlerror.Errorf("Error code is nil")
	}

	if content, ok := JobStartRequestCodes[*code]; ok {
		var reqToJob JobStartRequest
		if err := json.Unmarshal([]byte(content), &reqToJob); err != nil {
			return false, gqlerror.Errorf("Error code is nil")
		}
		if reqToJob.UUIDJob == uuid.Nil {
			return false, gqlerror.Errorf("Error UUID job is empty")
		}
		userUUID, err := middleware.ExtractUserInTokenMetadata(ctx, r.env)
		if err != nil {
			return false, gqlerror.Errorf("Error get user uuid from context")
		}
		tx, err := r.env.Cockroach.BeginX(ctx)
		if err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "RunJob").Err(err).Msg("Error run transaction")
			return false, gqlerror.Errorf("Error run transaction")
		}
		defer tx.Rollback(ctx)
		var personUUID uuid.UUID
		err = pglxqb.Select("uuid").From("persons").Where(pglxqb.Eq{"uuid_user": userUUID}).RunWith(tx).QueryRow(ctx).Scan(&personUUID)
		if err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "RunJob").Err(err).Msg("Error Select person from user ")
			return false, gqlerror.Errorf("Error run transaction")
		}

		_, err = pglxqb.Update("jobs").Set("status", models.JobStatusStart).Where(pglxqb.Eq{"uuid": reqToJob.UUIDJob}).RunWith(tx).Exec(ctx)
		if err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "RunJob").Err(err).Msg("Error update job status")
			return false, gqlerror.Errorf("Error run transaction")
		}

		// var candidates []candidate
		// rowsХ, err := pglxqb.Select("uuid_person", "candidate_tag").
		// 	From("candidates").
		// 	Where(pglxqb.Eq{"uuid_job": job.UUIDJob}).
		// 	RunWith(tx).Query(ctx)
		// if err != nil {
		// 	r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error get candidates from db")
		// 	return false, gqlerror.Errorf("Error Select person from user")
		// }
		// for rowsХ.Next() {
		// 	var candidate candidate
		// 	if err := rowsХ.Scan(&candidate.uuidPerson, &candidate.candidateTag); err != nil {
		// 		r.env.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct candidates")
		// 		return false, gqlerror.Errorf("Error scan response to struct Person")
		// 	}
		// 	candidates = append(candidates, candidate)
		// }

		// for _, candidate := range candidates {
		// 	if candidate.candidateTag == models.Secondary.String() {
		// 		if _, err = pglxqb.Update("persons").
		// 			Set("secondary", true).
		// 			Where(pglxqb.Eq{"uuid": candidate.uuidPerson}).
		// 			RunWith(tx).Exec(ctx); err != nil {
		// 			r.env.Logger.Error().Str("module", "flow").Str("func", "RunJob").Err(err).Msg("Error update job status")
		// 			return false, gqlerror.Errorf("Error run transaction")
		// 		}
		// 	}
		// }

		jRows, err := pglxqb.SelectAll().
			From("jobs").
			Where(pglxqb.Eq{"uuid": reqToJob.UUIDJob}).
			RunWith(tx).QueryX(ctx)
		if err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "AddMsg").Err(err).Msg("Error Select job")
			return false, gqlerror.Errorf("Error Select job")
		}
		var job models.Job
		for jRows.Next() {
			if err := jRows.StructScan(&job); err != nil {
				r.env.Logger.Error().Str("module", "flow").Str("func", "AddMsg").Err(err).Msg("Error parse job")
				return false, gqlerror.Errorf("Error parse job")
			}
		}

		var statusesUUID []uuid.UUID
		if err = pglxqb.Select("uuid_statuses").
			From("jobs").
			Where(pglxqb.Eq{"uuid": reqToJob.UUIDJob}).OrderBy("created ASC").
			RunWith(tx).QueryRow(ctx).Scan(&statusesUUID); err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "RunJob").Err(err).Msg("Error select status from jobs")

			return false, gqlerror.Errorf("Error select status from jobs")
		}

		// зафиксируем изменений статуса
		statusUUID := uuid.New()
		if _, err = pglxqb.Insert("statuses").
			Columns("uuid", "uuid_job", "status", "uuid_person", "uuid_executor").
			Values(statusUUID, reqToJob.UUIDJob, models.JobStatusStart, personUUID, job.UUIDExecutor).
			RunWith(tx).Exec(ctx); err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "RunJob").Err(err).Msg("Error set job status")
			return false, gqlerror.Errorf("Error run transaction")
		}

		if _, err = pglxqb.Update("jobs").
			Set("status", models.JobStatusStart).
			Set("uuid_statuses", append(statusesUUID, statusUUID)).
			Where(pglxqb.Eq{"uuid": reqToJob.UUIDJob}).
			RunWith(tx).Exec(ctx); err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "RunJob").Err(err).Msg("Error update job status")
			tx.Rollback(ctx)
			return false, gqlerror.Errorf("Error run transaction")
		}

		if err = tx.Commit(ctx); err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "RunJob").Err(err).Msg("Error commit transaction")
			return false, gqlerror.Errorf("Error commit transaction")
		}

		rows, err := pglxqb.SelectAll().From("jobs").Where(pglxqb.Eq{"uuid": reqToJob.UUIDJob}).RunWith(r.env.Cockroach).QueryX(ctx)
		if err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error get candidates from db")
			return false, gqlerror.Errorf("Error Select person from user")
		}
		var rJob models.Job
		for rows.Next() {
			if err := rows.StructScan(&rJob); err != nil {
				r.env.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct job")
				return false, gqlerror.Errorf("Error scan response to struct job")
			}
		}

		for _, c := range jobs.SubscriptionsMutateJobResults.MutateJobResults[uuid.Nil] {
			if err := rJob.ParseRequestedFields(ctx, graphql.CollectFieldsCtx(c.SubContext, nil), r.env, r.env.Cockroach); err != nil {
				r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBookMutation").Err(err).Msg("Error parse row in medicalBook")
				return false, gqlerror.Errorf("Error parse row in medicalBook")
			}

			c.Chanel <- &rJob
		}

		delete(JobStartRequestCodes, *code)

		return true, nil
	}
	return false, gqlerror.Errorf("Error decode job start request")
}
