package persons

import (
	"context"

	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/internal/middleware"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func (r *Resolver) GetMyRating(ctx context.Context) (result *float64, err error) {
	userUUID, err := middleware.ExtractUserInTokenMetadata(ctx, r.env)
	if err != nil {
		r.env.Logger.Error().Str("module", "jobs").Str("func", "jobs").Err(err).Msg("Error get user uuid from context")
		return nil, gqlerror.Errorf("Error get user uuid from context")
	}
	// достанем персону из пользователя
	var personUUID uuid.UUID
	err = pglxqb.Select("uuid").From("persons").Where(pglxqb.Eq{"uuid_user": userUUID}).RunWith(r.env.Cockroach).QueryRow(ctx).Scan(&personUUID)
	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error Select person from user")
		return nil, gqlerror.Errorf("Error Select person from user")
	}

	if err = pglxqb.Select("sum(rating)/(count(uuid)::float)::float as rating").From("person_ratings").
		Where(pglxqb.Eq{"uuid_person": personUUID}).
		RunWith(r.env.Cockroach).QueryRow(ctx).Scan(&result); err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error get user rating ")
		return nil, gqlerror.Errorf("Error get user rating ")
	}
	return

}
