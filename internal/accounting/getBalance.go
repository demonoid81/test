package accounting

import (
	"context"
	"time"

	"github.com/sphera-erp/sphera/internal/models"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func (r *Resolver) GetBalance(ctx context.Context, organization models.Organization, until *time.Time) (*float64, error) {
	var balance *float64
	sql := pglxqb.Select("SUM (amount) AS balance").
		From("balances").
		Where(pglxqb.Eq{"uuid_organization": organization.UUID})
	if until != nil {
		sql = sql.Where(pglxqb.LtOrEq{"created": until})
	}
	if err := sql.RunWith(r.env.Cockroach).QueryRow(ctx).Scan(&balance); err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "RunJob").Err(err).Msg("Error get balance organization")
		return nil, gqlerror.Errorf("Error get balance organization")
	}
	return balance, nil
}
