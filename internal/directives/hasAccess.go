package directives

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/sphera-erp/sphera/app"
)

type HasAccess func(ctx context.Context, obj interface{}, next graphql.Resolver, attributes ResourceAttributes) (res interface{}, err error)

func NewHasAccess(app *app.App) HasAccess {
	return func(ctx context.Context, obj interface{}, next graphql.Resolver, attributes ResourceAttributes) (res interface{}, err error) {
		//uuid, err := users.UserUUIDForContext(ctx)
		////if err != nil {
		////	app.Logger.Error().Msgf("Error while receiving user information: %v", err)
		////	return nil, gqlerror.Errorf("access denied due to problems on the server side")
		////}
		//fmt.Println(uuid)
		return next(ctx)
	}
}
