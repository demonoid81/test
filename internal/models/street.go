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

type Street struct {
	UUID             *uuid.UUID    `json:"uuid" db:"uuid" auto:"false"`
	Name             *string       `json:"name" db:"street"`
	Created          *time.Time    `json:"created" db:"created"`
	Updated          *time.Time    `json:"updated" db:"updated"`
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
	IsDeleted        *bool         `json:"isDeleted" db:"is_deleted"`
}

type StreetFilter struct {
	UUID         *UUIDFilter         `json:"uuid" db:"uuid"`
	Region       *RegionFilter       `json:"region" table:"regions" link:"uuid_region"`
	Area         *AreaFilter         `json:"area" table:"areas" link:"uuid_area"`
	City         *CityFilter         `json:"city" table:"cities" link:"uuid_city"`
	CityDistrict *CityDistrictFilter `json:"cityDistrict" table:"city_districts" link:"uuid_city_district"`
	Settlement   *SettlementFilter   `json:"settlement" table:"settlements" link:"uuid_settlement"`
	Name         *StringFilter       `json:"name" db:"name"`
	Created      *DateTimeFilter     `json:"created" db:"created"`
	Updated      *DateTimeFilter     `json:"updated" db:"updated"`
	IsDeleted    *bool               `json:"isDeleted" db:"is_deleted"`
}

func (s *Street) Mutation(ctx context.Context, db pglxqb.BaseRunner, app *app.App, rColumns interface{}, columns map[string]interface{}) (*pglx.Rows, *uuid.UUID, error) {
	fmt.Println(columns)
	// если есть uuid значит манипулируем обектом
	if s.UUID != nil {
		if utils.CountFillFields(s) == 1 && len(columns) == 0 {
			return nil, s.UUID, nil
		}
		//street, err := s.GetByUUID(ctx, app, db, s.UUID)
		//if err != nil {
		//	app.Logger.Error().Str("module", "models").Str("func", "Mutation").Err(err).Msg("Error get person")
		//	return nil, nil, gqlerror.Errorf("Error get person")
		//}
		//// востановим все ссылки
		//utils.RestoreUUID(s, street)
		//// востановим подчиненные структуры
		//if err = street.restoreStruct(ctx, app, db); err != nil {
		//	app.Logger.Error().Str("module", "users").Str("func", "Mutation").Err(err).Msg("Error restore struct person")
		//	return nil, nil, gqlerror.Errorf("Error restore struct person")
		//}
	} else {
		// иначе создадим с нуля Обьект
		newUUID := uuid.New()
		s.UUID = &newUUID
		columns["uuid"] = newUUID
	}
	// дополним пропущеные поля, если они есть
	parent := make(map[string]interface{})
	setColumns, err := SqlGenKeys(ctx, app, db, s, columns, parent)
	if err != nil {
		app.Logger.Error().Str("module", "persons").Str("func", "Mutation").Err(err).Msg("Error generate map of keys")
		return nil, nil, err
	}
	// только одна колонка, и это uuid то удаляем обьект
	setColumns = utils.ClearSQLFields(s, setColumns)
	if len(setColumns) > 0 {

		// Обновляем иначе
		rows, err := pglxqb.Insert("streets").
			SetMap(setColumns).
			OnConflictUpdateMap(setColumns, "uuid").
			Suffix(utils.PrepareSuffix(rColumns)).
			RunWith(db).QueryX(ctx)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "Mutation").Err(err).Msg("Error insert street")
			return nil, nil, gqlerror.Errorf("Error insert street")
		}
		return rows, s.UUID, nil
	}

	return nil, s.UUID, nil
}

func (s *Street) ParseRows(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) ([]*Street, error) {
	var streets []*Street
	defer rows.Close()
	for rows.Next() {
		var street Street
		err := rows.StructScan(&street)
		if err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct street")
			return nil, gqlerror.Errorf("Error scan response to struct street")
		}
		err = street.parseRequestedFields(ctx, fields, app, db)
		if err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
		streets = append(streets, &street)
	}
	return streets, nil
}

