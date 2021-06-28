package models

import (
	"context"
	"reflect"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/ekomobile/dadata/v2/api/model"
	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/app"
	"github.com/sphera-erp/sphera/internal/utils"
	"github.com/sphera-erp/sphera/pkg/pglx"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type CityDistrict struct {
	UUID        *uuid.UUID `json:"uuid" db:"uuid" auto:"false"`
	Name        *string    `json:"name" db:"city_district"`
	Created     *time.Time `json:"created" db:"created"`
	Updated     *time.Time `json:"updated" db:"updated"`
	UUIDCountry *uuid.UUID `db:"uuid_country"`
	Country     *Country   `json:"country" relay:"uuid_country" link:"UUIDCountry"`
	UUIDRegion  *uuid.UUID `db:"uuid_region"`
	Region      *Region    `json:"region" relay:"uuid_region" link:"UUIDRegion"`
	UUIDArea    *uuid.UUID `db:"uuid_area"`
	Area        *Area      `json:"area" relay:"uuid_area" link:"UUIDArea"`
	UUIDCity    *uuid.UUID `db:"uuid_city"`
	City        *City      `json:"city" relay:"uuid_city" link:"UUIDCity"`
	IsDeleted   *bool      `json:"isDeleted" db:"is_deleted"`
}

type CityDistrictFilter struct {
	UUID      *UUIDFilter     `json:"uuid" db:"uuid"`
	Region    *RegionFilter   `json:"region" table:"regions" link:"uuid_region"`
	Area      *AreaFilter     `json:"area" table:"areas" link:"uuid_area"`
	City      *CityFilter     `json:"city" table:"cities" link:"uuid_city"`
	Name      *StringFilter   `json:"name" db:"name"`
	Created   *DateTimeFilter `json:"created" db:"created"`
	Updated   *DateTimeFilter `json:"updated" db:"updated"`
	IsDeleted *bool           `json:"isDeleted" db:"is_deleted"`
}

func (cd *CityDistrict) Mutation(ctx context.Context, db pglxqb.BaseRunner, app *app.App, rColumns interface{}, columns map[string]interface{}) (*pglx.Rows, *uuid.UUID, error) {

	// если есть uuid значит манипулируем обектом
	if cd.UUID != nil {
		if utils.CountFillFields(cd) == 1 && len(columns) == 0 {
			return nil, cd.UUID, nil
		}
		//	//cityDistrict, err := cd.GetByUUID(ctx, app, db, cd.UUID)
		//	//if err != nil {
		//	//	app.Logger.Error().Str("module", "models").Str("func", "Mutation").Err(err).Msg("Error get person")
		//	//	return nil, nil, gqlerror.Errorf("Error get person")
		//	//}
		//	//// востановим все ссылки
		//	//utils.RestoreUUID(cd, cityDistrict)
		//	//// востановим подчиненные структуры
		//	//if err = cityDistrict.restoreStruct(ctx, app, db); err != nil {
		//	//	app.Logger.Error().Str("module", "users").Str("func", "Mutation").Err(err).Msg("Error restore struct person")
		//	//	return nil, nil, gqlerror.Errorf("Error restore struct person")
		//	//}
		//} else {
		//	// иначе создадим с нуля Обьект
		//	newUUID := uuid.New()
		//	cd.UUID = &newUUID
		//	columns["uuid"] = newUUID
	}
	// дополним пропущеные поля, если они есть
	parent := make(map[string]interface{})
	setColumns, err := SqlGenKeys(ctx, app, db, cd, columns, parent)
	if err != nil {
		app.Logger.Error().Str("module", "persons").Str("func", "Mutation").Err(err).Msg("Error generate map of keys")
		return nil, nil, err
	}
	// только одна колонка, и это uuid то удаляем обьект
	setColumns = utils.ClearSQLFields(cd, setColumns)
	if len(setColumns) > 0 {
		// Обновляем иначе
		rows, err := pglxqb.Insert("city_districts").
			SetMap(setColumns).
			OnConflictUpdateMap(columns, "uuid").
			Suffix(utils.PrepareSuffix(rColumns)).
			RunWith(db).QueryX(ctx)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "Mutation").Err(err).Msg("Error insert cityDistrict")
			return nil, nil, gqlerror.Errorf("Error insert cityDistrict")
		}
		return rows, cd.UUID, nil
	}
	return nil, cd.UUID, nil
}

