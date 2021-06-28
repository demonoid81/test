package directives

import (
	"context"
	"errors"
	"github.com/99designs/gqlgen/graphql"
	"github.com/sphera-erp/sphera/app"
	"github.com/sphera-erp/sphera/internal/middleware"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
)

type BlockParsePerson func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error)

func NewBlockParsePerson(app *app.App) Private {
	return func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
		UUIDUser, err := middleware.ExtractUserInTokenMetadata(ctx, app)
		var dest map[string]interface{}
		err = pglxqb.Select("recognized_fields").
			From("persons").
			Where(pglxqb.Eq{"uuid_user":UUIDUser}).
			RunWith(app.Cockroach).QueryRow(ctx).Scan(&dest)
		if err != nil {
			return nil,  errors.New("Error protect reqest")
		}
		blocker := 0
		if len(dest) > 0 {
			for k, item := range dest {
				if k == "error" {
					continue
				}
				if item.(map[string]interface{})["confidence"].(float64) > 0.8 {
					blocker++
				}
			}
		}
		if blocker >= 10 {
			return nil,  errors.New("Method not Valid")
		}
		return next(ctx)
	}
}

