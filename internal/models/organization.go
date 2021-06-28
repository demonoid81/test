package models

import (
	"context"
	"fmt"
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

type Organization struct {
	UUID                   *uuid.UUID      `json:"uuid" db:"uuid"`
	Created                *time.Time      `json:"created" db:"created"`
	Updated                *time.Time      `json:"updated" db:"updated"`
	Name                   *string         `json:"name" db:"name"`
	INN                    *string         `json:"inn" db:"inn"`
	KPP                    *string         `json:"kpp" db:"kpp"`
	UUIDAddressLegal       *uuid.UUID      `db:"uuid_address_legal"`
	AddressLegal           *Address        `json:"addressLegal" relay:"uuid_address_legal" link:"UUIDAddressLegal"`
	UUIDAddressFact        *uuid.UUID      `db:"uuid_address_fact"`
	AddressFact            *Address        `json:"addressFact" relay:"uuid_address_fact" link:"UUIDAddressFact"`
	UUIDDepartments        []*uuid.UUID    `db:"uuid_departments"`
	Departments            []*Organization `json:"departments" relay:"uuid_departments" link:"UUIDDepartments"`
	UUIDParentOrganization *uuid.UUID      `db:"uuid_parent_organization"`
	ParentOrganization     *Organization   `json:"parentOrganization" relay:"uuid_parent_organization" link:"UUIDParentOrganization"`
	UUIDParent             *uuid.UUID      `db:"uuid_parent"`
	Parent                 *Organization   `json:"parent" relay:"uuid_parent" link:"UUIDParent"`
	IsDeleted              *bool           `json:"isDeleted" db:"is_deleted"`
	UUIDLogo               *uuid.UUID      `db:"uuid_logo"`
	Logo                   *Content        `json:"logo" relay:"uuid_logo" link:"UUIDLogo"`
	Prefix                 *string         `json:"prefix" db:"prefix"`
	FullName               *string         `json:"fullName" db:"full_name"`
	ShortName              *string         `json:"shortName" db:"short_name"`
	Fee                    *float64        `json:"fee" db:"fee"`
	UUIDPersons            []*uuid.UUID    `db:"uuid_persons"`
	Persons                []*Person       `json:"persons" relay:"uuid_persons" link:"UUIDPersons"`
	IsGroup                *bool           `json:"isGroup" db:"is_group"`
	FirstReserveReward     *float64        `json:"firstReserveReward" db:"first_reserve_reward"`
	SecondReserveReward    *float64        `json:"secondReserveReward" db:"second_reserve_reward"`
	StDistance             *float64        `json:"stDistance" db:"st_distance"`
	StTime                 *time.Duration  `json:"stTime" db:"st_time"`
}

type OrganizationFilter struct {
	UUID                *UUIDFilter         `json:"uuid" db:"uuid"`
	Created             *DateTimeFilter     `json:"created" db:"created"`
	Updated             *DateTimeFilter     `json:"updated" db:"updated"`
	Name                *StringFilter       `json:"name" db:"name"`
	Inn                 *StringFilter       `json:"inn" db:"inn"`
	Kpp                 *StringFilter       `json:"kpp" db:"kpp"`
	AddressLegal        *AddressFilter      `json:"addressLegal" table:"addresses" link:"uuid_address_legal"`
	AddressFact         *AddressFilter      `json:"addressFact" table:"addresses" link:"uuid_address_fact"`
	Parent              *OrganizationFilter `json:"parent" table:"organizations" link:"uuid_organization"`
	ParentOrganization  *OrganizationFilter `json:"parentOrganization" table:"organizations" link:"uuid_parent_organization"`
	IsDeleted           *bool               `json:"isDeleted" db:"is_deleted"`
	Prefix              *StringFilter       `json:"prefix" db:"prefix"`
	FullName            *StringFilter       `json:"fullName" db:"full_name"`
	ShortName           *StringFilter       `json:"shortName" db:"short_name"`
	Fee                 *FloatFilter        `json:"fee" db:"fee"`
	FirstReserveReward  *FloatFilter        `json:"firstReserveReward" db:"first_reserve_reward"`
	SecondReserveReward *FloatFilter        `json:"secondReserveReward" db:"second_reserve_reward"`
}

func (o *Organization) Mutation(ctx context.Context, db pglxqb.BaseRunner, app *app.App, rColumns interface{}, columns map[string]interface{}) (*pglx.Rows, *uuid.UUID, error) {
	// Уберем лишние колонки
	if len(columns) > 0 {
		columns = utils.ClearSQLFields(o, columns)
	}
	isGroup := false
	update := false
	// если есть uuid значит манипулируем обектом
	if o.UUID != nil {
		if utils.CountFillFields(o) == 1 && len(columns) == 0 {
			return nil, o.UUID, nil
		}
		// Получим Объект
		organization, err := o.GetByUUID(ctx, app, db, o.UUID)

		if organization.IsGroup != nil && *organization.IsGroup {
			isGroup = true
			if utils.CountFillFields(o) == 1 {
				return nil, o.UUID, nil
			}
		}
		if err != nil {
			app.Logger.Error().Str("module", "organizations").Str("func", "Mutation").Err(err).Msg("Error get organizations")
			return nil, nil, gqlerror.Errorf("Error get organizations")
		}
		// Если не меняются родители то вернем uuid
		if Compare(organization, columns) && utils.CountFillFields(o) == 1 {
			return nil, o.UUID, nil
		}
		utils.RestoreUUID(o, organization)
		if err = o.restoreStruct(ctx, app, db); err != nil {
			app.Logger.Error().Str("module", "organizations").Str("func", "Mutation").Err(err).Msg("Error restore struct Organization")
			return nil, nil, gqlerror.Errorf("Error restore struct Organization")
		}
		update = true
	} else {
		// иначе создадим с нуля Объект
		newUUID := uuid.New()
		o.UUID = &newUUID
		_, err := pglxqb.Insert("organizations").
			Columns("uuid").
			Values(newUUID).
			RunWith(db).Exec(ctx)
		if err != nil {
			app.Logger.Error().Str("module", "organizations").Str("func", "Mutation").Err(err).Msg("Error insert user")
			return nil, nil, gqlerror.Errorf("Error insert user")
		}
		update = true
	}
	parentColumns := map[string]interface{}{}
	if o.IsGroup != nil {
		fmt.Println(*o.IsGroup)
		isGroup = *o.IsGroup
	}
	if !isGroup {
		fmt.Println("no Group")
		parentColumns = map[string]interface{}{"uuid_organization": o.UUID, "uuid_parent_organization": o.UUID}
	}

	// дополним пропущенные поля, если они есть
	setColumns, err := SqlGenKeys(ctx, app, db, o, columns, parentColumns)
	if err != nil {
		app.Logger.Error().Str("module", "persons").Str("func", "Mutation").Err(err).Msg("Error generate map of keys")
		return nil, nil, err
	}
	setColumns = utils.ClearSQLFields(o, setColumns)
	if len(setColumns) > 0 {
		if update {
			// todo Логика Обновления
			// Обновляем иначе
			rows, err := pglxqb.Update("organizations").
				SetMap(setColumns).
				Where("uuid = ?", o.UUID).
				Suffix(utils.PrepareSuffix(rColumns)).
				RunWith(db).QueryX(ctx)
			if err != nil {
				app.Logger.Error().Str("module", "organizations").Str("func", "manipulate").Err(err).Msg("Error update organization")
				return nil, nil, gqlerror.Errorf("Error update passport")
			}
			return rows, o.UUID, nil
		} else {
			// todo Логика вставки
			rows, err := pglxqb.Insert("organizations").
				SetMap(setColumns).
				Suffix(utils.PrepareSuffix(rColumns)).
				RunWith(db).QueryX(ctx)
			if err != nil {
				app.Logger.Error().Str("module", "organizations").Str("func", "manipulate").Err(err).Msg("Error insert organization")
				return nil, nil, gqlerror.Errorf("Error insert organization")
			}
			return rows, o.UUID, nil
		}
	}
	return nil, o.UUID, nil
}

func (o *Organization) ParseRows(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) ([]*Organization, error) {
	var organizations []*Organization
	defer rows.Close()
	for rows.Next() {
		var organization Organization
		err := rows.StructScan(&organization)
		if err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct person")
			return nil, gqlerror.Errorf("Error scan response to struct person")
		}
		organizations = append(organizations, &organization)
	}
	for _, organization := range organizations {
		if err := organization.ParseRequestedFields(ctx, fields, app, db); err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct person")
			return nil, gqlerror.Errorf("Error scan response to struct person")
		}
	}
	return organizations, nil
}

