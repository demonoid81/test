package flow

import (
	"context"

	"github.com/sphera-erp/sphera/internal/middleware"
	"github.com/sphera-erp/sphera/internal/models"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func (r *Resolver) GetMsgStats(ctx context.Context) ([]*models.MsgStat, error) {
	UUIDUser, err := middleware.ExtractUserInTokenMetadata(ctx, r.env)
	if err != nil {
		r.env.Logger.Error().Str("module", "persons").Str("func", "ParsePersonSub").Err(err).Msg("Error get user in token metadata")
		return nil, gqlerror.Errorf("Error get user in token metadata")
	}
	rows, err := pglxqb.Select("msg_stats.*").
		From("msg_stats").
		LeftJoin("persons p on p.uuid = msg_stats.uuid_person").
		Where(pglxqb.Eq{"p.uuid_user": UUIDUser}).
		Where(pglxqb.Eq{"msg_stats.reading": false}).
		RunWith(r.env.Cockroach).QueryX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "GetMsgStats").Err(err).Msg("Error select msg statuses")
		return nil, gqlerror.Errorf("Error select msg statuses")
	}
	var msgStats []*models.MsgStat
	for rows.Next() {
		var msgStat models.MsgStat
		if err := rows.StructScan(&msgStat); err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "GetMsgStats").Err(err).Msg("Error scan response to struct Person")
			return nil, gqlerror.Errorf("Error scan response to struct Person")
		}
		msgStats = append(msgStats, &msgStat)
	}

	for _, msgstat := range msgStats {
		jrows, err := pglxqb.SelectAll().
			From("jobs").
			Where(pglxqb.Eq{"uuid": msgstat.UUIDJob}).
			RunWith(r.env.Cockroach).QueryX(ctx)
		if err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "GetMsgStats").Err(err).Msg("Error select msg statuses")
			return nil, gqlerror.Errorf("Error select msg statuses")
		}
		var job models.Job

		for jrows.Next() {
			if err := jrows.StructScan(&job); err != nil {
				r.env.Logger.Error().Str("module", "flow").Str("func", "GetMsgStats").Err(err).Msg("Error scan response to struct job")
				return nil, gqlerror.Errorf("Error scan response to struct job")
			}
		}

		prows, err := pglxqb.SelectAll().
			From("persons").
			Where(pglxqb.Eq{"uuid": msgstat.UUIDPerson}).
			RunWith(r.env.Cockroach).QueryX(ctx)
		if err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "GetMsgStats").Err(err).Msg("Error select msg statuses")
			return nil, gqlerror.Errorf("Error select msg statuses")
		}
		var person models.Person
		for prows.Next() {
			if err := prows.StructScan(&person); err != nil {
				r.env.Logger.Error().Str("module", "flow").Str("func", "GetMsgStats").Err(err).Msg("Error scan response to struct Person")
				return nil, gqlerror.Errorf("Error scan response to struct Person")
			}
		}
		msgstat.Job = &job
		msgstat.Person = &person
	}

	return msgStats, nil
}
