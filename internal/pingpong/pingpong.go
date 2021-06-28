package pingpong

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/app"
)

var pong string = "pong"
var pongs string = "pongs"
var SubAll = "*"
type Resolver struct {
	env *app.App
	Resolvers
}

var Subscriptions = map[string]map[string]chan *string{}


type Resolvers interface {
	Ping(ctx context.Context) (*string, error)
	PingSub(ctx context.Context) (<-chan *string, error)
}

func NewPingResolvers(app *app.App) (*Resolver, error) {
	return &Resolver{
		env: app,
	}, nil
}

func (r *Resolver) Ping(ctx context.Context, id *string) (*string, error) {

	// Publish to filtered subscriptions
	for _, c := range Subscriptions[*id] {
		c <- &pong
	}
	// Publish to non-filtered subscriptions
	for _, c := range Subscriptions[SubAll] {
		c <- &pongs
	}
	return &pong, nil
}

func (r *Resolver) PingSub(ctx context.Context, id *string) (<-chan *string, error) {
	subId := uuid.New().String()

	fmt.Println(subId)
	c := make(chan *string, 1)

	if id == nil {
		id = &SubAll
	}

	go func() {
		<-ctx.Done()
		delete(Subscriptions[*id], subId)
	}()
	if Subscriptions[*id] == nil {
		Subscriptions[*id] = make(map[string]chan *string, 0)
	}
	Subscriptions[*id][subId] = c

	return c, nil
}