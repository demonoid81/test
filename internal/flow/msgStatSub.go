package flow

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/internal/middleware"
	"github.com/sphera-erp/sphera/internal/models"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type subscriptionMsgStatUpdate struct {
	mx            sync.RWMutex
	msgStatUpdate map[uuid.UUID]map[uuid.UUID]chan *models.MsgStat
}

var subscriptionsMsgStatUpdate subscriptionMsgStatUpdate

func init() {
	subscriptionsMsgStatUpdate = subscriptionMsgStatUpdate{
		msgStatUpdate: make(map[uuid.UUID]map[uuid.UUID]chan *models.MsgStat),
	}
}

func (s *subscriptionMsgStatUpdate) Load(key uuid.UUID) (map[uuid.UUID]chan *models.MsgStat, bool) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	val, ok := s.msgStatUpdate[key]
	return val, ok
}

func (s *subscriptionMsgStatUpdate) Insert(UUIDUser uuid.UUID, channel map[uuid.UUID]chan *models.MsgStat) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.msgStatUpdate[UUIDUser] = channel
}

func (s *subscriptionMsgStatUpdate) Subscription(UUIDUser, subId uuid.UUID, result chan *models.MsgStat) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.msgStatUpdate[UUIDUser][subId] = result
}

func (s *subscriptionMsgStatUpdate) Delete(UUIDUser, subId uuid.UUID) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	delete(s.msgStatUpdate[UUIDUser], subId)
}

func (r *Resolver) MsgStatSub(ctx context.Context) (<-chan *models.MsgStat, error) {
	UUIDUser, err := middleware.ExtractUserInTokenMetadata(ctx, r.env)
	if err != nil {
		r.env.Logger.Error().Str("module", "persons").Str("func", "ParsePersonSub").Err(err).Msg("Error get user in token metadata")
		return nil, gqlerror.Errorf("Error get user in token metadata")
	}
	fmt.Println("****************************************************** Подписка ****************************************************************")
	subId := uuid.New()
	c := make(chan *models.MsgStat, 1)
	go func() {
		<-ctx.Done()
		subscriptionsMsgStatUpdate.Delete(UUIDUser, subId)
	}()
	if _, ok := subscriptionsMsgStatUpdate.Load(UUIDUser); !ok {
		subscriptionsMsgStatUpdate.Insert(UUIDUser, make(map[uuid.UUID]chan *models.MsgStat))
	}
	subscriptionsMsgStatUpdate.Subscription(UUIDUser, subId, c)
	return c, nil
}
