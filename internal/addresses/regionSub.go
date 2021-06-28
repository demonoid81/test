package addresses

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/internal/models"
)

type channelMutateRegion struct {
	mx                 sync.RWMutex
	OrganizationResult map[uuid.UUID]chan *models.Region
}

type subscriptionMutateRegionResults struct {
	mx                  sync.RWMutex
	MutateRegionResults map[uuid.UUID]map[uuid.UUID]chan *models.Region
}

var subscriptionsMutateRegionResults subscriptionMutateRegionResults

func init() {
	subscriptionsMutateRegionResults = subscriptionMutateRegionResults{
		MutateRegionResults: make(map[uuid.UUID]map[uuid.UUID]chan *models.Region),
	}
}

func (s *subscriptionMutateRegionResults) Load(key uuid.UUID) (map[uuid.UUID]chan *models.Region, bool) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	val, ok := s.MutateRegionResults[key]
	return val, ok
}

func (s *subscriptionMutateRegionResults) Insert(UUIDUser uuid.UUID, channel map[uuid.UUID]chan *models.Region) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.MutateRegionResults[UUIDUser] = channel
}

func (s *subscriptionMutateRegionResults) Subscription(UUIDUser, subId uuid.UUID, result chan *models.Region) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.MutateRegionResults[UUIDUser][subId] = result
}

func (s *subscriptionMutateRegionResults) Delete(UUIDUser, subId uuid.UUID) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	delete(s.MutateRegionResults[UUIDUser], subId)
}

func (r *Resolver) RegionSub(ctx context.Context) (<-chan *models.Region, error) {
	// UUIDUser, err := middleware.ExtractUserInTokenMetadata(ctx, r.env)
	// if err != nil {
	// 	r.env.Logger.Error().Str("module", "persons").Str("func", "OrganizationSub").Err(err).Msg("Error get user in token metadata")
	// 	return nil, gqlerror.Errorf("Error get user in token metadata")
	// }
	subId := uuid.New()

	UUIDUser := uuid.Nil
	c := make(chan *models.Region, 1)
	go func() {
		<-ctx.Done()
		subscriptionsMutateRegionResults.Delete(UUIDUser, subId)
	}()
	if _, ok := subscriptionsMutateRegionResults.Load(UUIDUser); !ok {
		subscriptionsMutateRegionResults.Insert(UUIDUser, make(map[uuid.UUID]chan *models.Region))
	}
	subscriptionsMutateRegionResults.Subscription(UUIDUser, subId, c)
	return c, nil
}
