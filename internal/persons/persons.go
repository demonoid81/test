package persons

import (
	"context"
	"fmt"
	"sync"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/app"
	"github.com/sphera-erp/sphera/internal/middleware"
	"github.com/sphera-erp/sphera/internal/models"
	"github.com/sphera-erp/sphera/internal/utils"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type Resolver struct {
	env *app.App
	Resolvers
}

type Resolvers interface {
	PersonaMutation(ctx context.Context, person *models.Person) (*models.Person, error)
	Person(ctx context.Context, person models.Person) (*models.Person, error)
	Persons(ctx context.Context, person *models.Person, filter *models.PersonFilter, sort []models.PersonSort, offset *int, limit *int) ([]*models.Person, error)
	ParsePerson(ctx context.Context, photo *graphql.Upload, passport *graphql.Upload) (*models.PersonValidateStatus, error)
	ParsePersonSub(ctx context.Context, user *uuid.UUID) (<-chan *models.RecognizedFields, error)
	PersonCourses(ctx context.Context, course *models.PersonCourse) ([]*models.PersonCourse, error)
	PersonCourseMutation(ctx context.Context, course *models.PersonCourse) (*models.PersonCourse, error)
	PersonRatingMutation(ctx context.Context, personRating *models.PersonRating) (*models.PersonRating, error)
	PersonRating(ctx context.Context, personRating *models.PersonRating) (*models.PersonRating, error)
	PersonRatings(ctx context.Context, personRating *models.PersonRating) ([]*models.PersonRating, error)
	GetPersonRating(ctx context.Context, person models.Person) (*float64, error)
	GetMyRating(ctx context.Context) (*float64, error)
	PersonTax(ctx context.Context) (*models.Taxes, error)
	Agreement(ctx context.Context, incomeRegistration bool, taxPayment bool) (bool, error)
	RemoveContact(ctx context.Context, person *models.Person, contact *models.Contact) (bool, error)
	ReqToPartner(ctx context.Context) (bool, error)
	PersonSub(ctx context.Context) (<-chan *models.Person, error)
}

type channelParsePerson struct {
	mx           sync.RWMutex
	PersonResult map[uuid.UUID]chan *models.RecognizedFields
}

type subscriptionParsePerson struct {
	mx                sync.RWMutex
	parsePersonResult map[uuid.UUID]map[uuid.UUID]chan *models.RecognizedFields
}

var subscriptionsParsePerson subscriptionParsePerson

func init() {
	subscriptionsParsePerson = subscriptionParsePerson{
		parsePersonResult: make(map[uuid.UUID]map[uuid.UUID]chan *models.RecognizedFields),
	}
}

func (s *subscriptionParsePerson) Load(key uuid.UUID) (map[uuid.UUID]chan *models.RecognizedFields, bool) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	val, ok := s.parsePersonResult[key]
	return val, ok
}

func (s *subscriptionParsePerson) Insert(UUIDUser uuid.UUID, channel map[uuid.UUID]chan *models.RecognizedFields) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.parsePersonResult[UUIDUser] = channel
}

func (s *subscriptionParsePerson) Subscription(UUIDUser, subId uuid.UUID, result chan *models.RecognizedFields) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.parsePersonResult[UUIDUser][subId] = result
}

func (s *subscriptionParsePerson) Delete(UUIDUser, subId uuid.UUID) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	delete(s.parsePersonResult[UUIDUser], subId)
}

func NewPersonsResolvers(app *app.App) (*Resolver, error) {
	return &Resolver{
		env: app,
	}, nil
}

func (r *Resolver) ParsePerson(ctx context.Context, photo *graphql.Upload, passport *graphql.Upload) (*models.PersonValidateStatus, error) {
	result, err := ParsePerson(ctx, r.env, photo, passport)
	if err != nil {
		r.env.Logger.Error().Str("module", "persons").Str("func", "PersonaMutation").Err(err).Msg("Error commit transaction")
		return nil, gqlerror.Errorf("Error commit transaction")
	}
	return result, err
}

func (r *Resolver) PersonaMutation(ctx context.Context, person *models.Person) (*models.Person, error) {
	tx, err := r.env.Cockroach.BeginX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "persons").Str("func", "PersonaMutation").Err(err).Msg("Error run transaction")
		return nil, gqlerror.Errorf("Error run transaction")
	}
	defer tx.Rollback(ctx)

	columns := make(map[string]interface{})
	rows, _, err := person.Mutation(ctx, tx, r.env, nil, columns)
	if err != nil {
		r.env.Logger.Error().Str("module", "persons").Str("func", "PersonaMutation").Err(err).Msg("Error mutation person")
		return nil, err
	}
	person, err = person.ParseRow(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, tx)
	if err != nil {
		r.env.Logger.Error().Str("module", "persons").Str("func", "PersonaMutation").Err(err).Msg("Error commit transaction")
		return nil, gqlerror.Errorf("Error commit transaction")
	}
	err = tx.Commit(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "persons").Str("func", "PersonaMutation").Err(err).Msg("Error commit transaction")
		return nil, gqlerror.Errorf("Error commit transaction")
	}

	for _, c := range SubscriptionsMutatePersonResults.MutatePersonResults[uuid.Nil] {
		personSub := person
		if err := personSub.ParseRequestedFields(ctx, graphql.CollectFieldsCtx(c.SubContext, nil), r.env, r.env.Cockroach); err != nil {
			r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBookMutation").Err(err).Msg("Error parse row in medicalBook")
		}
		c.Chanel <- personSub
	}

	return person, nil
}

