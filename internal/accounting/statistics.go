package accounting

import (
	"context"

	"github.com/sphera-erp/sphera/internal/models"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func (r *Resolver) Statistics(ctx context.Context, organization *models.Organization) (*models.Stat, error) {
	query := `with T as (
				select
					extract(QUARTER FROM m.created) q,
					extract(MONTH FROM m.created) m,
					sum(m.amount) amount
				from movements as m
				where extract(QUARTER FROM m.created) >= extract(QUARTER FROM CURRENT_TIMESTAMP)-1
				and m.uuid_organization = $1
				group by q,m
				)
				select 'last_q' as name, COALESCE(sum(amount),0) amount from T where q = extract(QUARTER FROM CURRENT_TIMESTAMP)-1
				union
				select 'last_m' as name, COALESCE(sum(amount),0) from T where m = extract(MONTH FROM CURRENT_TIMESTAMP)-1
				union
				select 'this_m' as name, COALESCE(sum(amount),0) from T where m = extract(MONTH FROM CURRENT_TIMESTAMP)`

	values := make([]*float64, 4)
	rows, err := r.env.Cockroach.Query(ctx, query, organization.UUID)

	if err != nil {
		r.env.Logger.Err(err).Msg("Error select persons")
		return nil, gqlerror.Errorf("Error get stat")
	}
	count := 0
	var str string
	for rows.Next() {
		count++
		if err := rows.Scan(&str, &values[count]); err != nil {
			r.env.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct Person")
			return nil, gqlerror.Errorf("Error scan response to struct Person")
		}
	}
	result := models.Stat{
		Quarter:       values[1],
		PreviousMonth: values[2],
		Month:         values[3],
	}
	return &result, nil
}
