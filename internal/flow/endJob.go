package flow

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"

	"github.com/sphera-erp/sphera/internal/jobs"
	"github.com/sphera-erp/sphera/internal/middleware"
	"github.com/sphera-erp/sphera/internal/models"
	"github.com/sphera-erp/sphera/pkg/nalogSoap/incomeRequest"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func (r *Resolver) EndJob(ctx context.Context, code *string) (*models.PersonRating, error) {
	if code == nil {
		return nil, gqlerror.Errorf("Error code is nil")
	}

	if content, ok := JobStartRequestCodes[*code]; ok {
		var jobReq JobStartRequest
		if err := json.Unmarshal([]byte(content), &jobReq); err != nil {
			return nil, gqlerror.Errorf("Error code is nil")
		}
		if jobReq.UUIDJob == uuid.Nil {
			return nil, gqlerror.Errorf("Error UUID job is empty")
		}
		userUUID, err := middleware.ExtractUserInTokenMetadata(ctx, r.env)
		if err != nil {
			return nil, gqlerror.Errorf("Error get user uuid from context")
		}
		tx, err := r.env.Cockroach.BeginX(ctx)
		if err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "RunJob").Err(err).Msg("Error run transaction")
			return nil, gqlerror.Errorf("Error run transaction")
		}
		defer tx.Rollback(ctx)

		var person models.Person
		pRows, err := pglxqb.SelectAll().
			From("persons").
			Where(pglxqb.Eq{"uuid_user": userUUID}).
			RunWith(tx).QueryX(ctx)
		if err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "RunJob").Err(err).Msg("Error Select person from user ")
			return nil, gqlerror.Errorf("Error run transaction")
		}

		for pRows.Next() {
			if err = pRows.StructScan(&person); err != nil {
				r.env.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
				return nil, gqlerror.Errorf("Error scan response to struct user")
			}
		}

		var job models.Job
		rows, err := pglxqb.SelectAll().
			From("jobs").
			Where(pglxqb.Eq{"uuid": jobReq.UUIDJob}).
			RunWith(tx).QueryX(ctx)

		if err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "RunJob").Err(err).Msg("Error Select person from user ")
			return nil, gqlerror.Errorf("Error run transaction")
		}

		for rows.Next() {
			if err = rows.StructScan(&job); err != nil {
				r.env.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
				return nil, gqlerror.Errorf("Error scan response to struct user")
			}
		}

		var uuidOrganization uuid.UUID
		if err := pglxqb.Select("uuid_parent_organization").
			From("organizations").
			Where(pglxqb.Eq{"uuid": job.UUIDObject}).
			RunWith(tx).QueryRow(ctx).Scan(&uuidOrganization); err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "RunJob").Err(err).Msg("Error Select Org from user ")
			return nil, gqlerror.Errorf("Error run transaction")
		}

		var organization models.Organization
		oRows, err := pglxqb.SelectAll().
			From("organizations").
			Where(pglxqb.Eq{"uuid": uuidOrganization}).
			RunWith(tx).QueryX(ctx)
		if err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "RunJob").Err(err).Msg("Error Select Org from user ")
			return nil, gqlerror.Errorf("Error run transaction")
		}
		defer oRows.Close()

		for oRows.Next() {
			if err = oRows.StructScan(&organization); err != nil {
				r.env.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
				return nil, gqlerror.Errorf("Error scan response to struct user")
			}
		}

		var statusesUUID []uuid.UUID
		err = pglxqb.Select("uuid_statuses").From("jobs").Where(pglxqb.Eq{"uuid": jobReq.UUIDJob}).RunWith(tx).QueryRow(ctx).Scan(&statusesUUID)
		if err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error select status from jobs")
			return nil, gqlerror.Errorf("Error select status from jobs")
		}

		// зафиксируем изменений статуса
		statusUUID := uuid.New()
		if _, err = pglxqb.Insert("statuses").
			Columns("uuid", "uuid_job", "status", "uuid_person").
			Values(statusUUID, jobReq.UUIDJob, models.JobStatusEnd, person.UUID).
			RunWith(tx).Exec(ctx); err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "RunJob").Err(err).Msg("Error set job status")
			return nil, gqlerror.Errorf("Error run transaction")
		}

		if _, err = pglxqb.Update("jobs").
			Set("status", models.JobStatusEnd).
			Set("uuid_statuses", append(statusesUUID, statusUUID)).
			Where(pglxqb.Eq{"uuid": jobReq.UUIDJob}).
			RunWith(tx).Exec(ctx); err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "RunJob").Err(err).Msg("Error update job status")
			return nil, gqlerror.Errorf("Error run transaction")
		}

		// плати самозанятому
		setColumns := map[string]interface{}{
			"uuid_organization": uuidOrganization,
			"destination":       "selfEmployer",
			"uuid_person":       job.UUIDExecutor,
			"uuid_job":          job.UUID,
			"amount":            *job.Cost - *job.Cost*6.0/100,
		}

		var movementUUID uuid.UUID
		if err = pglxqb.
			Insert("movements").
			SetMap(setColumns).
			Suffix("RETURNING uuid").
			RunWith(tx).QueryRow(ctx).Scan(&movementUUID); err != nil {
			r.env.Logger.Error().Str("module", "contact").Str("func", "Mutation").Err(err).Msg("Error insert contact")
			return nil, gqlerror.Errorf("Error insert contact")
		}

		setColumns["destination"] = "taxing"
		setColumns["amount"] = *job.Cost * 6.0 / 100

		if _, err = pglxqb.
			Insert("movements").
			SetMap(setColumns).
			RunWith(tx).
			Exec(ctx); err != nil {
			r.env.Logger.Error().Str("module", "contact").Str("func", "Mutation").Err(err).Msg("Error insert contact")
			return nil, gqlerror.Errorf("Error insert contact")
		}

		commission := *job.Cost * *organization.Fee / 100

		// Выплатим вознаграждение за резерв предидущей смены
		if person.Reward != nil {
			setColumns["destination"] = "reward"
			setColumns["amount"] = *person.Reward
			//*job.Cost*.10 - *job.Cost*.10*6.0/100

			if _, err = pglxqb.
				Insert("movements").
				SetMap(setColumns).
				RunWith(tx).
				Exec(ctx); err != nil {
				r.env.Logger.Error().Str("module", "contact").Str("func", "Mutation").Err(err).Msg("Error insert contact")
				return nil, gqlerror.Errorf("Error insert contact")
			}

			setColumns["destination"] = "rewardTax"
			setColumns["amount"] = *person.Reward * 6.0 / 100

			if _, err = pglxqb.
				Insert("movements").
				SetMap(setColumns).
				RunWith(tx).
				Exec(ctx); err != nil {
				r.env.Logger.Error().Str("module", "contact").Str("func", "Mutation").Err(err).Msg("Error insert contact")
				return nil, gqlerror.Errorf("Error insert contact")
			}

			// уберем вознаграждение за резерв
			if _, err = pglxqb.Update("persons").
				Set("reward", nil).
				Where(pglxqb.Eq{"uuid": job.UUIDExecutor}).
				RunWith(tx).Exec(ctx); err != nil {
				r.env.Logger.Error().Str("module", "flow").Str("func", "RunJob").Err(err).Msg("Error update job status")
				return nil, gqlerror.Errorf("Error run transaction")
			}

			// 	commission = commission - (*job.Cost * .10)
		}

		setColumns["destination"] = "commission"
		setColumns["amount"] = commission

		if _, err = pglxqb.
			Insert("movements").
			SetMap(setColumns).
			RunWith(tx).
			Exec(ctx); err != nil {
			r.env.Logger.Error().Str("module", "contact").Str("func", "Mutation").Err(err).Msg("Error insert contact")
			return nil, gqlerror.Errorf("Error insert contact")
		}

		// создадим резерв оплаты для тех кто стоял в резерве

		var reservedPersonsUUID []uuid.UUID
		cRows, err := pglxqb.Select("candidates.uuid_person").From("candidates").
			LeftJoin("persons p on p.uuid = candidates.uuid_person").
			Where(pglxqb.Eq{"candidates.uuid_job": jobReq.UUIDJob}).
			Where(pglxqb.Eq{"candidates.candidate_tag": models.Secondary.String()}).
			Where(pglxqb.Eq{"p.reward": nil}).
			OrderBy("candidates.created").
			RunWith(tx).Query(ctx)
		if err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error select status from jobs")
			return nil, gqlerror.Errorf("Error select status from jobs")
		}

		defer cRows.Close()

		for cRows.Next() {
			var candidate uuid.UUID
			if err = cRows.Scan(&candidate); err != nil {
				r.env.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
				return nil, gqlerror.Errorf("Error scan response to struct user")
			}
			reservedPersonsUUID = append(reservedPersonsUUID, candidate)
		}

		firstReserveRewardAccrued := false
		for _, person := range reservedPersonsUUID {
			rewardCost := *organization.FirstReserveReward

			if firstReserveRewardAccrued {
				rewardCost = *organization.FirstReserveReward
			}

			if _, err = pglxqb.
				Update("persons").
				Set("reward", *job.Cost*rewardCost-*job.Cost*rewardCost*6.0/100).
				Where(pglxqb.Eq{"uuid": person}).
				RunWith(tx).
				Exec(ctx); err != nil {
				r.env.Logger.Error().Str("module", "contact").Str("func", "Mutation").Err(err).Msg("Error insert contact")
				return nil, gqlerror.Errorf("Error insert contact")
			}
			firstReserveRewardAccrued = true
		}

		// отправим всем сообщение TODO

		err = tx.Commit(ctx)

		if err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "RunJob").Err(err).Msg("Error commit transaction")
			return nil, gqlerror.Errorf("Error commit transaction")
		}

		// отправим всем сообщение

		for _, c := range jobs.SubscriptionsMutateJobResults.MutateJobResults[uuid.Nil] {

			rRows, err := pglxqb.SelectAll().From("jobs").Where(pglxqb.Eq{"uuid": job.UUID}).RunWith(r.env.Cockroach).QueryX(ctx)
			if err != nil {
				r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error get candidates from db")
				return nil, gqlerror.Errorf("Error Select person from user")
			}
			var rJob models.Job
			for rRows.Next() {
				if err := rRows.StructScan(&rJob); err != nil {
					r.env.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct candidates")
					return nil, gqlerror.Errorf("Error scan response to struct Person")
				}
			}
			if err := rJob.ParseRequestedFields(ctx, graphql.CollectFieldsCtx(c.SubContext, nil), r.env, r.env.Cockroach); err != nil {
				r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBookMutation").Err(err).Msg("Error parse row in medicalBook")
				// return nil, gqlerror.Errorf("Error parse row in medicalBook")
			}
			c.Chanel <- &rJob
		}

		var personInn string
		if err := pglxqb.Select("inn").
			From("persons").
			Where(pglxqb.Eq{"uuid": job.UUIDExecutor}).
			RunWith(r.env.Cockroach).QueryRow(ctx).Scan(&personInn); err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "RunJob").Err(err).Msg("Error Select Org from user ")
			return nil, gqlerror.Errorf("Error run transaction")
		}
		fmt.Println(*organization.INN)
		link, _, err := incomeRequest.IncomeRequest(r.env, *job.Cost, *job.Name, personInn, *organization.INN)
		fmt.Println(link)

		if _, err = pglxqb.Update("movements").
			Set("link", link).
			Where(pglxqb.Eq{"uuid": movementUUID}).
			RunWith(r.env.Cockroach).Exec(ctx); err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "RunJob").Err(err).Msg("Error update job status")
			return nil, gqlerror.Errorf("Error run transaction")
		}

		delete(JobStartRequestCodes, *code)

		return &models.PersonRating{
			Job:    &job,
			Person: &person,
		}, nil
	}
	return nil, gqlerror.Errorf("Error decode job start request")
}
