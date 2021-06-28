package accounting

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/sphera-erp/sphera/internal/models"
	"github.com/sphera-erp/sphera/internal/utils"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func (r *Resolver) Movements(ctx context.Context, movement *models.Movement, filter *models.MovementFilter, offset *int, limit *int) ([]*models.Movement, error) {
	var err error
	var result map[string]interface{}
	table := "movements"
	logger := r.env.Logger.Error().Str("module", "movements").Str("func", "Movements")
	sql := pglxqb.Select(fmt.Sprintf("%s.*", table)).From(table)
	if filter != nil {
		sql = utils.ReflectFilter(table, sql, filter)
	} else if movement != nil {
		result, sql, err = models.SqlGenSelectKeys(movement, sql, table, 1)
		if err != nil {
			logger.Err(err).Msg("Error generate select relations")
			return nil, gqlerror.Errorf("Error generate select relations")
		}
		fmt.Println(result)
		if len(result) > 0 {
			sql = sql.Where(pglxqb.Eq(result))
		}
	}
	if limit != nil {
		sql = sql.Limit(uint64(*limit))
	}
	if offset != nil {
		sql = sql.Offset(uint64(*offset))
	}
	rows, err := sql.RunWith(r.env.Cockroach).QueryX(ctx)
	if err != nil {
		logger.Err(err).Msg("Error select persons")
		return nil, gqlerror.Errorf("Error select persons")
	}
	return movement.ParseRows(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, r.env.Cockroach)
}
