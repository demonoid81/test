package persons

import (
	"context"
	"fmt"
	"sync"

	"github.com/sphera-erp/sphera/internal/models"

	"github.com/google/uuid"
)

var SUB_ALL = "*"

type personSub struct {
	Chanel     chan *models.Person
	SubContext context.Context
}

type SubscriptionMutatePersonResults struct {
	mx                  sync.RWMutex
	MutatePersonResults map[uuid.UUID]map[uuid.UUID]personSub
}

var SubscriptionsMutatePersonResults SubscriptionMutatePersonResults

func init() {
	SubscriptionsMutatePersonResults = SubscriptionMutatePersonResults{
		MutatePersonResults: make(map[uuid.UUID]map[uuid.UUID]personSub),
	}
}

func (s *SubscriptionMutatePersonResults) Load(key uuid.UUID) (map[uuid.UUID]personSub, bool) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	val, ok := s.MutatePersonResults[key]
	return val, ok
}

func (s *SubscriptionMutatePersonResults) Insert(UUIDUser uuid.UUID, channel map[uuid.UUID]personSub) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.MutatePersonResults[UUIDUser] = channel
}

func (s *SubscriptionMutatePersonResults) Subscription(ctx context.Context, UUIDUser, subId uuid.UUID, result personSub) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.MutatePersonResults[UUIDUser][subId] = result
}

func (s *SubscriptionMutatePersonResults) Delete(UUIDUser, subId uuid.UUID) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	delete(s.MutatePersonResults[UUIDUser], subId)
}

func (r *Resolver) PersonSub(ctx context.Context) (<-chan *models.Person, error) {
	// UUIDUser, err := middleware.ExtractUserInTokenMetadata(ctx, r.env)
	// if err != nil {
	// 	r.env.Logger.Error().Str("module", "persons").Str("func", "OrganizationSub").Err(err).Msg("Error get user in token metadata")
	// 	return nil, gqlerror.Errorf("Error get user in token metadata")
	// }
	subId := uuid.New()

	fmt.Println("********************** Подписка на Персону ************************")
	UUIDUser := uuid.Nil
	c := make(chan *models.Person, 1)
	go func() {
		<-ctx.Done()
		SubscriptionsMutatePersonResults.Delete(UUIDUser, subId)
	}()
	if _, ok := SubscriptionsMutatePersonResults.Load(UUIDUser); !ok {
		SubscriptionsMutatePersonResults.Insert(UUIDUser, make(map[uuid.UUID]personSub))
	}
	SubscriptionsMutatePersonResults.Subscription(ctx, UUIDUser, subId, personSub{
		Chanel:     c,
		SubContext: ctx,
	})
	return c, nil
}
