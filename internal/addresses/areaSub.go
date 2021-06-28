package addresses

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/internal/models"
)

var SUB_ALL = "*"

type channelMutateArea struct {
	mx                 sync.RWMutex
	OrganizationResult map[uuid.UUID]chan *models.Area
}

type subscriptionMutateAreaResults struct {
	mx                sync.RWMutex
	MutateAreaResults map[uuid.UUID]map[uuid.UUID]chan *models.Area
}

var subscriptionsMutateAreaResults subscriptionMutateAreaResults

func init() {
	subscriptionsMutateAreaResults = subscriptionMutateAreaResults{
		MutateAreaResults: make(map[uuid.UUID]map[uuid.UUID]chan *models.Area),
	}
}

func (s *subscriptionMutateAreaResults) Load(key uuid.UUID) (map[uuid.UUID]chan *models.Area, bool) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	val, ok := s.MutateAreaResults[key]
	return val, ok
}

func (s *subscriptionMutateAreaResults) Insert(UUIDUser uuid.UUID, channel map[uuid.UUID]chan *models.Area) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.MutateAreaResults[UUIDUser] = channel
}

func (s *subscriptionMutateAreaResults) Subscription(UUIDUser, subId uuid.UUID, result chan *models.Area) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.MutateAreaResults[UUIDUser][subId] = result
}

func (s *subscriptionMutateAreaResults) Delete(UUIDUser, subId uuid.UUID) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	delete(s.MutateAreaResults[UUIDUser], subId)
}

func (r *Resolver) AreaSub(ctx context.Context) (<-chan *models.Area, error) {
	// UUIDUser, err := middleware.ExtractUserInTokenMetadata(ctx, r.env)
	// if err != nil {
	// 	r.env.Logger.Error().Str("module", "persons").Str("func", "OrganizationSub").Err(err).Msg("Error get user in token metadata")
	// 	return nil, gqlerror.Errorf("Error get user in token metadata")
	// }
	subId := uuid.New()

	UUIDUser := uuid.Nil
	c := make(chan *models.Area, 1)
	go func() {
		<-ctx.Done()
		subscriptionsMutateAreaResults.Delete(UUIDUser, subId)
	}()
	if _, ok := subscriptionsMutateAreaResults.Load(UUIDUser); !ok {
		subscriptionsMutateAreaResults.Insert(UUIDUser, make(map[uuid.UUID]chan *models.Area))
	}
	subscriptionsMutateAreaResults.Subscription(UUIDUser, subId, c)
	return c, nil
}
