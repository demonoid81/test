package accounting

import (
	"context"
	"time"

	"github.com/sphera-erp/sphera/internal/models"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func (r *Resolver) FlowBalance(ctx context.Context, organization *models.Organization, from *time.Time, to *time.Time) ([]*models.Balance, error) {
	var balances []*models.Balance
	rows, err := pglxqb.Select("*").
		From("balances").
		Where(pglxqb.And{pglxqb.Gt{"amount": 0}, pglxqb.Eq{"uuid_organization": organization.UUID}, pglxqb.Between{Field: "created", X: from, Y: to.Add(time.Hour * 24)}}).
		OrderBy("created DESC").
		RunWith(r.env.Cockroach).QueryX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "RunJob").Err(err).Msg("Error get balance organization")
		return nil, gqlerror.Errorf("Error get balance organization")
	}
	for rows.Next() {
		var balance models.Balance
		err := rows.StructScan(&balance)
		if err != nil {
			r.env.Logger.Error().Str("module", "models").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct cityDistrict")
			return nil, gqlerror.Errorf("Error scan response to struct cityDistrict")
		}
		balances = append(balances, &balance)
	}
	return balances, nil
}
