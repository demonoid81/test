package models

import (
	"context"
	"fmt"
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

type City struct {
	UUID        *uuid.UUID `json:"uuid" db:"uuid" auto:"false"`
	Name        *string    `json:"name" db:"city"`
	Created     *time.Time `json:"created" db:"created"`
	Updated     *time.Time `json:"updated" db:"updated"`
	UUIDCountry *uuid.UUID `db:"uuid_country"`
	Country     *Country   `json:"country" relay:"uuid_country" link:"UUIDCountry"`
	UUIDRegion  *uuid.UUID `db:"uuid_region"`
	Region      *Region    `json:"region" relay:"uuid_region" link:"UUIDRegion"`
	UUIDArea    *uuid.UUID `db:"uuid_area"`
	Area        *Area      `json:"area" relay:"uuid_area" link:"UUIDArea"`
	IsDeleted   *bool      `json:"isDeleted" db:"is_deleted"`
}

type CityFilter struct {
	UUID      *UUIDFilter     `json:"uuid" db:"uuid"`
	Region    *RegionFilter   `json:"region" table:"regions" link:"uuid_region"`
	Area      *AreaFilter     `json:"area" table:"areas" link:"uuid_area"`
	Name      *StringFilter   `json:"name" db:"name"`
	Created   *DateTimeFilter `json:"created" db:"created"`
	Updated   *DateTimeFilter `json:"updated" db:"updated"`
	IsDeleted *bool           `json:"isDeleted" db:"is_deleted"`
}

func (c *City) Mutation(ctx context.Context, db pglxqb.BaseRunner, app *app.App, rColumns interface{}, columns map[string]interface{}) (*pglx.Rows, *uuid.UUID, error) {
	// если есть uuid значит манипулируем обектом
	if c.UUID != nil {
		// у на только полу UUID, то ничего не меняем отдаем ссылку
		if utils.CountFillFields(c) == 1 && len(columns) == 0 {
			return nil, c.UUID, nil
		}
		// получим обьект
		city, err := c.GetByUUID(ctx, app, db, c.UUID)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "Mutation").Err(err).Msg("Error get person")
			return nil, nil, gqlerror.Errorf("Error get person")
		}
		// востановим все ссылки
		utils.RestoreUUID(c, city)
		// востановим подчиненные ссылки на структуры
		if err = city.restoreStruct(ctx, app, db); err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "Mutation").Err(err).Msg("Error restore struct person")
			return nil, nil, gqlerror.Errorf("Error restore struct person")
		}
	} else {
		// иначе создадим с нуля Обьект
		newUUID := uuid.New()
		c.UUID = &newUUID
		columns["uuid"] = newUUID
	}
	// Создадим родственные связи
	parent := make(map[string]interface{})
	// мутируем дополниетельные структуры если они есть
	setColumns, err := SqlGenKeys(ctx, app, db, c, columns, parent)
	if err != nil {
		app.Logger.Error().Str("module", "persons").Str("func", "Mutation").Err(err).Msg("Error generate map of keys")
		return nil, nil, err
	}
	// Удалим лишние поля
	setColumns = utils.ClearSQLFields(c, setColumns)
	//
	if len(columns) > 0 {
		fmt.Println("Mutation City Insert")
		// Обновляем иначе
		rows, err := pglxqb.Insert("cities").
			SetMap(columns).
			OnConflictUpdateMap(columns, "uuid").
			Suffix(utils.PrepareSuffix(rColumns)).
			RunWith(db).QueryX(ctx)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "Mutation").Err(err).Msg("Error insert city")
			return nil, nil, gqlerror.Errorf("Error insert city")
		}
		return rows, c.UUID, nil
	}
	return nil, c.UUID, nil
}

func (c *City) ParseRows(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) ([]*City, error) {
	var cities []*City
	defer rows.Close()
	for rows.Next() {
		var city City
		err := rows.StructScan(&city)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct city")
			return nil, gqlerror.Errorf("Error scan response to struct cityDistrict")
		}
		err = city.parseRequestedFields(ctx, fields, app, db)
		if err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
		cities = append(cities, &city)
	}
	return cities, nil
}

func (c *City) ParseRow(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) (*City, error) {
	var err error
	var city City
	defer rows.Close()
	for rows.Next() {
		err = rows.StructScan(&city)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct cityDistrict")
			return nil, gqlerror.Errorf("Error scan response to struct cityDistrict")
		}
	}
	err = city.parseRequestedFields(ctx, fields, app, db)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return &city, nil
}

func (c *City) parseRequestedFields(ctx context.Context, fields []graphql.CollectedField, app *app.App, db pglxqb.BaseRunner) error {
	return parseRequestedFields(ctx, app, db, fields, c)
}

func (c *City) restoreStruct(ctx context.Context, app *app.App, db pglxqb.BaseRunner) error {
	v := reflect.ValueOf(c)
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

func (c *City) GetByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID) (*City, error) {
	rows, err := pglxqb.SelectAll().From("cities").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error get person from DB")
		return nil, gqlerror.Errorf("Error get person from DB")
	}
	var city City
	defer rows.Close()
	for rows.Next() {
		if err := rows.StructScan(&city); err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct Person")
			return nil, gqlerror.Errorf("Error scan response to struct Person")
		}
	}
	return &city, nil
}

func (c *City) GetParsedObjectByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID, column graphql.CollectedField) (*City, error) {
	rows, err := pglxqb.SelectAll().From("cities").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return c.ParseRow(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}

func (c *City) GetParsedObjectsByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid []*uuid.UUID, column graphql.CollectedField) ([]*City, error) {
	rows, err := pglxqb.SelectAll().From("cities").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return c.ParseRows(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}

func parserCity(suggestion *model.Address, address *Address, app *app.App, ctx context.Context) (err error) {
	if suggestion.Region == "Москва" {
		if suggestion.RegionFiasID != "" {
			suggestion.CityFiasID = suggestion.RegionFiasID
		}
		suggestion.City = suggestion.Region
	}
	city := new(City)
	cityUUID := uuid.New()
	if suggestion.CityFiasID != "" {
		cityUUID, err = uuid.Parse(suggestion.CityFiasID)
		if err != nil {
			app.Logger.Err(err).Msg("Error parse uuid city")
			return gqlerror.Errorf("Error parse uuid city")
		}
	}
	var cityDB City
	sql := pglxqb.Select("uuid, city").
		From("cities").
		Where(pglxqb.Eq{"city": suggestion.City})

	if address.Region != nil && address.Region.UUID != nil {
		sql = sql.Where(pglxqb.Eq{"uuid_region": address.Region.UUID})
	}

	if address.Area != nil && address.Area.UUID != nil {
		sql = sql.Where(pglxqb.Eq{"uuid_area": address.Area.UUID})
	}

	rows, err := sql.RunWith(app.Cockroach).
		QueryX(ctx)
	if err != nil {
		app.Logger.Err(err).Msg("Error get city from db")
		return gqlerror.Errorf("Error get city from db")
	}
	for rows.Next() {
		if err := rows.StructScan(&cityDB); err != nil {
			app.Logger.Err(err).Msg("Error scan response to struct City")
			return gqlerror.Errorf("Error scan response to struct City")
		}
	}
	// у нас уже записан такой регион
	if cityDB.UUID != nil {
		suggestion.CityFiasID = cityDB.UUID.String()
		city.UUID = cityDB.UUID
		address.City = city
		return nil
	}

	city.UUID = &cityUUID
	city.Name = &suggestion.City
	city.Region = address.Region
	city.Area = address.Area
	address.City = city

	return nil
}
