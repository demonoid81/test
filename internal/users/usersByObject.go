package users

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/sphera-erp/sphera/internal/models"
	"github.com/sphera-erp/sphera/internal/utils"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func (r *Resolver) UsersByObject(ctx context.Context, object *models.Organization, user *models.User, filter *models.UserFilter, sort []models.UserSort, offset *int, limit *int) ([]*models.User, error) {
	var err error

	table := "users"
	logger := r.env.Logger.Error().Str("package", "users").Str("func", "Users")
	if object == nil || object.UUID == nil {
		logger.Err(err).Msg("Error Object Empty")
		return nil, gqlerror.Errorf("Error Object Empty")
	}
	sql := pglxqb.Select(fmt.Sprintf("%s.*", table)).From(table)
	if filter != nil {
		sql = utils.ReflectFilter(table, sql, filter)
	} else if user != nil {
		var result map[string]interface{}
		result, sql, err = models.SqlGenSelectKeys(user, sql, table, 1)
		if err != nil {
			logger.Err(err).Msg("Error generate select relations")
			return nil, gqlerror.Errorf("Error generate select relations")
		}
		if len(result) > 0 {
			sql = sql.Where(pglxqb.Eq(result))
		}
		fmt.Println(sql.ToSql())
	}
	sql = sql.Where(pglxqb.Expr("?::uuid = ANY ("+fmt.Sprintf("%s.uuid_objects", table)+")", object.UUID))
	if sort != nil {
		for _, sortItem := range sort {
			sql = sql.OrderBy(fmt.Sprintf("%s.%s %s", table, sortItem.Field, sortItem.Order))
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
		logger.Err(err).Msg("Error select users")
		return nil, gqlerror.Errorf("Error select users")
	}
	return user.ParseRows(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, r.env.Cockroach)
}
