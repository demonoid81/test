package directives

import (
	"context"
	"errors"

	"github.com/99designs/gqlgen/graphql"
	"github.com/sphera-erp/sphera/app"
	"github.com/sphera-erp/sphera/internal/middleware"
)

type Private func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error)

func NewPrivate(app *app.App) Private {
	return func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
		_, err = middleware.VerifyToken(ctx, app)
		if err != nil {
			return nil, errors.New("AUTHORIZATION_REQUIRED")
		}
		return next(ctx)
	}
}
