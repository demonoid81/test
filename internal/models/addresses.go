package models

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/ekomobile/dadata/v2"
	"github.com/ekomobile/dadata/v2/api/model"
	"github.com/ekomobile/dadata/v2/client"
	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/app"
	"github.com/sphera-erp/sphera/internal/utils"
	"github.com/sphera-erp/sphera/pkg/pglx"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"reflect"
	"strconv"
	"time"
)

type Address struct {
	UUID             *uuid.UUID    `json:"uuid" db:"uuid" auto:"false"`
	UUIDCountry      *uuid.UUID    `db:"uuid_country"`
	Country          *Country      `json:"country" relay:"uuid_country" link:"UUIDCountry"`
	UUIDRegion       *uuid.UUID    `db:"uuid_region"`
	Region           *Region       `json:"region" relay:"uuid_region" link:"UUIDRegion"`
	UUIDArea         *uuid.UUID    `db:"uuid_area"`
	Area             *Area         `json:"area" relay:"uuid_area" link:"UUIDArea"`
	UUIDCity         *uuid.UUID    `db:"uuid_city"`
	City             *City         `json:"city" relay:"uuid_city" link:"UUIDCity"`
	UUIDCityDistrict *uuid.UUID    `db:"uuid_city_district"`
	CityDistrict     *CityDistrict `json:"cityDistrict" relay:"uuid_city_district" link:"UUIDCityDistrict"`
	UUIDSettlement   *uuid.UUID    `db:"uuid_settlement"`
	Settlement       *Settlement   `json:"settlement" relay:"uuid_settlement" link:"UUIDSettlement"`
	UUIDStreet       *uuid.UUID    `db:"uuid_street"`
	Street           *Street       `json:"street" relay:"uuid_street" link:"UUIDStreet"`
	House            *string       `json:"house" db:"house"`
	Block            *string       `json:"block" db:"block"`
	Flat             *string       `json:"flat" db:"flat"`
	FormattedAddress *string       `json:"formattedAddress" db:"formatted_address"`
	Lat              *float64      `json:"lat" db:"lat"`
	Lon              *float64      `json:"lon" db:"lon"`
	UUIDPerson       *uuid.UUID    `db:"uuid_person"`
	Person           *Person       `json:"person" relay:"uuid_person" link:"UUIDPerson"`
	UUIDOrganization *uuid.UUID    `db:"uuid_organization"`
	Organization     *Organization `json:"organization" relay:"uuid_organization" link:"UUIDOrganization"`
	Created          *time.Time    `json:"created" db:"created"`
	Updated          *time.Time    `json:"updated" db:"updated"`
	IsDeleted        *bool         `json:"isDeleted" db:"is_deleted"`
	GEOM             *string       `db:"geom"`
}

type AddressFilter struct {
	UUID             *UUIDFilter         `json:"uuid" db:"uuid"`
	FormattedAddress *StringFilter       `json:"formattedAddress" db:"formatted_address"`
	Country          *Country            `json:"country" table:"countries" link:"uuid_country"`
	Region           *Region             `json:"region" table:"regions" link:"uuid_region"`
	Area             *Area               `json:"area" table:"areas" link:"uuid_area"`
	City             *City               `json:"city" table:"cities" link:"uuid_city"`
	CityDistrict     *CityDistrict       `json:"cityDistrict" table:"city_districts" link:"uuid_city_district"`
	Settlement       *Settlement         `json:"settlement" table:"settlements" link:"uuid_settlement"`
	Street           *Street             `json:"street" table:"streets" link:"uuid_street"`
	House            *StringFilter       `json:"house" db:"house"`
	Block            *StringFilter       `json:"block" db:"block"`
	Flat             *StringFilter       `json:"flat" db:"flat"`
	Lat              *FloatFilter        `json:"lat" db:"lat"`
	Lon              *FloatFilter        `json:"lon" db:"lon"`
	Person           *PersonFilter       `json:"person" table:"persons" link:"uuid_person"`
	Organization     *OrganizationFilter `json:"organization" table:"organizations" link:"uuid_organization"`
	IsDeleted        *bool               `json:"isDeleted" db:"is_deleted"`
	And              []*AddressFilter    `json:"and"`
	Or               []*AddressFilter    `json:"or"`
	Not              *AddressFilter      `json:"not"`
}

