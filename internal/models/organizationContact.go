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

type OrganizationContact struct {
	UUID         *uuid.UUID            `json:"uuid" db:"uuid"`
	Created      *time.Time            `json:"created" db:"created"`
	Updated      *time.Time            `json:"updated" db:"updated"`
	IsDeleted    *bool                 `json:"isDeleted" db:"is_deleted"`
	UUIDPerson   *uuid.UUID            `db:"uuid_person"`
	Person       *Person               `json:"person" relay:"uuid_person" link:"UUIDPerson"`
	UUIDPosition *uuid.UUID            `db:"uuid_position"`
	Position     *OrganizationPosition `json:"position" relay:"uuid_position" link:"UUIDPosition"`
}

func (oc *OrganizationContact) Mutation(ctx context.Context, db pglxqb.BaseRunner, app *app.App, rColumns interface{}, columns map[string]interface{}) (*pglx.Rows, *uuid.UUID, error) {
	updateOrDelete := false
	// если есть uuid значит манипулируем обектом
	if oc.UUID != nil {
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
	if len(setColumns) > 0 {
		if updateOrDelete {
			// Обновляем иначе
			rows, err := pglxqb.Update("organization_contacts").
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
			rows, err := pglxqb.Insert("organization_contacts").
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

func (oc *OrganizationContact) ParseRows(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) ([]*OrganizationContact, error) {
	var organizationContacts []*OrganizationContact
	defer rows.Close()
	for rows.Next() {
		var organizationContact OrganizationContact
		err := rows.StructScan(&organizationContact)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct organizationContact")
			return nil, gqlerror.Errorf("Error scan response to struct organizationContact")
		}
		organizationContacts = append(organizationContacts, &organizationContact)
	}
	return organizationContacts, nil
}

func (oc *OrganizationContact) ParseRow(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) (*OrganizationContact, error) {
	var err error
	var organizationContact OrganizationContact
	defer rows.Close()
	for rows.Next() {
		err = rows.StructScan(&organizationContact)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "manipulate").Err(err).Msg("Error scan response to struct organizationContact")
			return nil, gqlerror.Errorf("Error scan response to struct organizationContact")
		}
	}
	return &organizationContact, nil
}

func (oc *OrganizationContact) GetByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID) (*OrganizationContact, error) {
	rows, err := pglxqb.SelectAll().From("organization_contacts").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error get person from DB")
		return nil, gqlerror.Errorf("Error get person from DB")
	}
	var organizationContact OrganizationContact
	defer rows.Close()
	for rows.Next() {
		if err := rows.StructScan(&organizationContact); err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct Person")
			return nil, gqlerror.Errorf("Error scan response to struct Person")
		}
	}
	return &organizationContact, nil
}

func (oc *OrganizationContact) GetParsedObjectByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID, column graphql.CollectedField) (*OrganizationContact, error) {
	rows, err := pglxqb.SelectAll().From("organization_contacts").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return oc.ParseRow(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}

func (oc *OrganizationContact) GetParsedObjectsByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid []*uuid.UUID, column graphql.CollectedField) ([]*OrganizationContact, error) {
	rows, err := pglxqb.SelectAll().From("organization_contacts").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return oc.ParseRows(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}