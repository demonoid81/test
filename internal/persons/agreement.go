package persons

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/internal/middleware"
	"github.com/sphera-erp/sphera/internal/models"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func (r *Resolver) Agreement(ctx context.Context, incomeRegistration bool, taxPayment bool) (bool, error) {
	userUUID, err := middleware.ExtractUserInTokenMetadata(ctx, r.env)
	if err != nil {
		r.env.Logger.Error().Str("module", "persons").Str("func", "jobs").Err(err).Msg("Error get user uuid from context")
		return false, gqlerror.Errorf("Error get user uuid from context")
	}

	tx, err := r.env.Cockroach.BeginX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "persons").Str("func", "Agreement").Err(err).Msg("Error run transaction")
		return false, gqlerror.Errorf("Error run transaction")
	}
	defer tx.Rollback(ctx)

	if _, err := pglxqb.Update("persons").
		Set("income_registration", incomeRegistration).
		Set("tax_payment", taxPayment).
		Where(pglxqb.Eq{"uuid_user": userUUID}).
		RunWith(tx).Exec(ctx); err != nil {
		r.env.Logger.Error().Str("module", "persons").Str("func", "Agreement").Err(err).Msg("Error set agreement to NPD")
		return false, gqlerror.Errorf("Error set agreement to NPD")
	}

	err = tx.Commit(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "persons").Str("func", "Agreement").Err(err).Msg("Error commit transaction")
		return false, gqlerror.Errorf("Error commit transaction")
	}

	rows, err := pglxqb.Select("persons.*").From("persons").Where(pglxqb.Eq{"uuid_user": userUUID}).RunWith(r.env.Cockroach).QueryX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "persons").Str("func", "Agreement").Err(err).Msg("Error get person")
		return false, gqlerror.Errorf("Error commit transaction")
	}

	var person models.Person
	for rows.Next() {
		if err := rows.StructScan(&person); err != nil {
			r.env.Logger.Error().Str("module", "persons").Str("func", "Agreement").Err(err).Msg("Error commit transaction")
			return false, gqlerror.Errorf("Error commit transaction")
		}
	}

	for _, c := range SubscriptionsMutatePersonResults.MutatePersonResults[uuid.Nil] {
		personSub := person
		if err := personSub.ParseRequestedFields(ctx, graphql.CollectFieldsCtx(c.SubContext, nil), r.env, r.env.Cockroach); err != nil {
			r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBookMutation").Err(err).Msg("Error parse row in medicalBook")
		}
		c.Chanel <- &personSub
	}

	return true, nil
}
