package models

import (
	"context"
	"fmt"
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

type Contact struct {
	UUID             *uuid.UUID    `json:"uuid" db:"uuid"`
	Created          *time.Time    `json:"created" db:"created"`
	Updated          *time.Time    `json:"updated" db:"updated"`
	UUIDPerson       *uuid.UUID    `db:"uuid_person"`
	Person           *Person       `json:"person" relay:"uuid_person" link:"UUIDPerson"`
	UUIDOrganization *uuid.UUID    `db:"uuid_organization"`
	Organization     *Organization `json:"organization" relay:"uuid_organization" link:"UUIDOrganization"`
	UUIDContactType  *uuid.UUID    `db:"uuid_contact_type"`
	ContactType      *ContactType  `json:"contactType" relay:"uuid_contact_type" link:"UUIDContactType"`
	Presentation     *string       `json:"presentation" db:"presentation"`
	IsDeleted        *bool         `json:"isDeleted" db:"is_deleted"`
}

type ContactFilter struct {
	UUID         *UUIDFilter      `json:"uuid" db:"uuid"`
	Created      *DateTimeFilter  `json:"created" db:"created"`
	Updated      *DateTimeFilter  `json:"updated" db:"updated"`
	Person       *PersonFilter    `json:"person" table:"persons" link:"uuid_person"`
	ContactType  *ContactType     `json:"contactType" table:"contact_type" link:"uuid_contact_type"`
	Presentation *StringFilter    `json:"presentation" db:"presentation"`
	IsDeleted    *bool            `json:"isDeleted" db:"is_deleted"`
	And          []*ContactFilter `json:"and"`
	Or           []*ContactFilter `json:"or"`
	Not          *ContactFilter   `json:"not"`
}

func (c *Contact) Mutation(ctx context.Context, db pglxqb.BaseRunner, app *app.App, rColumns interface{}, columns map[string]interface{}) (*pglx.Rows, *uuid.UUID, error) {
	// если есть uuid значит манипулируем объектом
	if c.UUID != nil {
		contact, err := c.GetByUUID(ctx, app, db, c.UUID)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "Mutation").Err(err).Msg("Error get person")
			return nil, nil, gqlerror.Errorf("Error get person")
		}
		// востановим все ссылки
		utils.RestoreUUID(c, contact)
		// востановим подчиненные структуры
		if err = contact.restoreStruct(ctx, app, db); err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "Mutation").Err(err).Msg("Error restore struct person")
			return nil, nil, gqlerror.Errorf("Error restore struct person")
		}
	} else {
		// иначе создадим с нуля Обьект
		var UUID *uuid.UUID

		if err := pglxqb.Select("uuid").From("contacts").Where(pglxqb.Eq{"presentation": c.Presentation}).RunWith(db).QueryRow(ctx).Scan(&UUID); err != nil {
			if err.Error() != "no rows in result set" {
				app.Logger.Error().Str("package", "models").Str("model", "user").Str("func", "GetParsedObjectByUUID").Err(err).Msg("Error get contact from DB by presentation")
				return nil, nil, gqlerror.Errorf("Error get contact from DB by presentation")
			}
		}

		fmt.Println(UUID)

		if UUID != nil {
			var surname, name, patronymic string
			if err := pglxqb.Select("p.surname, p.name, p.patronymic").
				From("users").
				LeftJoin("persons p on users.uuid = p.uuid_user").
				Where(pglxqb.Eq{"users.uuid_contact": UUID}).
				RunWith(db).QueryRow(ctx).Scan(&surname, &name, &patronymic); err != nil {
				if err.Error() != "no rows in result set" {
					app.Logger.Error().Str("module", "flow").Str("func", "RefuseJob").Err(err).Msg("Error select person info from contact")
					return nil, nil, gqlerror.Errorf("Error select person info from contact")
				}
			}
			return nil, nil, gqlerror.Errorf("This number is registered for %s %s %s", surname, name, patronymic)
		}
		newUUID := uuid.New()
		c.UUID = &newUUID
		_, err := pglxqb.Insert("contacts").
			Columns("uuid").
			Values(newUUID).
			RunWith(db).Exec(ctx)
		if err != nil {
			app.Logger.Error().Err(err).Msg("Error insert user")
			return nil, nil, gqlerror.Errorf("Error insert user")
		}

	}
	// дополним пропущенные поля, если они есть
	parent := make(map[string]interface{})
	setColumns, err := SqlGenKeys(ctx, app, db, c, columns, parent)
	if err != nil {
		app.Logger.Error().Str("module", "persons").Str("func", "Mutation").Err(err).Msg("Error generate map of keys")
		return nil, nil, err
	}
	// только одна колонка, и это uuid то удаляем обьект
	setColumns = utils.ClearSQLFields(c, setColumns)
	if len(setColumns) > 0 {
		// Обновляем иначе
		rows, err := pglxqb.Update("contacts").
			SetMap(setColumns).
			Where("uuid = ?", c.UUID).
			Suffix(utils.PrepareSuffix(rColumns)).
			RunWith(db).QueryX(ctx)
		if err != nil {
			app.Logger.Error().Str("module", "contact").Str("func", "Mutation").Err(err).Msg("Error update contact")
			return nil, nil, gqlerror.Errorf("Error update contact")
		}
		return rows, c.UUID, nil
	}
	return nil, c.UUID, nil
}

func (c *Contact) ParseRows(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) ([]*Contact, error) {
	var contacts []*Contact
	defer rows.Close()
	for rows.Next() {
		var contact Contact
		err := rows.StructScan(&contact)
		if err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
		contacts = append(contacts, &contact)
	}
	fmt.Println(fields)
	for _, contact := range contacts {
		fmt.Println(contact)
		if err := contact.parseRequestedFields(ctx, fields, app, db); err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
	}
	return contacts, nil
}

func (c *Contact) ParseRow(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) (*Contact, error) {
	var err error
	var contact Contact
	defer rows.Close()
	for rows.Next() {
		err = rows.StructScan(&contact)
		if err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
	}
	err = contact.parseRequestedFields(ctx, fields, app, db)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return &contact, nil
}

func (c *Contact) parseRequestedFields(ctx context.Context, fields []graphql.CollectedField, app *app.App, db pglxqb.BaseRunner) error {
	return parseRequestedFields(ctx, app, db, fields, c)
}

func (c *Contact) restoreStruct(ctx context.Context, app *app.App, db pglxqb.BaseRunner) error {
	v := reflect.ValueOf(c)
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

func (c *Contact) GetByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID) (*Contact, error) {
	rows, err := pglxqb.SelectAll().From("contacts").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error get person from DB")
		return nil, gqlerror.Errorf("Error get person from DB")
	}
	var contact Contact
	defer rows.Close()
	for rows.Next() {
		if err := rows.StructScan(&contact); err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct Person")
			return nil, gqlerror.Errorf("Error scan response to struct Person")
		}
	}
	return &contact, nil
}

func (c *Contact) GetParsedObjectByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID, column graphql.CollectedField) (*Contact, error) {
	rows, err := pglxqb.SelectAll().From("contacts").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return c.ParseRow(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}

func (c *Contact) GetParsedObjectsByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid []*uuid.UUID, column graphql.CollectedField) ([]*Contact, error) {
	rows, err := pglxqb.SelectAll().From("contacts").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return c.ParseRows(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}