func (o *Organization) ParseRow(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) (*Organization, error) {
	var err error
	var organization Organization
	defer rows.Close()
	for rows.Next() {
		err = rows.StructScan(&organization)
		if err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct Organization")
			return nil, gqlerror.Errorf("Error scan response to struct Organization")
		}
	}
	err = organization.ParseRequestedFields(ctx, fields, app, db)
	if err != nil {
		app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct Organization")
		return nil, gqlerror.Errorf("Error scan response to struct Organization")
	}
	return &organization, nil
}

func (o *Organization) ParseRequestedFields(ctx context.Context, fields []graphql.CollectedField, app *app.App, db pglxqb.BaseRunner) error {
	return parseRequestedFields(ctx, app, db, fields, o)
}

func (o *Organization) restoreStruct(ctx context.Context, app *app.App, db pglxqb.BaseRunner) error {
	v := reflect.ValueOf(o)
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
	fmt.Println(v)
	return nil
}

func (o *Organization) GetByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID) (*Organization, error) {
	fmt.Println(uuid)
	rows, err := pglxqb.SelectAll().From("organizations").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error get person from DB")
		return nil, gqlerror.Errorf("Error get person from DB")
	}
	var organization Organization
	defer rows.Close()
	for rows.Next() {
		if err := rows.StructScan(&organization); err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct Person")
			return nil, gqlerror.Errorf("Error scan response to struct Person")
		}
	}
	fmt.Println(organization)
	return &organization, nil
}

func (o *Organization) GetParsedObjectByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID, column graphql.CollectedField) (*Organization, error) {
	rows, err := pglxqb.SelectAll().From("organizations").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return o.ParseRow(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}

func (o *Organization) GetParsedObjectsByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid []*uuid.UUID, column graphql.CollectedField) ([]*Organization, error) {
	rows, err := pglxqb.SelectAll().From("organizations").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return o.ParseRows(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}
