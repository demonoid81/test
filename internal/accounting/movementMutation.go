package accounting

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/internal/models"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func (r *Resolver) MovementMutation(ctx context.Context, movement *models.Movement) (*models.Movement, error) {
	tx, err := r.env.Cockroach.BeginX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBookMutation").Err(err).Msg("Error run transaction")
		return nil, gqlerror.Errorf("Error run transaction")
	}
	defer tx.Rollback(ctx)
	columns := make(map[string]interface{})
	rows, _, err := movement.Mutation(ctx, tx, r.env, nil, columns)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBookMutation").Err(err).Msg("Error mutation medicalBook")
		return nil, err
	}
	movement, err = movement.ParseRow(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, tx)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBookMutation").Err(err).Msg("Error parse row in medicalBook")
		return nil, gqlerror.Errorf("Error parse row in medicalBook")
	}
	var org uuid.UUID
	if err = pglxqb.Select("uuid_parent_organization").
		From("organizations").
		Where(pglxqb.Eq{"uuid": movement.Organization.UUID}).
		RunWith(tx).QueryRow(ctx).Scan(&org); err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "RefuseJob").Err(err).Msg("Error get base organization ")
		return nil, gqlerror.Errorf("Error get base organization")
	}
	if org == uuid.Nil {
		org = *movement.Organization.UUID
	}
	if _, err = pglxqb.Insert("balances").
		Columns("uuid_organization", "amount", "uuid_movement").
		Values(org, -*movement.Amount, movement.UUID).
		RunWith(tx).Exec(ctx); err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "RunJob").Err(err).Msg("Error set job status")
		return nil, gqlerror.Errorf("Error update balance")
	}
	err = tx.Commit(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBookMutation").Err(err).Msg("Error commit transaction")
		return nil, gqlerror.Errorf("Error commit transaction")
	}
	return movement, err
}
