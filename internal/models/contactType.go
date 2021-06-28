package models

import (
	"context"
	"fmt"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/app"
	"github.com/sphera-erp/sphera/internal/utils"
	"github.com/sphera-erp/sphera/pkg/pglx"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type ContactType struct {
	UUID      *uuid.UUID `json:"uuid" db:"uuid"`
	Created   *time.Time `json:"created" db:"created"`
	Updated   *time.Time `json:"updated" db:"updated"`
	Name      string     `json:"Name" db:"name"`
	IsDeleted *bool      `json:"isDeleted" db:"is_deleted"`
}

type ContactTypeFilter struct {
	UUID      *UUIDFilter     `json:"uuid" db:"uuid"`
	Created   *DateTimeFilter `json:"created" db:"created"`
	Updated   *DateTimeFilter `json:"updated" db:"updated"`
	Name      *StringFilter   `json:"name" db:"name"`
	IsDeleted *bool           `json:"isDeleted" db:"is_deleted"`
}

func (ct *ContactType) Mutation(ctx context.Context, db pglxqb.BaseRunner, app *app.App, rColumns interface{}, columns map[string]interface{}) (*pglx.Rows, *uuid.UUID, error) {
	updateOrDelete := false
	if len(columns) > 0 {
		columns = utils.ClearSQLFields(ct, columns)
	}
	// если есть uuid значит манипулируем обектом
	if ct.UUID != nil {
		if utils.CountFillFields(ct) == 1 && len(columns) == 0 {
			return nil, ct.UUID, nil
		}

		updateOrDelete = true
	}
	parent := make(map[string]interface{})
	setColumns, err := SqlGenKeys(ctx, app, db, ct, columns, parent)
	if err != nil {
		app.Logger.Error().Str("module", "persons").Str("func", "Mutation").Err(err).Msg("Error generate map of keys")
		return nil, nil, err
	} else {
		// иначе создадим с нуля Обьект
		newUUID := uuid.New()
		ct.UUID = &newUUID
		columns["uuid"] = newUUID
	}
	// только одна колонка, и это uuid то удаляем обьект
	setColumns = utils.ClearSQLFields(ct, setColumns)
	if len(setColumns) > 0 {
		if updateOrDelete {
			// Обновляем иначе
			rows, err := pglxqb.Update("contact_type").
				SetMap(setColumns).
				Where("uuid = ?", ct.UUID).
				Suffix(utils.PrepareSuffix(rColumns)).
				RunWith(db).QueryX(ctx)
			if err != nil {
				app.Logger.Error().Str("module", "contact_type").Str("func", "Mutation").Err(err).Msg("Error update contact_type")
				return nil, nil, gqlerror.Errorf("Error update contact_type")
			}
			return rows, ct.UUID, nil
		} else {
			rows, err := pglxqb.Insert("contact_type").
				SetMap(setColumns).
				Suffix("ON CONFLICT (name) DO NOTHING").
				Suffix(utils.PrepareSuffix(rColumns)).
				RunWith(db).QueryX(ctx)
			if err != nil {
				app.Logger.Error().Str("module", "contact_type").Str("func", "Mutation").Err(err).Msg("Error insert contact_type")
				return nil, nil, gqlerror.Errorf("Error insert contact_type")
			}
			if rows == nil {
				rows, err = pglxqb.Select("*").
					From("contact_type").
					Where(pglxqb.Eq(setColumns)).
					RunWith(db).QueryX(ctx)
				if err != nil {
					app.Logger.Error().Str("module", "contact").Str("func", "Mutation").Err(err).Msg("Error select contact_type")
					return nil, nil, gqlerror.Errorf("Error select contact_type")
				}
			}
			return rows, ct.UUID, nil
		}
	}
	return nil, ct.UUID, nil
}

func (ct *ContactType) ParseRow(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) (*ContactType, error) {
	var err error
	var contactType ContactType
	defer rows.Close()
	for rows.Next() {
		err = rows.StructScan(&contactType)
		if err != nil {
			app.Logger.Error().Str("module", "contacts").Str("func", "parseContactTypeRow").Err(err).Msg("Error scan response to struct contactType")
			return nil, gqlerror.Errorf("Error scan response to struct contactType")
		}
	}
	err = contactType.parseRequestedFields(ctx, fields, app, db)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return &contactType, nil
}

func (ct *ContactType) ParseRows(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) ([]*ContactType, error) {
	var err error
	var contactTypes []*ContactType
	defer rows.Close()
	for rows.Next() {
		var contactType ContactType
		err = rows.StructScan(&contactType)
		if err != nil {
			app.Logger.Error().Str("module", "contacts").Str("func", "parseContactTypeRow").Err(err).Msg("Error scan response to struct contactType")
			return nil, gqlerror.Errorf("Error scan response to struct contactType")
		}
		contactTypes = append(contactTypes, &contactType)
	}
	for _, contactType := range contactTypes {
		err = contactType.parseRequestedFields(ctx, fields, app, db)
		if err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
	}
	return contactTypes, nil
}

func (ct *ContactType) parseRequestedFields(ctx context.Context, fields []graphql.CollectedField, app *app.App, db pglxqb.BaseRunner) error {
	return parseRequestedFields(ctx, app, db, fields, ct)
}

func (ct *ContactType) GetByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID) (*ContactType, error) {
	rows, err := pglxqb.SelectAll().From("contact_type").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error get person from DB")
		return nil, gqlerror.Errorf("Error get person from DB")
	}
	var contactType ContactType
	defer rows.Close()
	for rows.Next() {
		if err := rows.StructScan(&contactType); err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct Person")
			return nil, gqlerror.Errorf("Error scan response to struct Person")
		}
	}
	return &contactType, nil
}

func (ct *ContactType) GetParsedObjectByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID, column graphql.CollectedField) (*ContactType, error) {
	rows, err := pglxqb.SelectAll().From("contact_type").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	fmt.Println("GetParsedObjectByUUID ContactType")
	return ct.ParseRow(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}

func (ct *ContactType) GetParsedObjectsByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid []*uuid.UUID, column graphql.CollectedField) ([]*ContactType, error) {
	rows, err := pglxqb.SelectAll().From("contact_type").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return ct.ParseRows(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}