func (a *Address) Mutation(ctx context.Context, db pglxqb.BaseRunner, app *app.App, rColumns interface{}, columns map[string]interface{}) (*pglx.Rows, *uuid.UUID, error) {
	var rawAddress *model.Address
	var err error
	if a.FormattedAddress != nil && *a.FormattedAddress != "" {
		rawAddress, err = GetAddressSuggestion(a, app)
		if err != nil {
			return nil, nil, err
		}
		fmt.Println(rawAddress)
		if err := PrepareStruct(rawAddress, a, app, ctx); err != nil {
			fmt.Println(err)
		}
	} else if a.UUID != nil {
		address, err := a.GetByUUID(ctx, app, db, a.UUID)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "Mutation").Err(err).Msg("Error get person")
			return nil, nil, gqlerror.Errorf("Error get person")
		}
		// востановим все ссылки
		utils.RestoreUUID(a, address)
		// востановим подчиненные структуры
		if err = a.restoreStruct(ctx, app, db); err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "Mutation").Err(err).Msg("Error restore struct person")
			return nil, nil, gqlerror.Errorf("Error restore struct person")
		}
		//update = true
	} else {
		return nil, nil, nil
	}
	//	// иначе создадим с нуля Обьект
	//	newUUID := uuid.New()
	//	a.UUID = &newUUID
	//	columns["uuid"] = newUUID
	//}
	//if err := addresses.ParseRawAddressToStruct(a, columns); err != nil {
	//	app.Logger.Error().Str("module", "persons").Str("func", "Mutation").Err(err).Msg("Error generate map of keys")
	//	return nil, err
	//}
	parent := make(map[string]interface{})
	// дополним пропущеные поля, если они есть
	setColumns, err := SqlGenKeys(ctx, app, db, a, columns, parent)
	if err != nil {
		app.Logger.Error().Str("module", "persons").Str("func", "Mutation").Err(err).Msg("Error generate map of keys")
		return nil, nil, err
	}
	fmt.Println(setColumns)
	// только одна колонка, и это uuid то удаляем обьект
	setColumns = utils.ClearSQLFields(a, setColumns)
	if len(setColumns) > 0 {
		rows, err := pglxqb.Insert("addresses").
			SetMap(setColumns).
			OnConflictUpdateMap(setColumns, "uuid").
			Suffix(utils.PrepareSuffix(rColumns)).
			RunWith(db).QueryX(ctx)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "Mutation").Err(err).Msg("Error insert address")
			return nil, nil, gqlerror.Errorf("Error insert address")
		}
		return rows, a.UUID, nil
	}
	return nil, a.UUID, nil
}

func (a *Address) ParseRows(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) ([]*Address, error) {
	var addresses []*Address
	for rows.Next() {
		var address Address
		if err := rows.StructScan(&address); err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
		addresses = append(addresses, &address)
	}
	for _, address := range addresses {
		if err := address.parseRequestedFields(ctx, fields, app, db); err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
	}
	return addresses, nil
}

func (a *Address) ParseRow(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) (*Address, error) {
	var address Address
	defer rows.Close()
	for rows.Next() {
		if err := rows.StructScan(&address); err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
	}
	if err := address.parseRequestedFields(ctx, fields, app, db); err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return &address, nil
}

func (a *Address) parseRequestedFields(ctx context.Context, fields []graphql.CollectedField, app *app.App, db pglxqb.BaseRunner) error {
	return parseRequestedFields(ctx, app, db, fields, a)
}

func (a *Address) restoreStruct(ctx context.Context, app *app.App, db pglxqb.BaseRunner) error {
	v := reflect.ValueOf(a)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if !v.IsValid() {
		return nil
	}
	for i := 0; i < v.NumField(); i++ {
		if err := restoreStructReflect(ctx, app, db, v, v.Field(i), v.Type().Field(i)); err != nil {
			return err
		}
	}
	return nil
}

