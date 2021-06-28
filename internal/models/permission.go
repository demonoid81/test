package models

import (
	"context"
	"reflect"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/app"
	"github.com/sphera-erp/sphera/internal/utils"
	"github.com/sphera-erp/sphera/pkg/pglx"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type Permission struct {
	UUID      *uuid.UUID `json:"uuid" db:"uuid"`
	Created   *time.Time `json:"created" db:"created"`
	Updated   *time.Time `json:"updated" db:"updated"`
	IsDeleted *bool      `json:"isDeleted" db:"is_deleted"`
	Object    *string    `json:"object" db:"object"`
	Insert    *bool      `json:"insert" db:"i"`
	Read      *bool      `json:"read" db:"r"`
	Update    *bool      `json:"update"  db:"u"`
	Delete    *bool      `json:"delete"  db:"d"`
}

func (p *Permission) Mutation(ctx context.Context, db pglxqb.BaseRunner, app *app.App, rColumns interface{}, columns map[string]interface{}) (*pglx.Rows, *uuid.UUID, error) {
	// Уберем лишние колонки
	if len(columns) > 0 {
		columns = utils.ClearSQLFields(p, columns)
	}
	// если есть uuid значит манипулируем обектом
	if p.UUID != nil {
		if utils.CountFillFields(p) == 1 && len(columns) == 0 {
			return nil, p.UUID, nil
		}
		permission, err := p.GetByUUID(ctx, app, db, p.UUID)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "Mutation").Err(err).Msg("Error get person")
			return nil, nil, gqlerror.Errorf("Error get person")
		}
		// Если не меняются родители то вернем uuid
		if Compare(permission, columns) && utils.CountFillFields(p) == 1 {
			return nil, p.UUID, nil
		}
		// востановим все ссылки
		utils.RestoreUUID(p, permission)
		// востановим подчиненные структуры
		if err = p.restoreStruct(ctx, app, db); err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "Mutation").Err(err).Msg("Error restore struct person")
			return nil, nil, gqlerror.Errorf("Error restore struct person")
		}
	} else {
		// иначе создадим с нуля Обьект
		newUUID := uuid.New()
		p.UUID = &newUUID
		_, err := pglxqb.Insert("permissions").
			Columns("uuid").
			Values(newUUID).
			RunWith(db).Exec(ctx)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "Mutation").Err(err).Msg("Error insert tag")
			return nil, nil, gqlerror.Errorf("Error insert user")
		}
	}
	// дополним пропущенные поля, если они есть
	parent := make(map[string]interface{})
	setColumns, err := SqlGenKeys(ctx, app, db, p, columns, parent)
	if err != nil {
		app.Logger.Error().Str("module", "persons").Str("func", "Mutation").Err(err).Msg("Error generate map of keys")
		return nil, nil, err
	}
	// почистим колонки от мусора
	setColumns = utils.ClearSQLFields(p, setColumns)
	if len(setColumns) > 0 {
		// Обновляем иначе
		rows, err := pglxqb.Update("permissions").
			SetMap(setColumns).
			Where("uuid = ?", p.UUID).
			Suffix(utils.PrepareSuffix(rColumns)).
			RunWith(db).QueryX(ctx)
		if err != nil {
			app.Logger.Error().Str("module", "contact").Str("func", "Mutation").Err(err).Msg("Error update contact")
			return nil, nil, gqlerror.Errorf("Error update contact")
		}
		return rows, p.UUID, nil
	}
	return nil, p.UUID, nil
}

func (p *Permission) ParseRows(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) ([]*Permission, error) {
	var permissions []*Permission
	for rows.Next() {
		var permission Permission
		if err := rows.StructScan(&permission); err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct Tag")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
		permissions = append(permissions, &permission)
	}
	for _, role := range permissions {
		if err := role.parseRequestedFields(ctx, fields, app, db); err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct Tag")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
	}
	return permissions, nil
}

func (p *Permission) ParseRow(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) (*Permission, error) {
	var permission Permission
	defer rows.Close()
	for rows.Next() {
		if err := rows.StructScan(&permission); err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
	}
	if err := permission.parseRequestedFields(ctx, fields, app, db); err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return &permission, nil
}

func (p *Permission) parseRequestedFields(ctx context.Context, fields []graphql.CollectedField, app *app.App, db pglxqb.BaseRunner) error {
	return parseRequestedFields(ctx, app, db, fields, p)
}

func (p *Permission) restoreStruct(ctx context.Context, app *app.App, db pglxqb.BaseRunner) error {
	v := reflect.ValueOf(p)
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

func (p *Permission) GetByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID) (*Permission, error) {
	rows, err := pglxqb.SelectAll().From("permissions").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error get person from DB")
		return nil, gqlerror.Errorf("Error get person from DB")
	}
	var permission Permission
	for rows.Next() {
		if err := rows.StructScan(&permission); err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct Person")
			return nil, gqlerror.Errorf("Error scan response to struct Person")
		}
	}
	return &permission, nil
}

func (p *Permission) GetParsedObjectByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID, column graphql.CollectedField) (*Permission, error) {
	rows, err := pglxqb.SelectAll().From("permissions").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return p.ParseRow(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}

func (p *Permission) GetParsedObjectsByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid []*uuid.UUID, column graphql.CollectedField) ([]*Permission, error) {
	rows, err := pglxqb.SelectAll().From("permissions").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return p.ParseRows(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}
