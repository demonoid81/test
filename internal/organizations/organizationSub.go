package organizations

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/internal/models"
)

var SUB_ALL = "*"

type organizationSub struct {
	Chanel     chan *models.Organization
	SubContext context.Context
}

type channelMutateOrganization struct {
	mx                 sync.RWMutex
	OrganizationResult map[uuid.UUID]chan organizationSub
}

type subscriptionMutateOrganizationResults struct {
	mx                        sync.RWMutex
	MutateOrganizationResults map[uuid.UUID]map[uuid.UUID]organizationSub
}

var SubscriptionsMutateOrganizationResults subscriptionMutateOrganizationResults

func init() {
	SubscriptionsMutateOrganizationResults = subscriptionMutateOrganizationResults{
		MutateOrganizationResults: make(map[uuid.UUID]map[uuid.UUID]organizationSub),
	}
}

func (s *subscriptionMutateOrganizationResults) Load(key uuid.UUID) (map[uuid.UUID]organizationSub, bool) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	val, ok := s.MutateOrganizationResults[key]
	return val, ok
}

func (s *subscriptionMutateOrganizationResults) Insert(UUIDUser uuid.UUID, channel map[uuid.UUID]organizationSub) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.MutateOrganizationResults[UUIDUser] = channel
}

func (s *subscriptionMutateOrganizationResults) Subscription(UUIDUser, subId uuid.UUID, result organizationSub) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.MutateOrganizationResults[UUIDUser][subId] = result
}

func (s *subscriptionMutateOrganizationResults) Delete(UUIDUser, subId uuid.UUID) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	delete(s.MutateOrganizationResults[UUIDUser], subId)
}

func (r *Resolver) OrganizationSub(ctx context.Context) (<-chan *models.Organization, error) {
	// UUIDUser, err := middleware.ExtractUserInTokenMetadata(ctx, r.env)
	// if err != nil {
	// 	r.env.Logger.Error().Str("module", "persons").Str("func", "OrganizationSub").Err(err).Msg("Error get user in token metadata")
	// 	return nil, gqlerror.Errorf("Error get user in token metadata")
	// }
	subId := uuid.New()

	UUIDUser := uuid.Nil
	c := make(chan *models.Organization, 1)
	go func() {
		<-ctx.Done()
		SubscriptionsMutateOrganizationResults.Delete(UUIDUser, subId)
	}()
	if _, ok := SubscriptionsMutateOrganizationResults.Load(UUIDUser); !ok {
		SubscriptionsMutateOrganizationResults.Insert(UUIDUser, make(map[uuid.UUID]organizationSub))
	}
	SubscriptionsMutateOrganizationResults.Subscription(UUIDUser, subId, organizationSub{
		Chanel:     c,
		SubContext: ctx,
	})
	return c, nil
}