func (cd *CityDistrict) ParseRows(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) ([]*CityDistrict, error) {
	var cityDistricts []*CityDistrict
	defer rows.Close()
	for rows.Next() {
		var cityDistrict CityDistrict
		err := rows.StructScan(&cityDistrict)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct cityDistrict")
			return nil, gqlerror.Errorf("Error scan response to struct cityDistrict")
		}
		err = cityDistrict.parseRequestedFields(ctx, fields, app, db)
		if err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
		cityDistricts = append(cityDistricts, &cityDistrict)
	}
	return cityDistricts, nil
}

func (cd *CityDistrict) ParseRow(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) (*CityDistrict, error) {
	var err error
	var cityDistrict CityDistrict
	defer rows.Close()
	for rows.Next() {
		err = rows.StructScan(&cityDistrict)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct cityDistrict")
			return nil, gqlerror.Errorf("Error scan response to struct cityDistrict")
		}
	}
	err = cityDistrict.parseRequestedFields(ctx, fields, app, db)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return &cityDistrict, nil
}

func (cd *CityDistrict) parseRequestedFields(ctx context.Context, fields []graphql.CollectedField, app *app.App, db pglxqb.BaseRunner) error {
	return parseRequestedFields(ctx, app, db, fields, cd)
}

func (cd *CityDistrict) restoreStruct(ctx context.Context, app *app.App, db pglxqb.BaseRunner) error {
	v := reflect.ValueOf(cd)
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

func (cd *CityDistrict) GetByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID) (*CityDistrict, error) {
	rows, err := pglxqb.SelectAll().From("city_districts").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error get person from DB")
		return nil, gqlerror.Errorf("Error get person from DB")
	}
	var cityDistrict CityDistrict
	defer rows.Close()
	for rows.Next() {
		if err := rows.StructScan(&cityDistrict); err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct Person")
			return nil, gqlerror.Errorf("Error scan response to struct Person")
		}
	}
	return &cityDistrict, nil
}

func (cd *CityDistrict) GetParsedObjectByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID, column graphql.CollectedField) (*CityDistrict, error) {
	rows, err := pglxqb.SelectAll().From("city_districts").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return cd.ParseRow(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}

func (cd *CityDistrict) GetParsedObjectsByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid []*uuid.UUID, column graphql.CollectedField) ([]*CityDistrict, error) {
	rows, err := pglxqb.SelectAll().From("city_districts").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return cd.ParseRows(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}

func parserCityDistrict(suggestion *model.Address, address *Address, app *app.App, ctx context.Context) (err error) {
	cityDistrict := new(CityDistrict)
	cityDistrictUUID := uuid.New()
	if suggestion.CityDistrictFiasID != "" {
		cityDistrictUUID, err = uuid.Parse(suggestion.CityDistrictFiasID)
		if err != nil {
			app.Logger.Err(err).Msg("Error parse uuid cityDistrict")
			return gqlerror.Errorf("Error parse uuid cityDistrict")
		}
	}
	var cityDistrictDB CityDistrict
	sql := pglxqb.Select("uuid, city_district").
		From("city_districts").
		Where(pglxqb.Eq{"city_district": suggestion.CityDistrict})

	if address.Region != nil && address.Region.UUID != nil {
		sql = sql.Where(pglxqb.Eq{"uuid_region": address.Region.UUID})
	}

	if address.Area != nil && address.Area.UUID != nil {
		sql = sql.Where(pglxqb.Eq{"uuid_area": address.Area.UUID})
	}

	if address.City != nil && address.City.UUID != nil {
		sql = sql.Where(pglxqb.Eq{"uuid_city": address.City.UUID})
	}

	rows, err := sql.RunWith(app.Cockroach).QueryX(ctx)
	if err != nil {
		app.Logger.Err(err).Msg("Error get cityDistrict from db")
		return gqlerror.Errorf("Error get cityDistrict from db")
	}
	for rows.Next() {
		if err := rows.StructScan(&cityDistrictDB); err != nil {
			app.Logger.Err(err).Msg("Error scan response to struct CityDistrict")
			return gqlerror.Errorf("Error scan response to struct CityDistrict")
		}
	}
	// у нас уже записан такой регион
	if cityDistrictDB.UUID != nil {
		suggestion.CityDistrictFiasID = cityDistrictDB.UUID.String()
		cityDistrict.UUID = cityDistrictDB.UUID
		address.CityDistrict = cityDistrict
		return nil
	}

	cityDistrict.UUID = &cityDistrictUUID
	cityDistrict.Name = &suggestion.CityDistrict
	cityDistrict.Region = address.Region
	cityDistrict.Area = address.Area
	cityDistrict.City = address.City
	address.CityDistrict = cityDistrict

	return nil
}
