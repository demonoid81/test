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

type Nationality struct {
	UUID    *uuid.UUID `json:"uuid" db:"uuid"`
	Created *time.Time `json:"created" db:"created"`
	Updated *time.Time `json:"updated" db:"updated"`
	Name    *string    `json:"name" db:"name"`
	IsDeleted        *bool         `json:"isDeleted" db:"is_deleted"`
}

func (n *Nationality) Mutation(ctx context.Context, db pglxqb.BaseRunner, app *app.App, rColumns interface{}, columns map[string]interface{}) (*pglx.Rows, *uuid.UUID, error) {
	updateOrDelete := false
	// если есть uuid значит манипулируем обектом
	if n.UUID != nil {
		updateOrDelete = true
	} else {
		// иначе создадим с нуля Обьект
		newUUID := uuid.New()
		n.UUID = &newUUID
		columns["uuid"] = newUUID
	}
	parent := make(map[string]interface{})
	setColumns, err := SqlGenKeys(ctx, app, db, n, columns, parent)
	if err != nil {
		app.Logger.Error().Str("module", "persons").Str("func", "Mutation").Err(err).Msg("Error generate map of keys")
		return nil, nil, err
	}
	// только одна колонка, и это uuid то удаляем обьект
	setColumns = utils.ClearSQLFields(n, setColumns)
	if len(setColumns) == 0 {
		if updateOrDelete {
			// todo Логика Обновления
			// Обновляем иначе
			rows, err := pglxqb.Update("nationalities").
				SetMap(setColumns).
				Where("uuid = ?", n.UUID).
				Suffix(utils.PrepareSuffix(rColumns)).
				RunWith(db).QueryX(ctx)
			if err != nil {
				app.Logger.Error().Str("module", "persons").Str("func", "Mutation").Err(err).Msg("Error update nationality")
				return nil, nil, gqlerror.Errorf("Error update person")
			}
			return rows, n.UUID, nil
		} else {
			// todo Логика вставки
			rows, err := pglxqb.Insert("nationalities").
				SetMap(setColumns).
				Suffix("ON CONFLICT (name) DO NOTHING RETURNING *").
				Suffix(utils.PrepareSuffix(rColumns)).
				RunWith(db).QueryX(ctx)
			if err != nil {
				app.Logger.Error().Str("module", "persons").Str("func", "Mutation").Err(err).Msg("Error insert nationality")
				return nil, nil, gqlerror.Errorf("Error insert person")
			}
			if rows == nil {
				rows, err = pglxqb.SelectAll().
					From("nationalities").
					Where(pglxqb.Eq(columns)).
					RunWith(db).QueryX(ctx)
				if err != nil {
					app.Logger.Error().Str("module", "contact").Str("func", "Mutation").Err(err).Msg("Error select nationality")
					return nil, nil, gqlerror.Errorf("Error select contact_type")
				}
			}
			return rows, n.UUID, nil
		}
	}
	return nil, n.UUID, nil
}

func (n *Nationality) ParseRows(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) ([]*Nationality, error) {
	var nationalities []*Nationality
	defer rows.Close()
	for rows.Next() {
		var nationality Nationality
		err := rows.StructScan(&nationality)
		if err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct person")
			return nil, gqlerror.Errorf("Error scan response to struct person")
		}
		err = nationality.parseRequestedFields(ctx, fields, app, db)
		if err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct person")
			return nil, gqlerror.Errorf("Error scan response to struct person")
		}
		nationalities = append(nationalities, &nationality)
	}
	return nationalities, nil
}

func (n *Nationality) ParseRow(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) (*Nationality, error) {
	var err error
	var nationality Nationality
	defer rows.Close()
	for rows.Next() {
		err = rows.StructScan(&nationality)
		if err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct person")
			return nil, gqlerror.Errorf("Error scan response to struct person")
		}
	}
	err = nationality.parseRequestedFields(ctx, fields, app, db)
	if err != nil {
		app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct person")
		return nil, gqlerror.Errorf("Error scan response to struct person")
	}
	return &nationality, nil
}

func (n *Nationality) parseRequestedFields(ctx context.Context, fields []graphql.CollectedField, app *app.App, db pglxqb.BaseRunner) error {
	return parseRequestedFields(ctx, app, db, fields, n)
}

func (n *Nationality) GetByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID) (*Nationality, error) {
	rows, err := pglxqb.SelectAll().From("nationalities").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error get person from DB")
		return nil, gqlerror.Errorf("Error get person from DB")
	}
	var nationality Nationality
	defer rows.Close()
	for rows.Next() {
		if err := rows.StructScan(&nationality); err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct Person")
			return nil, gqlerror.Errorf("Error scan response to struct Person")
		}
	}
	return &nationality, nil
}

func (n *Nationality) GetParsedObjectByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID, column graphql.CollectedField) (*Nationality, error) {
	rows, err := pglxqb.SelectAll().From("nationalities").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return n.ParseRow(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}

func (n *Nationality) GetParsedObjectsByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid []*uuid.UUID, column graphql.CollectedField) ([]*Nationality, error) {
	rows, err := pglxqb.SelectAll().From("nationalities").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return n.ParseRows(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}