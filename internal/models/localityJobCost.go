package models

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/app"
	"github.com/sphera-erp/sphera/internal/utils"
	"github.com/sphera-erp/sphera/pkg/pglx"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"reflect"
	"time"
)

type LocalityJobCost struct {
	UUID             *uuid.UUID    `json:"uuid" db:"uuid"`
	Created          *time.Time    `json:"created" db:"created"`
	Updated          *time.Time    `json:"updated" db:"updated"`
	UUIDOrganization *uuid.UUID    `db:"uuid_organization"`
	Organization     *Organization `json:"organization" relay:"uuid_organization" link:"UUIDOrganization"`
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
	MaxCost          *float64      `json:"maxCost" db:"max_cost"`
	IsDeleted        *bool         `json:"isDeleted" db:"is_deleted"`
}

type LocalityJobCostFilter struct {
	UUID         *UUIDFilter         `json:"uuid" db:"uuid"`
	Created      *DateTimeFilter     `json:"created" db:"created"`
	Updated      *DateTimeFilter     `json:"updated" db:"updated"`
	Organization *OrganizationFilter `json:"organization" table:"organizations" link:"uuid_organization"`
	Country      *CountryFilter      `json:"country" table:"countries" link:"uuid_country"`
	Region       *RegionFilter       `json:"region" table:"regions" link:"uuid_region"`
	Area         *AreaFilter         `json:"area" table:"areas" link:"uuid_area"`
	City         *CityFilter         `json:"city" table:"cities" link:"uuid_city"`
	CityDistrict *CityDistrictFilter `json:"cityDistrict" table:"city_districts" link:"uuid_city_district"`
	Settlement   *SettlementFilter   `json:"settlement" table:"settlements" link:"uuid_settlement"`
	MaxCost      *FloatFilter        `json:"maxCost" db:"max_cost"`
	IsDeleted    *bool               `json:"isDeleted" db:"is_deleted"`
}

func (ljc *LocalityJobCost) Mutation(ctx context.Context, db pglxqb.BaseRunner, app *app.App, rColumns interface{}, columns map[string]interface{}) (*pglx.Rows, *uuid.UUID, error) {
	updateOrDelete := false
	// если есть uuid значит манипулируем обектом
	if ljc.UUID != nil {
		localityJobCost, err := ljc.GetByUUID(ctx, app, db, ljc.UUID)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "Mutation").Err(err).Msg("Error get person")
			return nil, nil, gqlerror.Errorf("Error get person")
		}
		// востановим все ссылки
		utils.RestoreUUID(ljc, localityJobCost)
		// востановим подчиненные структуры
		if err = localityJobCost.restoreStruct(ctx, app, db); err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "Mutation").Err(err).Msg("Error restore struct person")
			return nil, nil, gqlerror.Errorf("Error restore struct person")
		}
		updateOrDelete = true
	} else {
		// иначе создадим с нуля Обьект
		newUUID := uuid.New()
		ljc.UUID = &newUUID
		columns["uuid"] = newUUID
	}
	// дополним пропущеные поля, если они есть
	parent := make(map[string]interface{})
	setColumns, err := SqlGenKeys(ctx, app, db, ljc, columns, parent)
	if err != nil {
		app.Logger.Error().Str("module", "persons").Str("func", "Mutation").Err(err).Msg("Error generate map of keys")
		return nil, nil, err
	}
	// только одна колонка, и это uuid то удаляем обьект
	setColumns = utils.ClearSQLFields(ljc, setColumns)
	if len(setColumns) > 0 {
		if updateOrDelete {
			// Обновляем иначе
			rows, err := pglxqb.Update("locality_job_costs").
				SetMap(setColumns).
				Where("uuid = ?", ljc.UUID).
				Suffix(utils.PrepareSuffix(rColumns)).
				RunWith(db).QueryX(ctx)
			if err != nil {
				app.Logger.Error().Str("module", "contact").Str("func", "Mutation").Err(err).Msg("Error update contact")
				return nil, nil, gqlerror.Errorf("Error update contact")
			}
			return rows, ljc.UUID, nil
		} else {
			rows, err := pglxqb.Insert("locality_job_costs").
				SetMap(setColumns).
				Suffix(utils.PrepareSuffix(rColumns)).
				RunWith(db).QueryX(ctx)
			if err != nil {
				app.Logger.Error().Str("module", "contact").Str("func", "Mutation").Err(err).Msg("Error insert contact")
				return nil, nil, gqlerror.Errorf("Error insert contact")
			}
			return rows, ljc.UUID, nil
		}
	}
	return nil, ljc.UUID, nil
}


func (ljc *LocalityJobCost) ParseRows(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) ([]*LocalityJobCost, error) {
	var localityJobCosts []*LocalityJobCost
	defer rows.Close()
	for rows.Next() {
		var localityJobCost LocalityJobCost
		err := rows.StructScan(&localityJobCost)
		if err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
		localityJobCosts = append(localityJobCosts, &localityJobCost)
	}
	for _, localityJobCost := range localityJobCosts {
		err := localityJobCost.parseRequestedFields(ctx, fields, app, db)
		if err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
	}
	return localityJobCosts, nil
}

func (ljc *LocalityJobCost) ParseRow(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) (*LocalityJobCost, error) {
	var err error
	var localityJobCost LocalityJobCost
	defer rows.Close()
	for rows.Next() {
		err = rows.StructScan(&localityJobCost)
		if err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
	}
	err = localityJobCost.parseRequestedFields(ctx, fields, app, db)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return &localityJobCost, nil
}

func (ljc *LocalityJobCost) parseRequestedFields(ctx context.Context, fields []graphql.CollectedField, app *app.App, db pglxqb.BaseRunner) error {
	return parseRequestedFields(ctx, app, db, fields, ljc)
}

func (ljc *LocalityJobCost) restoreStruct (ctx context.Context, app *app.App, db pglxqb.BaseRunner) error {
	v := reflect.ValueOf(ljc)
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

func (ljc *LocalityJobCost) GetByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID) (*LocalityJobCost, error) {
	rows, err := pglxqb.SelectAll().From("locality_job_costs").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error get person from DB")
		return nil, gqlerror.Errorf("Error get person from DB")
	}
	var localityJobCost LocalityJobCost
	defer rows.Close()
	for rows.Next() {
		if err := rows.StructScan(&localityJobCost); err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct Person")
			return nil, gqlerror.Errorf("Error scan response to struct Person")
		}
	}
	return &localityJobCost, nil
}

func (ljc *LocalityJobCost) GetParsedObjectByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID, column graphql.CollectedField) (*LocalityJobCost, error) {
	rows, err := pglxqb.SelectAll().From("locality_job_costs").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return ljc.ParseRow(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}

func (ljc *LocalityJobCost) GetParsedObjectsByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid []*uuid.UUID, column graphql.CollectedField) ([]*LocalityJobCost, error) {
	rows, err := pglxqb.SelectAll().From("locality_job_costs").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return ljc.ParseRows(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}