func (r *Resolver) Person(ctx context.Context, person models.Person) (*models.Person, error) {
	var err error
	sql := pglxqb.Select("persons.*").From("persons")
	result, sql, err := models.SqlGenSelectKeys(person, sql, "persons", 1)
	if err != nil {
		r.env.Logger.Error().Str("module", "persons").Str("func", "Person").Err(err).Msg("Error generate select relations")
		return nil, gqlerror.Errorf("Error generate select relations")
	}
	rows, err := sql.Where(pglxqb.Eq(result)).RunWith(r.env.Cockroach).QueryX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "persons").Str("func", "Person").Err(err).Msg("Error select person")
		return nil, gqlerror.Errorf("Error select person")
	}
	return person.ParseRow(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, r.env.Cockroach)
}

func (r *Resolver) Persons(ctx context.Context, person *models.Person, filter *models.PersonFilter, sort []models.PersonSort, offset *int, limit *int) ([]*models.Person, error) {
	var err error
	table := "persons"
	logger := r.env.Logger.Error().Str("module", "persons").Str("func", "Persons")
	sql := pglxqb.Select(fmt.Sprintf("%s.*", table)).From(table)
	if filter != nil {
		sql = utils.ReflectFilter(table, sql, filter)
	} else if person != nil {
		result, sql, err := models.SqlGenSelectKeys(person, sql, table, 1)
		if err != nil {
			logger.Err(err).Msg("Error generate select relations")
			return nil, gqlerror.Errorf("Error generate select relations")
		}
		if len(result) > 0 {
			sql = sql.Where(pglxqb.Eq(result))
		}
	}
	if len(sort) > 0 {
		for _, sortItem := range sort {
			sql = sql.OrderBy(fmt.Sprintf("%s.%s %s", table, sortItem.Field, sortItem.Order))
		}
	}
	if limit != nil {
		sql = sql.Limit(uint64(*limit))
	}
	if offset != nil {
		sql = sql.Offset(uint64(*offset))
	}
	rows, err := sql.RunWith(r.env.Cockroach).QueryX(ctx)
	if err != nil {
		logger.Err(err).Msg("Error select persons")
		return nil, gqlerror.Errorf("Error select persons")
	}
	return person.ParseRows(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, r.env.Cockroach)
}

func (r *Resolver) ParsePersonSub(ctx context.Context) (<-chan *models.RecognizedFields, error) {
	UUIDUser, err := middleware.ExtractUserInTokenMetadata(ctx, r.env)
	fmt.Println("-------------------", UUIDUser, "-------------------")
	if err != nil {
		r.env.Logger.Error().Str("module", "persons").Str("func", "ParsePersonSub").Err(err).Msg("Error get user in token metadata")
		return nil, gqlerror.Errorf("Error get user in token metadata")
	}
	subId := uuid.New()
	c := make(chan *models.RecognizedFields, 1)
	go func() {
		<-ctx.Done()
		subscriptionsParsePerson.Delete(UUIDUser, subId)
	}()
	if _, ok := subscriptionsParsePerson.Load(UUIDUser); !ok {
		subscriptionsParsePerson.Insert(UUIDUser, make(map[uuid.UUID]chan *models.RecognizedFields))
	}
	subscriptionsParsePerson.Subscription(UUIDUser, subId, c)
	return c, nil
}

// UUIDUser, _ := uuid.Parse("a2227dca-3602-4666-abcf-28903ea89231")
// subId := uuid.New()
// fmt.Println("Подписался")
// c := make(chan *models.ParsePersonResult, 1)
// go func() {
// 	<-ctx.Done()
// 	delete(SubscriptionsParsePerson[UUIDUser], subId)
// }()
// if SubscriptionsParsePerson[UUIDUser] == nil {
// 	SubscriptionsParsePerson[UUIDUser] = make(map[uuid.UUID]chan *models.ParsePersonResult, 0)
// }
// SubscriptionsParsePerson[UUIDUser][subId] = c
// return c, nil

// for _, c := range SubscriptionsParsePerson[UUIDUser] {
// 	c <- res
// }
