package flow

import (
	"context"
	"fmt"
	"reflect"

	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/internal/middleware"
	"github.com/sphera-erp/sphera/internal/models"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func (r *Resolver) AddMsg(ctx context.Context, job *models.Job, description string, content []*models.Content) (bool, error) {

	if job == nil || job.UUID == nil {
		return false, gqlerror.Errorf("Error Job  is nil")
	}
	userUUID, err := middleware.ExtractUserInTokenMetadata(ctx, r.env)
	if err != nil {
		return false, gqlerror.Errorf("Error get user uuid from context")
	}

	tx, err := r.env.Cockroach.BeginX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "AddMsg").Err(err).Msg("Error run transaction")
		return false, gqlerror.Errorf("Error run transaction")
	}
	defer tx.Rollback(ctx)
	// найдем персону по uuid

	pRows, err := pglxqb.SelectAll().
		From("persons").
		Where(pglxqb.Eq{"uuid_user": userUUID}).
		RunWith(tx).QueryX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "AddMsg").Err(err).Msg("Error Select person from user ")
		return false, gqlerror.Errorf("Error run transaction")
	}

	var person models.Person
	for pRows.Next() {
		if err := pRows.StructScan(&person); err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "AddMsg").Err(err).Msg("Error select status from jobs")
			return false, gqlerror.Errorf("Error select status from jobs")
		}
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

	//

	var statusesUUID []uuid.UUID
	err = pglxqb.Select("uuid_statuses").From("jobs").Where(pglxqb.Eq{"uuid": job.UUID}).RunWith(tx).QueryRow(ctx).Scan(&statusesUUID)
	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "AddMsg").Err(err).Msg("Error select status from jobs")
		tx.Rollback(ctx)
		return false, gqlerror.Errorf("Error select status from jobs")
	}

	statusUUID := uuid.New()
	sql := pglxqb.Insert("statuses").
		Columns("uuid", "uuid_job", "status", "description", "uuid_person", "uuid_executor").
		Values(statusUUID, job.UUID, job.Status, description, person.UUID, job.UUIDExecutor)
	if len(content) > 0 {
		var contentUUID []uuid.UUID
		for _, c := range content {
			contentUUID = append(contentUUID, *c.UUID)
		}
		sql = sql.Columns("content").Values(contentUUID)
	}
	if _, err = sql.
		RunWith(tx).Exec(ctx); err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "AddMsg").Err(err).Msg("Error set job status")
		tx.Rollback(ctx)
		return false, gqlerror.Errorf("Error set job status")
	}
	if _, err = pglxqb.Update("jobs").
		Set("status", job.Status).
		Set("uuid_statuses", append(statusesUUID, statusUUID)).
		Where(pglxqb.Eq{"uuid": job.UUID}).
		RunWith(tx).Exec(ctx); err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "RunJob").Err(err).Msg("Error update job status")
		tx.Rollback(ctx)
		return false, gqlerror.Errorf("Error run transaction")
	}
	// добавим статусы чтения
	rows, err := pglxqb.SelectAll().
		From("msg_stats").
		Where(pglxqb.Eq{"uuid_job": job.UUID}).
		RunWith(tx).QueryX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "AddMsg").Err(err).Msg("Error select status from jobs")
		return false, gqlerror.Errorf("Error select status from jobs")
	}

	findPersonInStat := false
	var msgStats []models.MsgStat
	for rows.Next() {
		var msgStat models.MsgStat
		if err := rows.StructScan(&msgStat); err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "AddMsg").Err(err).Msg("Error select status from jobs")
			return false, gqlerror.Errorf("Error select status from jobs")
		}
		msgStats = append(msgStats, msgStat)
	}

	for _, msgStat := range msgStats {
		fmt.Println("send response", msgStat)
		if reflect.DeepEqual(msgStat.UUIDPerson, person.UUID) {
			findPersonInStat = true
			if _, err := pglxqb.Update("msg_stats").
				Set("reading", true).
				Where(pglxqb.Eq{"uuid_person": person.UUID}).
				Where(pglxqb.Eq{"uuid_job": job.UUID}).
				RunWith(tx).Exec(ctx); err != nil {
				r.env.Logger.Error().Str("module", "flow").Str("func", "AddMsg").Err(err).Msg("Error select status from jobs")
				return false, gqlerror.Errorf("Error select status from jobs")
			}
			for _, c := range subscriptionsMsgStatUpdate.msgStatUpdate[userUUID] {
				fmt.Println("send response")
				c <- &models.MsgStat{
					Job:     job,
					Person:  &person,
					Reading: true,
				}
			}
		} else {
			if _, err := pglxqb.Update("msg_stats").Set("reading", false).
				Where(pglxqb.Eq{"uuid_person": msgStat.UUIDPerson}).
				Where(pglxqb.Eq{"uuid_job": job.UUID}).
				RunWith(tx).Exec(ctx); err != nil {
				r.env.Logger.Error().Str("module", "flow").Str("func", "AddMsg").Err(err).Msg("Error select status from jobs")
				return false, gqlerror.Errorf("Error select status from jobs")
			}
			var sendUserUUID uuid.UUID
			if err = pglxqb.Select("uuid_user").
				From("persons").
				Where(pglxqb.Eq{"uuid": msgStat.UUIDPerson}).
				RunWith(tx).QueryRow(ctx).Scan(&sendUserUUID); err != nil {
				r.env.Logger.Error().Str("module", "flow").Str("func", "AddMsg").Err(err).Msg("Error select status from jobs")
				return false, gqlerror.Errorf("Error select status from jobs")
			}
			for _, c := range subscriptionsMsgStatUpdate.msgStatUpdate[sendUserUUID] {
				fmt.Println("send response")
				c <- &models.MsgStat{
					Job:     job,
					Person:  &person,
					Reading: false,
				}
			}
		}
	}

	if !findPersonInStat {
		if _, err := pglxqb.Insert("msg_stats").
			Columns("uuid_person", "uuid_job", "reading").
			Values(person.UUID, job.UUID, true).
			RunWith(tx).Exec(ctx); err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "AddMsg").Err(err).Msg("Error select status from jobs")
			return false, gqlerror.Errorf("Error select status from jobs")
		}
		for _, c := range subscriptionsMsgStatUpdate.msgStatUpdate[userUUID] {
			fmt.Println("send response")
			c <- &models.MsgStat{
				Job:     job,
				Person:  &person,
				Reading: true,
			}
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error commit transaction")
		tx.Rollback(ctx)
		return false, gqlerror.Errorf("Error commit transaction")
	}
	return true, nil
}
