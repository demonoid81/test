package addresses

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/sphera-erp/sphera/app"
	"github.com/sphera-erp/sphera/internal/models"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type Resolver struct {
	env *app.App
	Resolvers
}

type Resolvers interface {
	ParseAddress(ctx context.Context, rawAddress *string) (*models.Address, error)
	CountryMutation(ctx context.Context, country *models.Country) (*models.Country, error)
	RegionMutation(ctx context.Context, region *models.Region) (*models.Region, error)
	AreaMutation(ctx context.Context, area *models.Area) (*models.Area, error)
	CityMutation(ctx context.Context, city *models.City) (*models.City, error)
	CityDistrictMutation(ctx context.Context, cityDistrict *models.CityDistrict) (*models.CityDistrict, error)
	SettlementMutation(ctx context.Context, settlement *models.Settlement) (*models.Settlement, error)
	StreetMutation(ctx context.Context, street *models.Street) (*models.Street, error)
	AddressMutation(ctx context.Context, address *models.Address) (*models.Address, error)
	//
	Country(ctx context.Context, country *models.Country) (*models.Country, error)
	Countries(ctx context.Context, country *models.Country, offset *int, limit *int) ([]*models.Country, error)
	Region(ctx context.Context, region *models.Region) (*models.Region, error)
	Regions(ctx context.Context, region *models.Region, offset *int, limit *int) ([]*models.Region, error)
	Area(ctx context.Context, area *models.Area) (*models.Area, error)
	Areas(ctx context.Context, area *models.Area, offset *int, limit *int) ([]*models.Area, error)
	City(ctx context.Context, city *models.City) (*models.City, error)
	Cities(ctx context.Context, city *models.City, offset *int, limit *int) ([]*models.City, error)
	CityDistrict(ctx context.Context, cityDistrict *models.CityDistrict) (*models.CityDistrict, error)
	CityDistricts(ctx context.Context, cityDistrict *models.CityDistrict, offset *int, limit *int) ([]*models.CityDistrict, error)
	Settlement(ctx context.Context, settlement *models.Settlement) (*models.Settlement, error)
	Settlements(ctx context.Context, settlement *models.Settlement, offset *int, limit *int) ([]*models.Settlement, error)
	Street(ctx context.Context, street *models.Street) (*models.Street, error)
	Streets(ctx context.Context, street *models.Street, offset *int, limit *int) ([]*models.Street, error)
	Address(ctx context.Context, address *models.Address) (*models.Address, error)
	Addresses(ctx context.Context, address *models.Address, offset *int, limit *int) ([]*models.Address, error)

	RegionSub(ctx context.Context) (<-chan *models.Region, error)
	AreaSub(ctx context.Context) (<-chan *models.Area, error)
	CitySub(ctx context.Context) (<-chan *models.City, error)
}

func NewAddressesResolvers(app *app.App) (*Resolver, error) {
	return &Resolver{
		env: app,
	}, nil
}

func (r *Resolver) ParseAddress(ctx context.Context, rawAddress *string) ([]*string, error) {
	return parser(*rawAddress, r.env)
}

func (r *Resolver) AddressMutation(ctx context.Context, address *models.Address) (*models.Address, error) {
	logger := r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBookMutation")
	var err error
	tx, err := r.env.Cockroach.BeginX(ctx)
	if err != nil {
		logger.Err(err).Msg("Error run transaction")
		return nil, gqlerror.Errorf("Error run transaction")
	}
	defer tx.Rollback(ctx)
	columns := make(map[string]interface{})
	rows, _, err := address.Mutation(ctx, tx, r.env, nil, columns)
	if err != nil {
		logger.Err(err).Msg("Error mutation medicalBook")
		return nil, err
	}
	address, err = address.ParseRow(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, tx)
	if err != nil {
		logger.Err(err).Msg("Error parse row in medicalBook")
		return nil, gqlerror.Errorf("Error parse row in medicalBook")
	}
	err = tx.Commit(ctx)
	if err != nil {
		logger.Err(err).Msg("Error commit transaction")
		return nil, gqlerror.Errorf("Error commit transaction")
	}
	return address, err
}

func (r *Resolver) Address(ctx context.Context, address *models.Address) (*models.Address, error) {
	var err error
	sql := pglxqb.Select("addresses.*").From("addresses")
	result, sql, err := models.SqlGenSelectKeys(address, sql, "addresses", 1)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBook").Err(err).Msg("Error generate select relations")
		return nil, gqlerror.Errorf("Error generate select relations")
	}
	rows, err := sql.Where(pglxqb.Eq(result)).RunWith(r.env.Cockroach).QueryX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBook").Err(err).Msg("Error select medicalBook")
		return nil, gqlerror.Errorf("Error select medicalBooks")
	}
	return address.ParseRow(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, r.env.Cockroach)
}

func (r *Resolver) Addresses(ctx context.Context, address *models.Address, offset *int, limit *int) ([]*models.Address, error) {
	var err error
	sql := pglxqb.Select("addresses.*").From("addresses")
	result, sql, err := models.SqlGenSelectKeys(address, sql, "addresses", 1)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBooks").Err(err).Msg("Error generate select relations")
		return nil, gqlerror.Errorf("Error generate select relations")
	}
	sql = sql.Where(pglxqb.Eq(result))
	if limit != nil {
		sql = sql.Limit(uint64(*limit))
	}
	if offset != nil {
		sql = sql.Offset(uint64(*offset))
	}
	rows, err := sql.RunWith(r.env.Cockroach).QueryX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBooks").Err(err).Msg("Error select medicalBooks")
		return nil, gqlerror.Errorf("Error select medicalBooks")
	}
	return address.ParseRows(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, r.env.Cockroach)
}
