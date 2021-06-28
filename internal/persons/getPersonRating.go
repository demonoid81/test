package persons

import (
	"context"

	"github.com/sphera-erp/sphera/internal/models"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func (r *Resolver) GetPersonRating(ctx context.Context, person models.Person) (result *float64, err error) {
	if person.UUID == nil {
		return nil, gqlerror.Errorf("Error person uuid is nil")
	}

	if err = pglxqb.Select("sum(rating)/(count(uuid)::float)::float as rating").From("person_ratings").
		Where(pglxqb.Eq{"uuid_person": person.UUID}).
		RunWith(r.env.Cockroach).QueryRow(ctx).Scan(&result); err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error Select person from user ")
		return nil, gqlerror.Errorf("Error Select person from user")
	}
	return

}
