package persons

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/sphera-erp/sphera/internal/models"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func (r *Resolver) PersonCourseMutation(ctx context.Context, course *models.PersonCourse) (*models.PersonCourse, error) {
	tx, err := r.env.Cockroach.BeginX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "users").Str("func", "UserMutation").Err(err).Msg("Error run transaction")
		return nil, gqlerror.Errorf("Error run transaction")
	}
	defer tx.Rollback(ctx)

	columns := make(map[string]interface{})
	rows, _, err := course.Mutation(ctx, tx, r.env, nil, columns)
	if err != nil {
		r.env.Logger.Error().Str("module", "users").Str("func", "UserMutation").Err(err).Msg("Error mutation user")
		return nil, err
	}
	course, err = course.ParseRow(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, tx)
	if err != nil {
		r.env.Logger.Error().Str("module", "users").Str("func", "createEmptyUser").Err(err).Msg("Error commit transaction")
		return nil, gqlerror.Errorf("Error commit transaction")
	}
	err = tx.Commit(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "users").Str("func", "createEmptyUser").Err(err).Msg("Error commit transaction")
		return nil, gqlerror.Errorf("Error commit transaction")
	}
	return course, nil
}
