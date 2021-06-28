package flow

import (
	"context"
	"fmt"

	"github.com/sphera-erp/sphera/internal/middleware"
	"github.com/sphera-erp/sphera/internal/models"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func (r *Resolver) ReadMsg(ctx context.Context, job models.Job) (bool, error) {
	if job.UUID == nil {
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

	if _, err := pglxqb.Update("msg_stats").
		Set("reading", true).
		Where(pglxqb.Eq{"uuid_person": person.UUID}).
		Where(pglxqb.Eq{"uuid_job": job.UUID}).
		RunWith(tx).Exec(ctx); err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "AddMsg").Err(err).Msg("Error select status from jobs")
		return false, gqlerror.Errorf("Error select status from jobs")
	}

	err = tx.Commit(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error commit transaction")
		return false, gqlerror.Errorf("Error commit transaction")
	}

	for _, c := range subscriptionsMsgStatUpdate.msgStatUpdate[userUUID] {
		fmt.Println("send response")
		c <- &models.MsgStat{
			Job:     &job,
			Person:  &person,
			Reading: true,
		}
	}

	return true, nil

}
