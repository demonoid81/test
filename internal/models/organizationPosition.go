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

type OrganizationPosition struct {
	UUID    *uuid.UUID `json:"uuid" db:"uuid"`
	Name    *string    `json:"name" db:"name"`
	Created *time.Time `json:"created" db:"created"`
	Updated *time.Time `json:"updated" db:"updated"`
	IsDeleted        *bool         `json:"isDeleted" db:"is_deleted"`
}

func (oc *OrganizationPosition) Mutation(ctx context.Context, db pglxqb.BaseRunner, app *app.App, rColumns interface{}, columns map[string]interface{}) (*pglx.Rows, *uuid.UUID, error) {
	updateOrDelete := false
		fmt.Println(columns)
	// если есть uuid значит манипулируем обектом
	if oc.UUID != nil {
		if utils.CountFillFields(oc) == 1 && len(columns) == 0 {
			return nil, oc.UUID, nil
		}
		// востановим объект
		person, err := oc.GetByUUID(ctx, app, db, oc.UUID)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "Mutation").Err(err).Msg("Error get person")
			return nil, nil, gqlerror.Errorf("Error get person")
		}
		// востановим все ссылки
		utils.RestoreUUID(oc, person)
		// востановим подчиненные структуры
		if err = oc.restoreStruct(ctx, app, db); err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "Mutation").Err(err).Msg("Error restore struct person")
			return nil, nil, gqlerror.Errorf("Error restore struct person")
		}
		updateOrDelete = true
	} else {
		// иначе создадим с нуля Обьект
		newUUID := uuid.New()
		oc.UUID = &newUUID
		columns["uuid"] = newUUID
	}
	parent := make(map[string]interface{})
	setColumns, err := SqlGenKeys(ctx, app, db, oc, columns, parent)
	if err != nil {
		app.Logger.Error().Str("module", "persons").Str("func", "Mutation").Err(err).Msg("Error generate map of keys")
		return nil, nil, err
	}
	// только одна колонка, и это uuid то удаляем обьект
	setColumns = utils.ClearSQLFields(oc, setColumns)
	fmt.Println(setColumns)
	if len(setColumns) > 0 {
		if updateOrDelete {
			// Обновляем иначе
			rows, err := pglxqb.Update("organization_positions").
				SetMap(setColumns).
				Where("uuid = ?", oc.UUID).
				Suffix(utils.PrepareSuffix(rColumns)).
				RunWith(db).QueryX(ctx)
			if err != nil {
				app.Logger.Error().Str("module", "models").Str("func", "Mutation").Err(err).Msg("Error update country")
				return nil, nil, gqlerror.Errorf("Error update country")
			}
			return rows, oc.UUID, nil
		} else {
			rows, err := pglxqb.Insert("organization_positions").
				SetMap(setColumns).
				Suffix(utils.PrepareSuffix(rColumns)).
				RunWith(db).QueryX(ctx)
			if err != nil {
				app.Logger.Error().Str("module", "models").Str("func", "Mutation").Err(err).Msg("Error insert country")
				return nil, nil, gqlerror.Errorf("Error insert country")
			}
			return rows, oc.UUID, nil
		}
	}
	return nil, oc.UUID, nil
}

func (oc *OrganizationPosition) ParseRows(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) ([]*OrganizationPosition, error) {
	var organizationPositions []*OrganizationPosition
	defer rows.Close()
	for rows.Next() {
		var organizationPosition OrganizationPosition
		err := rows.StructScan(&organizationPosition)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct organizationPosition")
			return nil, gqlerror.Errorf("Error scan response to struct organizationPosition")
		}
		organizationPositions = append(organizationPositions, &organizationPosition)
	}
	return organizationPositions, nil
}

func (oc *OrganizationPosition) ParseRow(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) (*OrganizationPosition, error) {
	var err error
	var organizationPosition OrganizationPosition
	defer rows.Close()
	for rows.Next() {
		err = rows.StructScan(&organizationPosition)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "manipulate").Err(err).Msg("Error scan response to struct organizationPosition")
			return nil, gqlerror.Errorf("Error scan response to struct organizationPosition")
		}
	}
	return &organizationPosition, nil
}

func (oc *OrganizationPosition) GetByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID) (*OrganizationPosition, error) {
	rows, err := pglxqb.SelectAll().From("organization_positions").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error get person from DB")
		return nil, gqlerror.Errorf("Error get person from DB")
	}
	var organizationPosition OrganizationPosition
	defer rows.Close()
	for rows.Next() {
		if err := rows.StructScan(&organizationPosition); err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct Person")
			return nil, gqlerror.Errorf("Error scan response to struct Person")
		}
	}
	return &organizationPosition, nil
}

func (oc *OrganizationPosition) restoreStruct (ctx context.Context, app *app.App, db pglxqb.BaseRunner) error {
	v := reflect.ValueOf(oc)
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

func (oc *OrganizationPosition) GetParsedObjectByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID, column graphql.CollectedField) (*OrganizationPosition, error) {
	rows, err := pglxqb.SelectAll().From("organization_positions").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return oc.ParseRow(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}

func (oc *OrganizationPosition) GetParsedObjectsByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid []*uuid.UUID, column graphql.CollectedField) ([]*OrganizationPosition, error) {
	rows, err := pglxqb.SelectAll().From("organization_positions").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return oc.ParseRows(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}