func (s *Street) ParseRow(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) (*Street, error) {
	var err error
	var street Street
	defer rows.Close()
	for rows.Next() {
		err = rows.StructScan(&street)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct street")
			return nil, gqlerror.Errorf("Error scan response to struct street")
		}
	}
	err = street.parseRequestedFields(ctx, fields, app, db)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return &street, nil
}

func (s *Street) parseRequestedFields(ctx context.Context, fields []graphql.CollectedField, app *app.App, db pglxqb.BaseRunner) error {
	return parseRequestedFields(ctx, app, db, fields, s)
}

func (s *Street) restoreStruct(ctx context.Context, app *app.App, db pglxqb.BaseRunner) error {
	v := reflect.ValueOf(s)
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

func (s *Street) GetByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID) (*Street, error) {
	rows, err := pglxqb.SelectAll().From("streets").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error get person from DB")
		return nil, gqlerror.Errorf("Error get person from DB")
	}
	var street Street
	defer rows.Close()
	for rows.Next() {
		if err := rows.StructScan(&street); err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct Person")
			return nil, gqlerror.Errorf("Error scan response to struct Person")
		}
	}
	return &street, nil
}

func (s *Street) GetParsedObjectByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID, column graphql.CollectedField) (*Street, error) {
	rows, err := pglxqb.SelectAll().From("streets").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return s.ParseRow(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}

func (s *Street) GetParsedObjectsByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid []*uuid.UUID, column graphql.CollectedField) ([]*Street, error) {
	rows, err := pglxqb.SelectAll().From("streets").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return s.ParseRows(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}

func parserStreet(suggestion *model.Address, address *Address, app *app.App, ctx context.Context) (err error) {
	street := new(Street)
	streetUUID := uuid.New()
	if suggestion.StreetFiasID != "" {
		streetUUID, err = uuid.Parse(suggestion.StreetFiasID)
		if err != nil {
			app.Logger.Err(err).Msg("Error parse uuid street")
			return gqlerror.Errorf("Error parse uuid street")
		}
	}
	var streetDB Street
	sql := pglxqb.Select("uuid, street").
		From("streets").
		Where(pglxqb.Eq{"street": suggestion.Street})

	if address.Region != nil && address.Region.UUID != nil {
		sql = sql.Where(pglxqb.Eq{"uuid_region": address.Region.UUID})
	}

	if address.Area != nil && address.Area.UUID != nil {
		sql = sql.Where(pglxqb.Eq{"uuid_area": address.Area.UUID})
	}

	if address.City != nil && address.City.UUID != nil {
		sql = sql.Where(pglxqb.Eq{"uuid_city": address.City.UUID})
	}

	if address.CityDistrict != nil && address.CityDistrict.UUID != nil {
		sql = sql.Where(pglxqb.Eq{"uuid_city_district": address.CityDistrict.UUID})
	}

	if address.Settlement != nil && address.Settlement.UUID != nil {
		sql = sql.Where(pglxqb.Eq{"uuid_settlement": address.Settlement.UUID})
	}

	rows, err := sql.RunWith(app.Cockroach).QueryX(ctx)
	if err != nil {
		app.Logger.Err(err).Msg("Error get street from db")
		return gqlerror.Errorf("Error get street from db")
	}
	for rows.Next() {
		if err := rows.StructScan(&streetDB); err != nil {
			app.Logger.Err(err).Msg("Error scan response to struct Street")
			return gqlerror.Errorf("Error scan response to struct Street")
		}
	}
	// у нас уже записан такой регион
	if streetDB.UUID != nil {
		suggestion.StreetFiasID = streetDB.UUID.String()
		street.UUID = streetDB.UUID
		address.Street = street
		return nil
	}

	street.UUID = &streetUUID
	street.Name = &suggestion.Street
	street.Region = address.Region
	street.Area = address.Area
	street.City = address.City
	street.CityDistrict = address.CityDistrict
	street.Settlement = address.Settlement
	address.Street = street

	return nil
}
