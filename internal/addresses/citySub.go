package addresses

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/internal/models"
)

type channelMutateCity struct {
	mx         sync.RWMutex
	CityResult map[uuid.UUID]chan *models.City
}

type subscriptionMutateCityResults struct {
	mx               sync.RWMutex
	MutateCityResult map[uuid.UUID]map[uuid.UUID]chan *models.City
}

var subscriptionsMutateCityResults subscriptionMutateCityResults

func init() {
	subscriptionsMutateCityResults = subscriptionMutateCityResults{
		MutateCityResult: make(map[uuid.UUID]map[uuid.UUID]chan *models.City),
	}
}

func (c *subscriptionMutateCityResults) Load(key uuid.UUID) (map[uuid.UUID]chan *models.City, bool) {
	c.mx.RLock()
	defer c.mx.RUnlock()
	val, ok := c.MutateCityResult[key]
	return val, ok
}

func (c *subscriptionMutateCityResults) Insert(UUIDUser uuid.UUID, channel map[uuid.UUID]chan *models.City) {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.MutateCityResult[UUIDUser] = channel
}

func (c *subscriptionMutateCityResults) Subscription(UUIDUser, subId uuid.UUID, result chan *models.City) {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.MutateCityResult[UUIDUser][subId] = result
}

func (c *subscriptionMutateCityResults) Delete(UUIDUser, subId uuid.UUID) {
	c.mx.RLock()
	defer c.mx.RUnlock()
	delete(c.MutateCityResult[UUIDUser], subId)
}

func (r *Resolver) CitySub(ctx context.Context) (<-chan *models.City, error) {
	// UUIDUser, err := middleware.ExtractUserInTokenMetadata(ctx, r.env)
	// if err != nil {
	// 	r.env.Logger.Error().Str("module", "persons").Str("func", "OrganizationSub").Err(err).Msg("Error get user in token metadata")
	// 	return nil, gqlerror.Errorf("Error get user in token metadata")
	// }
	subId := uuid.New()

	UUIDUser := uuid.Nil
	c := make(chan *models.City, 1)
	go func() {
		<-ctx.Done()
		subscriptionsMutateCityResults.Delete(UUIDUser, subId)
	}()
	if _, ok := subscriptionsMutateCityResults.Load(UUIDUser); !ok {
		subscriptionsMutateCityResults.Insert(UUIDUser, make(map[uuid.UUID]chan *models.City))
	}
	subscriptionsMutateCityResults.Subscription(UUIDUser, subId, c)
	return c, nil
}
