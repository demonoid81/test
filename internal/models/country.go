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
	"time"
)

type Country struct {
	UUID    *uuid.UUID `json:"uuid" db:"uuid"`
	Name    *string    `json:"name" db:"country"`
	Created *time.Time `json:"created" db:"created"`
	Updated *time.Time `json:"updated" db:"updated"`
	IsDeleted        *bool         `json:"isDeleted" db:"is_deleted"`
}

type CountryFilter struct {
	UUID      *UUIDFilter     `json:"uuid" db:"uuid"`
	Name      *StringFilter   `json:"name" db:"name"`
	Created   *DateTimeFilter `json:"created" db:"created"`
	Updated   *DateTimeFilter `json:"updated" db:"updated"`
	IsDeleted *bool           `json:"isDeleted" db:"is_deleted"`
}

func (op *Country) Mutation(ctx context.Context, db pglxqb.BaseRunner, app *app.App, rColumns interface{}, columns map[string]interface{}) (*pglx.Rows, *uuid.UUID, error) {
	updateOrDelete := false
	// если есть uuid значит манипулируем обектом
	if op.UUID != nil {
		updateOrDelete = true
	} else {
		// иначе создадим с нуля Обьект
		newUUID := uuid.New()
		op.UUID = &newUUID
		columns["uuid"] = newUUID
	}
	parent := make(map[string]interface{})
	setColumns, err := SqlGenKeys(ctx, app, db, op, columns, parent)
	if err != nil {
		app.Logger.Error().Str("module", "persons").Str("func", "Mutation").Err(err).Msg("Error generate map of keys")
		return nil, nil, err
	}
	// только одна колонка, и это uuid то удаляем обьект
	setColumns = utils.ClearSQLFields(op, setColumns)
	if len(setColumns) > 0 {
		if updateOrDelete {
			// Обновляем иначе
			rows, err := pglxqb.Update("countries").
				SetMap(setColumns).
				Where("uuid = ?", op.UUID).
				Suffix(utils.PrepareSuffix(rColumns)).
				RunWith(db).QueryX(ctx)
			if err != nil {
				app.Logger.Error().Str("module", "models").Str("func", "Mutation").Err(err).Msg("Error update country")
				return nil, nil, gqlerror.Errorf("Error update country")
			}
			return rows, op.UUID, nil
		} else {
			rows, err := pglxqb.Insert("countries").
				SetMap(setColumns).
				Suffix("ON CONFLICT (country) DO NOTHING").
				Suffix(utils.PrepareSuffix(rColumns)).
				RunWith(db).QueryX(ctx)
			if err != nil {
				app.Logger.Error().Str("module", "models").Str("func", "Mutation").Err(err).Msg("Error insert country")
				return nil, nil, gqlerror.Errorf("Error insert country")
			}
			return rows, op.UUID, nil
		}
	}
	return nil, op.UUID, nil
}

func (op *Country) ParseRows(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) ([]*Country, error) {
	var countries []*Country
	defer rows.Close()
	for rows.Next() {
		var country Country
		err := rows.StructScan(&country)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct country")
			return nil, gqlerror.Errorf("Error scan response to struct country")
		}
		countries = append(countries, &country)
	}
	return countries, nil
}

func (op *Country) ParseRow(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) (*Country, error) {
	var err error
	var country Country
	defer rows.Close()
	for rows.Next() {
		err = rows.StructScan(&country)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "manipulate").Err(err).Msg("Error scan response to struct country")
			return nil, gqlerror.Errorf("Error scan response to struct country")
		}
	}
	return &country, nil
}

func (op *Country) GetByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID) (*Country, error) {
	rows, err := pglxqb.SelectAll().From("persons").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error get person from DB")
		return nil, gqlerror.Errorf("Error get person from DB")
	}
	var country Country
	defer rows.Close()
	for rows.Next() {
		if err := rows.StructScan(&country); err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct Person")
			return nil, gqlerror.Errorf("Error scan response to struct Person")
		}
	}
	return &country, nil
}

func (op *Country) GetParsedObjectByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID, column graphql.CollectedField) (*Country, error) {
	rows, err := pglxqb.SelectAll().From("persons").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return op.ParseRow(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}

func (op *Country) GetParsedObjectsByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid []*uuid.UUID, column graphql.CollectedField) ([]*Country, error) {
	rows, err := pglxqb.SelectAll().From("persons").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return op.ParseRows(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}