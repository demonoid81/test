package users

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/internal/models"
)

var SUB_ALL = "*"

type userSub struct {
	Chanel     chan *models.User
	SubContext context.Context
}

type SubscriptionMutateUserResults struct {
	mx                sync.RWMutex
	MutateUserResults map[uuid.UUID]map[uuid.UUID]userSub
}

var SubscriptionsMutateUserResults SubscriptionMutateUserResults

func init() {
	SubscriptionsMutateUserResults = SubscriptionMutateUserResults{
		MutateUserResults: make(map[uuid.UUID]map[uuid.UUID]userSub),
	}
}

func (s *SubscriptionMutateUserResults) Load(key uuid.UUID) (map[uuid.UUID]userSub, bool) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	val, ok := s.MutateUserResults[key]
	return val, ok
}

func (s *SubscriptionMutateUserResults) Insert(UUIDUser uuid.UUID, channel map[uuid.UUID]userSub) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.MutateUserResults[UUIDUser] = channel
}

func (s *SubscriptionMutateUserResults) Subscription(ctx context.Context, UUIDUser, subId uuid.UUID, result userSub) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.MutateUserResults[UUIDUser][subId] = result
}

func (s *SubscriptionMutateUserResults) Delete(UUIDUser, subId uuid.UUID) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	delete(s.MutateUserResults[UUIDUser], subId)
}

func (r *Resolver) UserSub(ctx context.Context) (<-chan *models.User, error) {
	// UUIDUser, err := middleware.ExtractUserInTokenMetadata(ctx, r.env)
	// if err != nil {
	// 	r.env.Logger.Error().Str("module", "persons").Str("func", "OrganizationSub").Err(err).Msg("Error get user in token metadata")
	// 	return nil, gqlerror.Errorf("Error get user in token metadata")
	// }
	subId := uuid.New()

	UUIDUser := uuid.Nil
	c := make(chan *models.User, 1)
	go func() {
		<-ctx.Done()
		SubscriptionsMutateUserResults.Delete(UUIDUser, subId)
	}()
	if _, ok := SubscriptionsMutateUserResults.Load(UUIDUser); !ok {
		SubscriptionsMutateUserResults.Insert(UUIDUser, make(map[uuid.UUID]userSub))
	}
	SubscriptionsMutateUserResults.Subscription(ctx, UUIDUser, subId, userSub{
		Chanel:     c,
		SubContext: ctx,
	})
	return c, nil
}
