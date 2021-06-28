package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/sphera-erp/sphera/internal"
	"go.opentelemetry.io/otel"
)

func (r *mutationResolver) Ping(ctx context.Context) (*string, error) {
	tr := otel.Tracer("ping")
	ctx, span := tr.Start(ctx, "ping")
	defer span.End()
	return &pong, nil
}

func (r *queryResolver) Ping(ctx context.Context, id *string) (*string, error) {
	return r.ping.Ping(ctx, id)
}

func (r *subscriptionResolver) PingSub(ctx context.Context, id *string) (<-chan *string, error) {
	return r.ping.PingSub(ctx, id)
}

// Mutation returns internal.MutationResolver implementation.
func (r *Resolver) Mutation() internal.MutationResolver { return &mutationResolver{r} }

// Query returns internal.QueryResolver implementation.
func (r *Resolver) Query() internal.QueryResolver { return &queryResolver{r} }

// Subscription returns internal.SubscriptionResolver implementation.
func (r *Resolver) Subscription() internal.SubscriptionResolver { return &subscriptionResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
var pong string = "pong"
