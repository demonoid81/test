package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/sphera-erp/sphera/internal/models"
)

func (r *mutationResolver) ParseAddress(ctx context.Context, rawAddress *string) ([]*string, error) {
	return r.addresses.ParseAddress(ctx, rawAddress)
}

func (r *mutationResolver) CountryMutation(ctx context.Context, country *models.Country) (*models.Country, error) {
	return r.addresses.CountryMutation(ctx, country)
}

func (r *mutationResolver) RegionMutation(ctx context.Context, region *models.Region) (*models.Region, error) {
	return r.addresses.RegionMutation(ctx, region)
}

func (r *mutationResolver) AreaMutation(ctx context.Context, area *models.Area) (*models.Area, error) {
	return r.addresses.AreaMutation(ctx, area)
}

func (r *mutationResolver) CityMutation(ctx context.Context, city *models.City) (*models.City, error) {
	return r.addresses.CityMutation(ctx, city)
}

func (r *mutationResolver) CityDistrictMutation(ctx context.Context, cityDistrict *models.CityDistrict) (*models.CityDistrict, error) {
	return r.addresses.CityDistrictMutation(ctx, cityDistrict)
}

func (r *mutationResolver) SettlementMutation(ctx context.Context, settlement *models.Settlement) (*models.Settlement, error) {
	return r.addresses.SettlementMutation(ctx, settlement)
}

func (r *mutationResolver) StreetMutation(ctx context.Context, street *models.Street) (*models.Street, error) {
	return r.addresses.StreetMutation(ctx, street)
}

func (r *mutationResolver) AddressMutation(ctx context.Context, address *models.Address) (*models.Address, error) {
	return r.addresses.AddressMutation(ctx, address)
}

func (r *queryResolver) Country(ctx context.Context, country *models.Country) (*models.Country, error) {
	return r.addresses.Country(ctx, country)
}

func (r *queryResolver) Countries(ctx context.Context, country *models.Country, offset *int, limit *int) ([]*models.Country, error) {
	return r.addresses.Countries(ctx, country, offset, limit)
}

func (r *queryResolver) Region(ctx context.Context, region *models.Region) (*models.Region, error) {
	return r.addresses.Region(ctx, region)
}

func (r *queryResolver) Regions(ctx context.Context, region *models.Region, offset *int, limit *int) ([]*models.Region, error) {
	return r.addresses.Regions(ctx, region, offset, limit)
}

func (r *queryResolver) Area(ctx context.Context, area *models.Area) (*models.Area, error) {
	return r.addresses.Area(ctx, area)
}

func (r *queryResolver) Areas(ctx context.Context, area *models.Area, offset *int, limit *int) ([]*models.Area, error) {
	return r.addresses.Areas(ctx, area, offset, limit)
}

func (r *queryResolver) City(ctx context.Context, city *models.City) (*models.City, error) {
	return r.addresses.City(ctx, city)
}

func (r *queryResolver) Cities(ctx context.Context, city *models.City, offset *int, limit *int) ([]*models.City, error) {
	return r.addresses.Cities(ctx, city, offset, limit)
}

func (r *queryResolver) CityDistrict(ctx context.Context, cityDistrict *models.CityDistrict) (*models.CityDistrict, error) {
	return r.addresses.CityDistrict(ctx, cityDistrict)
}

func (r *queryResolver) CityDistricts(ctx context.Context, cityDistrict *models.CityDistrict, offset *int, limit *int) ([]*models.CityDistrict, error) {
	return r.addresses.CityDistricts(ctx, cityDistrict, offset, limit)
}

func (r *queryResolver) Settlement(ctx context.Context, settlement *models.Settlement) (*models.Settlement, error) {
	return r.addresses.Settlement(ctx, settlement)
}

func (r *queryResolver) Settlements(ctx context.Context, settlement *models.Settlement, offset *int, limit *int) ([]*models.Settlement, error) {
	return r.addresses.Settlements(ctx, settlement, offset, limit)
}

func (r *queryResolver) Street(ctx context.Context, street *models.Street) (*models.Street, error) {
	return r.addresses.Street(ctx, street)
}

func (r *queryResolver) Streets(ctx context.Context, street *models.Street, offset *int, limit *int) ([]*models.Street, error) {
	return r.addresses.Streets(ctx, street, offset, limit)
}

func (r *queryResolver) Address(ctx context.Context, address *models.Address) (*models.Address, error) {
	return r.addresses.Address(ctx, address)
}

func (r *queryResolver) Addresses(ctx context.Context, address *models.Address, offset *int, limit *int) ([]*models.Address, error) {
	return r.addresses.Addresses(ctx, address, offset, limit)
}

func (r *subscriptionResolver) RegionSub(ctx context.Context) (<-chan *models.Region, error) {
	return r.addresses.RegionSub(ctx)
}

func (r *subscriptionResolver) AreaSub(ctx context.Context) (<-chan *models.Area, error) {
	return r.addresses.AreaSub(ctx)
}

func (r *subscriptionResolver) CitySub(ctx context.Context) (<-chan *models.City, error) {
	return r.addresses.CitySub(ctx)
}
