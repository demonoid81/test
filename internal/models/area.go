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

type Area struct {
	UUID        *uuid.UUID `json:"uuid" db:"uuid" auto:"false"`
	Name        *string    `json:"name" db:"area"`
	Created     *time.Time `json:"created" db:"created"`
	Updated     *time.Time `json:"updated" db:"updated"`
	UUIDCountry *uuid.UUID `db:"uuid_country"`
	Country     *Country   `json:"country" relay:"uuid_country" link:"UUIDCountry"`
	UUIDRegion  *uuid.UUID `db:"uuid_region"`
	Region      *Region    `json:"region" relay:"uuid_region" link:"UUIDRegion"`
	IsDeleted   *bool      `json:"isDeleted" db:"is_deleted"`
}

type AreaFilter struct {
	UUID      *UUIDFilter     `json:"uuid" db:"uuid"`
	Region    *RegionFilter   `json:"region" table:"regions" link:"uuid_region"`
	Name      *StringFilter   `json:"name" db:"name"`
	Created   *DateTimeFilter `json:"created" db:"created"`
	Updated   *DateTimeFilter `json:"updated" db:"updated"`
	IsDeleted *bool           `json:"isDeleted" db:"is_deleted"`
}

func (a *Area) Mutation(ctx context.Context, db pglxqb.BaseRunner, app *app.App, rColumns interface{}, columns map[string]interface{}) (*pglx.Rows, *uuid.UUID, error) {
	// если есть uuid значит манипулируем обектом
	if a.UUID != nil {
		if utils.CountFillFields(a) == 1 && len(columns) == 0 {
			return nil, a.UUID, nil
		}
		//area, err := a.GetByUUID(ctx, app, db, a.UUID)
		//if err != nil {
		//	app.Logger.Error().Str("module", "models").Str("func", "Mutation").Err(err).Msg("Error get person")
		//	return nil, nil, gqlerror.Errorf("Error get person")
		//}
		//// востановим все ссылки
		//utils.RestoreUUID(a, area)
		//// востановим подчиненные структуры
		//if err = area.restoreStruct(ctx, app, db); err != nil {
		//	app.Logger.Error().Str("module", "users").Str("func", "Mutation").Err(err).Msg("Error restore struct person")
		//	return nil, nil, gqlerror.Errorf("Error restore struct person")
		//}
	} else {
		// иначе создадим с нуля Обьект
		newUUID := uuid.New()
		a.UUID = &newUUID
		columns["uuid"] = newUUID
	}
	// дополним пропущеные поля, если они есть
	parent := make(map[string]interface{})
	setColumns, err := SqlGenKeys(ctx, app, db, a, columns, parent)
	if err != nil {
		app.Logger.Error().Str("module", "persons").Str("func", "Mutation").Err(err).Msg("Error generate map of keys")
		return nil, nil, err
	}
	// только одна колонка, и это uuid то удаляем обьект
	setColumns = utils.ClearSQLFields(a, setColumns)
	if len(setColumns) > 0 {
		rows, err := pglxqb.Insert("areas").
			SetMap(setColumns).
			OnConflictUpdateMap(setColumns, "uuid").
			Suffix(utils.PrepareSuffix(rColumns)).
			RunWith(db).QueryX(ctx)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "Mutation").Err(err).Msg("Error insert area")
			return nil, nil, gqlerror.Errorf("Error insert area")
		}
		return rows, a.UUID, nil
	}
	return nil, a.UUID, nil
}

func (a *Area) ParseRows(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) ([]*Area, error) {
	var areas []*Area
	defer rows.Close()
	for rows.Next() {
		var area Area
		err := rows.StructScan(&area)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct area")
			return nil, gqlerror.Errorf("Error scan response to struct area")
		}
		err = area.parseRequestedFields(ctx, fields, app, db)
		if err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
		areas = append(areas, &area)
	}
	return areas, nil
}

func (a *Area) ParseRow(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) (*Area, error) {
	var err error
	var area Area
	defer rows.Close()
	for rows.Next() {
		err = rows.StructScan(&area)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "manipulate").Err(err).Msg("Error scan response to struct area")
			return nil, gqlerror.Errorf("Error scan response to struct area")
		}
	}
	err = area.parseRequestedFields(ctx, fields, app, db)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return &area, nil
}

func (a *Area) parseRequestedFields(ctx context.Context, fields []graphql.CollectedField, app *app.App, db pglxqb.BaseRunner) error {
	return parseRequestedFields(ctx, app, db, fields, a)
}

func (a *Area) restoreStruct(ctx context.Context, app *app.App, db pglxqb.BaseRunner) error {
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

func (a *Area) GetByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID) (*Area, error) {
	rows, err := pglxqb.SelectAll().From("persons").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error get person from DB")
		return nil, gqlerror.Errorf("Error get person from DB")
	}
	var area Area
	defer rows.Close()
	for rows.Next() {
		if err := rows.StructScan(&area); err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct Person")
			return nil, gqlerror.Errorf("Error scan response to struct Person")
		}
	}
	return &area, nil
}

func (a *Area) GetParsedObjectByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID, column graphql.CollectedField) (*Area, error) {
	rows, err := pglxqb.SelectAll().From("persons").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return a.ParseRow(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}

func (a *Area) GetParsedObjectsByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid []*uuid.UUID, column graphql.CollectedField) ([]*Area, error) {
	rows, err := pglxqb.SelectAll().From("persons").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return a.ParseRows(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}

func parserArea(suggestion *model.Address, address *Address, app *app.App, ctx context.Context) (err error) {
	area := new(Area)
	areaUUID := uuid.New()
	if suggestion.AreaFiasID != "" {
		areaUUID, err = uuid.Parse(suggestion.AreaFiasID)
		if err != nil {
			app.Logger.Err(err).Msg("Error parse uuid area")
			return gqlerror.Errorf("Error parse uuid area")
		}
	}
	var areaDB Area
	sql := pglxqb.Select("uuid, area").
		From("areas").
		Where(pglxqb.Eq{"area": suggestion.Area})
	if address.Region != nil && address.Region.UUID != nil {
		sql = sql.Where(pglxqb.Eq{"uuid_region": address.Region.UUID})
	}
	rows, err := sql.RunWith(app.Cockroach).
		QueryX(ctx)
	if err != nil {
		app.Logger.Err(err).Msg("Error get area from db")
		return gqlerror.Errorf("Error get area from db")
	}
	for rows.Next() {
		if err := rows.StructScan(&areaDB); err != nil {
			app.Logger.Err(err).Msg("Error scan response to struct Area")
			return gqlerror.Errorf("Error scan response to struct Area")
		}
	}
	// у нас уже записан такой регион
	if areaDB.UUID != nil {
		suggestion.RegionFiasID = areaDB.UUID.String()
		area.UUID = areaDB.UUID
		address.Area = area
		return nil
	}
	area.UUID = &areaUUID
	area.Name = &suggestion.Area
	area.Region = address.Region
	address.Area = area

	return nil
}
