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

type Settlement struct {
	UUID             *uuid.UUID    `json:"uuid" db:"uuid" auto:"false"`
	Name             *string       `json:"name" db:"settlement"`
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
	IsDeleted        *bool         `json:"isDeleted" db:"is_deleted"`
}

type SettlementFilter struct {
	UUID         *UUIDFilter         `json:"uuid" db:"uuid"`
	Region       *RegionFilter       `json:"region" table:"regions" link:"uuid_region"`
	Area         *AreaFilter         `json:"area" table:"areas" link:"area_area"`
	City         *CityFilter         `json:"city" table:"cities" link:"uuid_city"`
	CityDistrict *CityDistrictFilter `json:"cityDistrict" table:"city_districts" link:"uuid_city_district"`
	Name         *StringFilter       `json:"name" db:"name"`
	Created      *DateTimeFilter     `json:"created" db:"created"`
	Updated      *DateTimeFilter     `json:"updated" db:"updated"`
	IsDeleted    *bool               `json:"isDeleted" db:"is_deleted"`
}

func (st *Settlement) Mutation(ctx context.Context, db pglxqb.BaseRunner, app *app.App, rColumns interface{}, columns map[string]interface{}) (*pglx.Rows, *uuid.UUID, error) {

	// если есть uuid значит манипулируем обектом
	if st.UUID != nil {
		if utils.CountFillFields(st) == 1 && len(columns) == 0 {
			return nil, st.UUID, nil
		}
	} else {
		// иначе создадим с нуля Обьект
		newUUID := uuid.New()
		st.UUID = &newUUID
		columns["uuid"] = newUUID
	}
	// дополним пропущеные поля, если они есть
	parent := make(map[string]interface{})
	setColumns, err := SqlGenKeys(ctx, app, db, st, columns, parent)
	if err != nil {
		app.Logger.Error().Str("module", "persons").Str("func", "Mutation").Err(err).Msg("Error generate map of keys")
		return nil, nil, err
	}
	// только одна колонка, и это uuid то удаляем обьект
	setColumns = utils.ClearSQLFields(st, setColumns)
	if len(columns) > 0 {

		// Обновляем иначе
		rows, err := pglxqb.Insert("settlements").
			SetMap(setColumns).
			OnConflictUpdateMap(setColumns, "uuid").
			Suffix(utils.PrepareSuffix(rColumns)).
			RunWith(db).QueryX(ctx)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "Mutation").Err(err).Msg("Error insert settlement")
			return nil, nil, gqlerror.Errorf("Error insert settlement")
		}
		return rows, st.UUID, nil
	}
	return nil, st.UUID, nil
}

func (st *Settlement) ParseRows(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) ([]*Settlement, error) {
	var settlements []*Settlement
	defer rows.Close()
	for rows.Next() {
		var settlement Settlement
		err := rows.StructScan(&settlement)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct settlement")
			return nil, gqlerror.Errorf("Error scan response to struct settlement")
		}
		err = settlement.parseRequestedFields(ctx, fields, app, db)
		if err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
		settlements = append(settlements, &settlement)
	}
	return settlements, nil
}

func (st *Settlement) ParseRow(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) (*Settlement, error) {
	var err error
	var settlement Settlement
	defer rows.Close()
	for rows.Next() {
		err = rows.StructScan(&settlement)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct settlement")
			return nil, gqlerror.Errorf("Error scan response to struct settlement")
		}
	}
	err = settlement.parseRequestedFields(ctx, fields, app, db)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return &settlement, nil
}

func (st *Settlement) parseRequestedFields(ctx context.Context, fields []graphql.CollectedField, app *app.App, db pglxqb.BaseRunner) error {
	return parseRequestedFields(ctx, app, db, fields, st)
}

func (st *Settlement) restoreStruct(ctx context.Context, app *app.App, db pglxqb.BaseRunner) error {
	v := reflect.ValueOf(st)
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

func (st *Settlement) GetByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID) (*Settlement, error) {
	rows, err := pglxqb.SelectAll().From("settlements").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error get person from DB")
		return nil, gqlerror.Errorf("Error get person from DB")
	}
	var settlement Settlement
	defer rows.Close()
	for rows.Next() {
		if err := rows.StructScan(&settlement); err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct Person")
			return nil, gqlerror.Errorf("Error scan response to struct Person")
		}
	}
	return &settlement, nil
}

func (st *Settlement) GetParsedObjectByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID, column graphql.CollectedField) (*Settlement, error) {
	rows, err := pglxqb.SelectAll().From("settlements").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return st.ParseRow(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}

func (st *Settlement) GetParsedObjectsByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid []*uuid.UUID, column graphql.CollectedField) ([]*Settlement, error) {
	rows, err := pglxqb.SelectAll().From("settlements").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return st.ParseRows(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}

func parserSettlement(suggestion *model.Address, address *Address, app *app.App, ctx context.Context) (err error) {
	settlement := new(Settlement)
	settlementUUID := uuid.New()
	if suggestion.CityDistrictFiasID != "" {
		settlementUUID, err = uuid.Parse(suggestion.SettlementFiasID)
		if err != nil {
			app.Logger.Err(err).Msg("Error parse uuid settlement")
			return gqlerror.Errorf("Error parse uuid settlement")
		}
	}
	var settlementDB Settlement
	sql := pglxqb.Select("uuid, settlement").
		From("settlements").
		Where(pglxqb.Eq{"settlement": suggestion.Settlement})
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
	rows, err := sql.RunWith(app.Cockroach).QueryX(ctx)
	if err != nil {
		app.Logger.Err(err).Msg("Error get settlement from db")
		return gqlerror.Errorf("Error get settlement from db")
	}
	for rows.Next() {
		if err := rows.StructScan(&settlementDB); err != nil {
			app.Logger.Err(err).Msg("Error scan response to struct Settlement")
			return gqlerror.Errorf("Error scan response to struct Settlement")
		}
	}
	// у нас уже записан такой регион
	if settlementDB.UUID != nil {
		suggestion.SettlementFiasID = settlementDB.UUID.String()
		settlement.UUID = settlementDB.UUID
		address.Settlement = settlement
		return nil
	}

	settlement.UUID = &settlementUUID
	settlement.Name = &suggestion.City
	settlement.Region = address.Region
	settlement.Area = address.Area
	settlement.City = address.City
	settlement.CityDistrict = address.CityDistrict
	address.Settlement = settlement

	return nil
}
