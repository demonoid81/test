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

type Content struct {
	UUID      *uuid.UUID      `json:"uuid" db:"uuid"`
	Created   *time.Time      `json:"created" db:"created"`
	Updated   *time.Time      `json:"updated" db:"updated"`
	Bucket    *string         `json:"bucket" db:"bucket"`
	IsDeleted *bool           `json:"isDeleted" db:"is_deleted"`
	File      *graphql.Upload `json:"file"`
}

func (c *Content) Mutation(ctx context.Context, db pglxqb.BaseRunner, app *app.App, rColumns interface{}, columns map[string]interface{}) (*pglx.Rows, *uuid.UUID, error) {
	updateOrDelete := false
	// если есть uuid значит манипулируем обектом
	if c.UUID != nil {
		updateOrDelete = true
	} else {
		// иначе создадим с нуля Обьект
		newUUID := uuid.New()
		c.UUID = &newUUID
		columns["uuid"] = newUUID
	}
	parent := make(map[string]interface{})
	setColumns, err := SqlGenKeys(ctx, app, db, c, columns, parent)
	if err != nil {
		app.Logger.Error().Str("module", "persons").Str("func", "Mutation").Err(err).Msg("Error generate map of keys")
		return nil, nil, err
	}
	// только одна колонка, и это uuid то удаляем обьект
	setColumns = utils.ClearSQLFields(c, setColumns)
	if len(setColumns) == 0 {
		if updateOrDelete {
			// todo Логика Обновления
			// Обновляем иначе
			rows, err := pglxqb.Update("content").
				SetMap(setColumns).
				Where("uuid = ?", c.UUID).
				Suffix(utils.PrepareSuffix(rColumns)).
				RunWith(db).QueryX(ctx)
			if err != nil {
				app.Logger.Error().Str("module", "content").Str("func", "manipulate").Err(err).Msg("Error update content")
				return nil, nil, gqlerror.Errorf("Error update content")
			}
			return rows, c.UUID, nil
		} else {
			// todo Логика вставки
			rows, err := pglxqb.Insert("content").
				SetMap(setColumns).
				Suffix(utils.PrepareSuffix(rColumns)).
				RunWith(db).QueryX(ctx)
			if err != nil {
				app.Logger.Error().Str("module", "passports").Str("func", "manipulate").Err(err).Msg("Error insert content")
				return nil, nil, gqlerror.Errorf("Error insert content")
			}
			return rows, c.UUID, nil
		}
	}
	return nil, c.UUID, nil
}

func (c *Content) ParseRows(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) ([]*Content, error) {
	var contents []*Content
	defer rows.Close()
	for rows.Next() {
		var content Content
		err := rows.StructScan(&content)
		if err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct person")
			return nil, gqlerror.Errorf("Error scan response to struct person")
		}
		contents = append(contents, &content)
	}
	return contents, nil
}

func (c *Content) ParseRow(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) (*Content, error) {
	var err error
	var content Content
	defer rows.Close()
	for rows.Next() {
		err = rows.StructScan(&content)
		if err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct person")
			return nil, gqlerror.Errorf("Error scan response to struct person")
		}
	}
	return &content, nil
}

func (c *Content) parseRequestedFields(ctx context.Context, fields []graphql.CollectedField, app *app.App, db pglxqb.BaseRunner) error {
	return nil
}

func (c *Content) GetByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID) (*Content, error) {
	rows, err := pglxqb.SelectAll().From("content").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error get person from DB")
		return nil, gqlerror.Errorf("Error get person from DB")
	}
	var content Content
	defer rows.Close()
	for rows.Next() {
		if err := rows.StructScan(&content); err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct Person")
			return nil, gqlerror.Errorf("Error scan response to struct Person")
		}
	}
	return &content, nil
}

func (c *Content) GetParsedObjectByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID, column graphql.CollectedField) (*Content, error) {
	rows, err := pglxqb.SelectAll().From("content").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return c.ParseRow(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}

func (c *Content) GetParsedObjectsByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid []*uuid.UUID, column graphql.CollectedField) ([]*Content, error) {
	rows, err := pglxqb.SelectAll().From("content").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return c.ParseRows(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}