func (a *Address) GetByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID) (*Address, error) {
	rows, err := pglxqb.SelectAll().From("addresses").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error get person from DB")
		return nil, gqlerror.Errorf("Error get person from DB")
	}
	var address Address
	for rows.Next() {
		if err := rows.StructScan(&address); err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct Person")
			return nil, gqlerror.Errorf("Error scan response to struct Person")
		}
	}
	return &address, nil
}

func (a *Address) GetParsedObjectByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID, column graphql.CollectedField) (*Address, error) {
	rows, err := pglxqb.SelectAll().From("addresses").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return a.ParseRow(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}

func (a *Address) GetParsedObjectsByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid []*uuid.UUID, column graphql.CollectedField) ([]*Address, error) {
	rows, err := pglxqb.SelectAll().From("addresses").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return a.ParseRows(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}

func GetAddressSuggestion(address *Address, app *app.App) (*model.Address, error) {
	var err error
	cred := client.Credentials{
		ApiKeyValue:    app.Cfg.Api.DadataApi,
		SecretKeyValue: app.Cfg.Api.DadataSecret,
	}
	Api := dadata.NewCleanApi(client.WithCredentialProvider(&cred))
	result, err := Api.Address(context.Background(), *address.FormattedAddress)
	if err != nil {
		return nil, err
	}
	var addressSuggestion *model.Address
	if len(result) > 1 {
		for _, s := range result {
			fmt.Println(s.Result, *address.FormattedAddress)
			if s.Result == *address.FormattedAddress {
				addressSuggestion = s
			}
		}
	} else {
		addressSuggestion = result[0]
	}

	return addressSuggestion, nil
}

func PrepareStruct(suggestion *model.Address, address *Address, app *app.App, ctx context.Context) error {
	var err error
	addressUUID := uuid.New()
	if suggestion.FiasID != "" {
		addressUUID, err = uuid.Parse(suggestion.FiasID)
		if err != nil {
			fmt.Println(err)
		}
	}
	if err != nil {
		fmt.Println(err)
	}
	address.UUID = &addressUUID
	address.FormattedAddress = &suggestion.Result
	if suggestion.House != "" {
		address.House = &suggestion.House
	}
	if suggestion.Block != "" {
		address.Block = &suggestion.Block
	}
	lat, err := strconv.ParseFloat(suggestion.GeoLat, 64)
	if err != nil {
		return err
	}
	address.Lat = &lat
	lon, err := strconv.ParseFloat(suggestion.GeoLon, 64)
	if err != nil {
		return err
	}
	address.Lon = &lon
	if suggestion.Region != "" {
		fmt.Printf("RegionFiasID: %s - %s\n", suggestion.RegionFiasID, suggestion.Region)
		if err = parserRegion(suggestion, address, app, ctx); err != nil {
			return err
		}
	}
	if suggestion.Area != "" {
		fmt.Printf("AreaFiasID: %s - %s\n", suggestion.AreaFiasID, suggestion.Area)
		if err = parserArea(suggestion, address, app, ctx); err != nil {
			return err
		}
	}
	if suggestion.City != "" || suggestion.Region == "Москва" {
		fmt.Printf("CityFiasID: %s - %s\n", suggestion.CityFiasID, suggestion.City)
		if err = parserCity(suggestion, address, app, ctx); err != nil {
			return err
		}
	}
	if suggestion.CityDistrict != "" {
		fmt.Printf("CityDistrictFiasID: %s - %s\n", suggestion.CityDistrictFiasID, suggestion.CityDistrict)
		if err = parserCityDistrict(suggestion, address, app, ctx); err != nil {
			return err
		}
	}
	if suggestion.Settlement != "" {
		fmt.Printf("SettlementFiasID: %s - %s\n", suggestion.SettlementFiasID, suggestion.Settlement)
		if err = parserSettlement(suggestion, address, app, ctx); err != nil {
			return err
		}
	}
	if suggestion.Street != "" {
		fmt.Printf("StreetFiasID: %s - %s\n", suggestion.StreetFiasID, suggestion.Street)
		if err = parserStreet(suggestion, address, app, ctx); err != nil {
			return err
		}
	}
	return nil
}
