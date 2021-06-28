package organizations

import (
	"context"

	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func remove(slice []uuid.UUID, s int) []uuid.UUID {
	return append(slice[:s], slice[s+1:]...)
}

func (r *Resolver) ExcludePersonInObject(ctx context.Context, organization uuid.UUID, person uuid.UUID) (bool, error) {

	tx, err := r.env.Cockroach.BeginX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "RunJob").Err(err).Msg("Error run transaction")
		return false, gqlerror.Errorf("Error run transaction")
	}
	defer tx.Rollback(ctx)

	var isGroup bool
	if err = pglxqb.Select("is_group").From("organizations").
		Where(pglxqb.Eq{"uuid": organization}).
		RunWith(r.env.Cockroach).QueryRow(ctx).Scan(&isGroup); err != nil {
		r.env.Logger.Error().Str("module", "organizations").Str("func", "ExcludePersonInObject").Err(err).Msg("Error get person objects")
		return false, gqlerror.Errorf("Error get person objects")
	}

	if isGroup {
		var objects []uuid.UUID
		var user uuid.UUID
		if err = pglxqb.Select("users.uuid_groups, users.uuid").From("users").
			LeftJoin("persons p on p.uuid_user = users.uuid").
			Where(pglxqb.Eq{"p.uuid": person}).
			RunWith(r.env.Cockroach).QueryRow(ctx).Scan(&objects, &user); err != nil {
			r.env.Logger.Error().Str("module", "organizations").Str("func", "ExcludePersonInObject").Err(err).Msg("Error get person objects")
			return false, gqlerror.Errorf("Error get person objects")
		}

		for i, obj := range objects {
			if obj == organization {
				objects = remove(objects, i)
				break
			}
		}

		if _, err = pglxqb.Update("users").
			Set("uuid_groups", objects).
			Where(pglxqb.Eq{"uuid": user}).
			RunWith(tx).Exec(ctx); err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "RunJob").Err(err).Msg("Error update job status")
			return false, gqlerror.Errorf("Error run transaction")
		}
	} else {
		var objects []uuid.UUID
		var user uuid.UUID
		if err = pglxqb.Select("users.uuid_objects, users.uuid").From("users").
			LeftJoin("persons p on p.uuid_user = users.uuid").
			Where(pglxqb.Eq{"p.uuid": person}).
			RunWith(r.env.Cockroach).QueryRow(ctx).Scan(&objects, &user); err != nil {
			r.env.Logger.Error().Str("module", "organizations").Str("func", "ExcludePersonInObject").Err(err).Msg("Error get person objects")
			return false, gqlerror.Errorf("Error get person objects")
		}

		for i, obj := range objects {
			if obj == organization {
				objects = remove(objects, i)
				break
			}
		}

		if _, err = pglxqb.Update("users").
			Set("uuid_objects", objects).
			Where(pglxqb.Eq{"uuid": user}).
			RunWith(tx).Exec(ctx); err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "RunJob").Err(err).Msg("Error update job status")
			return false, gqlerror.Errorf("Error run transaction")
		}
	}

	if err = tx.Commit(ctx); err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "RunJob").Err(err).Msg("Error commit transaction")
		return false, gqlerror.Errorf("Error commit transaction")
	}
	return true, nil
}
