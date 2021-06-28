package jobs

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/internal/models"
)

var SUB_ALL = "*"

type jobSub struct {
	Chanel     chan *models.Job
	SubContext context.Context
}

type SubscriptionMutateJobResults struct {
	mx               sync.RWMutex
	MutateJobResults map[uuid.UUID]map[uuid.UUID]jobSub
}

var SubscriptionsMutateJobResults SubscriptionMutateJobResults

func init() {
	SubscriptionsMutateJobResults = SubscriptionMutateJobResults{
		MutateJobResults: make(map[uuid.UUID]map[uuid.UUID]jobSub),
	}
}

func (s *SubscriptionMutateJobResults) Load(key uuid.UUID) (map[uuid.UUID]jobSub, bool) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	val, ok := s.MutateJobResults[key]
	return val, ok
}

func (s *SubscriptionMutateJobResults) Insert(UUIDUser uuid.UUID, channel map[uuid.UUID]jobSub) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.MutateJobResults[UUIDUser] = channel
}

func (s *SubscriptionMutateJobResults) Subscription(ctx context.Context, UUIDUser, subId uuid.UUID, result jobSub) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.MutateJobResults[UUIDUser][subId] = result
}

func (s *SubscriptionMutateJobResults) Delete(UUIDUser, subId uuid.UUID) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	delete(s.MutateJobResults[UUIDUser], subId)
}

func (r *Resolver) JobSub(ctx context.Context) (<-chan *models.Job, error) {
	// UUIDUser, err := middleware.ExtractUserInTokenMetadata(ctx, r.env)
	// if err != nil {
	// 	r.env.Logger.Error().Str("module", "persons").Str("func", "OrganizationSub").Err(err).Msg("Error get user in token metadata")
	// 	return nil, gqlerror.Errorf("Error get user in token metadata")
	// }
	subId := uuid.New()

	UUIDUser := uuid.Nil
	c := make(chan *models.Job, 1)
	go func() {
		<-ctx.Done()
		SubscriptionsMutateJobResults.Delete(UUIDUser, subId)
	}()
	if _, ok := SubscriptionsMutateJobResults.Load(UUIDUser); !ok {
		SubscriptionsMutateJobResults.Insert(UUIDUser, make(map[uuid.UUID]jobSub))
	}
	SubscriptionsMutateJobResults.Subscription(ctx, UUIDUser, subId, jobSub{
		Chanel:     c,
		SubContext: ctx,
	})
	return c, nil
}
