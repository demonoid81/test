package organizations

import (
	"context"

	"github.com/sphera-erp/sphera/internal/models"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func (r *Resolver) GetOrganizationRating(ctx context.Context, organization *models.Organization) (result *float64, err error) {
	if organization == nil || organization.UUID == nil {
		return nil, gqlerror.Errorf("Error organization  is nil")
	}

	err = pglxqb.Select("sum(rating)/(count(uuid)::float)::float as rating").From("jobs").
		Where(pglxqb.Eq{"uuid_object": organization.UUID}).
		RunWith(r.env.Cockroach).QueryRow(ctx).Scan(&result)
	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error Select person from user ")
		return nil, gqlerror.Errorf("Error Select person from user")
	}
	return
}
