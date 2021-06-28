package models

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/app"
	"github.com/sphera-erp/sphera/internal/utils"
	"github.com/sphera-erp/sphera/pkg/pglx"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type Role struct {
	UUID             *uuid.UUID    `json:"uuid" db:"uuid"`
	Name             *string       `json:"name" db:"name"`
	Created          *time.Time    `json:"created" db:"created"`
	Updated          *time.Time    `json:"updated" db:"updated"`
	IsDeleted        *bool         `json:"isDeleted" db:"is_deleted"`
	UUIDOrganization *uuid.UUID    `db:"uuid_organization"`
	Organization     *Organization `json:"organization" relay:"uuid_organization" link:"UUIDOrganization"`
	RoleType         *TypeRole     `json:"roleType" db:"role_type"`
	UUIDPermissions  []*uuid.UUID  `db:"uuid_permissions"`
	Permissions      []*Permission `json:"permissions" relay:"uuid_permissions" link:"UUIDPermissions"`
}

type TypeRole string

const (
	TypeRoleSystem              TypeRole = "system"
	TypeRoleOrganizationManager TypeRole = "organizationManager"
	TypeRolGroupManager         TypeRole = "groupManager"
	TypeRoleObjectManager       TypeRole = "objectManager"
)

var AllTypeRole = []TypeRole{
	TypeRoleSystem,
	TypeRoleOrganizationManager,
	TypeRolGroupManager,
	TypeRoleObjectManager,
}

func (e TypeRole) IsValid() bool {
	switch e {
	case TypeRoleSystem, TypeRoleOrganizationManager, TypeRolGroupManager, TypeRoleObjectManager:
		return true
	}
	return false
}

func (e TypeRole) String() string {
	return string(e)
}

func (e *TypeRole) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = TypeRole(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid TypeRole", str)
	}
	return nil
}

func (e TypeRole) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

func (t *Role) Mutation(ctx context.Context, db pglxqb.BaseRunner, app *app.App, rColumns interface{}, columns map[string]interface{}) (*pglx.Rows, *uuid.UUID, error) {
	// Уберем лишние колонки
	if len(columns) > 0 {
		columns = utils.ClearSQLFields(t, columns)
	}
	update := false
	// если есть uuid значит манипулируем обектом
	if t.UUID != nil {
		if utils.CountFillFields(t) == 1 && len(columns) == 0 {
			return nil, t.UUID, nil
		}
		tag, err := t.GetByUUID(ctx, app, db, t.UUID)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "Mutation").Err(err).Msg("Error get person")
			return nil, nil, gqlerror.Errorf("Error get person")
		}
		// Если не меняются родители то вернем uuid
		if Compare(tag, columns) && utils.CountFillFields(t) == 1 {
			return nil, t.UUID, nil
		}
		// востановим все ссылки
		utils.RestoreUUID(t, tag)
		// востановим подчиненные структуры
		if err = t.restoreStruct(ctx, app, db); err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "Mutation").Err(err).Msg("Error restore struct person")
			return nil, nil, gqlerror.Errorf("Error restore struct person")
		}
		update = true
	} else {
		// иначе создадим с нуля Обьект
		newUUID := uuid.New()
		t.UUID = &newUUID
		_, err := pglxqb.Insert("roles").
			Columns("uuid").
			Values(newUUID).
			RunWith(db).Exec(ctx)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "Mutation").Err(err).Msg("Error insert tag")
			return nil, nil, gqlerror.Errorf("Error insert user")
		}
	}
	// дополним пропущеные поля, если они есть
	parent := make(map[string]interface{})
	setColumns, err := SqlGenKeys(ctx, app, db, t, columns, parent)
	if err != nil {
		app.Logger.Error().Str("module", "persons").Str("func", "Mutation").Err(err).Msg("Error generate map of keys")
		return nil, nil, err
	}
	// почистим колонки от мусора
	setColumns = utils.ClearSQLFields(t, setColumns)
	if len(setColumns) > 0 {
		if update {
			// Обновляем иначе
			rows, err := pglxqb.Update("roles").
				SetMap(setColumns).
				Where("uuid = ?", t.UUID).
				Suffix(utils.PrepareSuffix(rColumns)).
				RunWith(db).QueryX(ctx)
			if err != nil {
				app.Logger.Error().Str("module", "contact").Str("func", "Mutation").Err(err).Msg("Error update contact")
				return nil, nil, gqlerror.Errorf("Error update contact")
			}
			return rows, t.UUID, nil
		} else {
			rows, err := pglxqb.Insert("roles").
				SetMap(setColumns).
				Suffix(utils.PrepareSuffix(rColumns)).
				RunWith(db).QueryX(ctx)
			if err != nil {
				app.Logger.Error().Str("module", "contact").Str("func", "Mutation").Err(err).Msg("Error insert contact")
				return nil, nil, gqlerror.Errorf("Error insert contact")
			}
			return rows, t.UUID, nil
		}
	}
	return nil, t.UUID, nil
}

func (t *Role) ParseRows(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) ([]*Role, error) {
	var roles []*Role
	for rows.Next() {
		var role Role
		if err := rows.StructScan(&role); err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct Tag")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
		roles = append(roles, &role)
	}
	for _, role := range roles {
		if err := role.parseRequestedFields(ctx, fields, app, db); err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct Tag")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
	}
	return roles, nil
}

func (t *Role) ParseRow(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) (*Role, error) {
	var role Role
	defer rows.Close()
	for rows.Next() {
		if err := rows.StructScan(&role); err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
	}
	if err := role.parseRequestedFields(ctx, fields, app, db); err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return &role, nil
}

func (t *Role) parseRequestedFields(ctx context.Context, fields []graphql.CollectedField, app *app.App, db pglxqb.BaseRunner) error {
	return parseRequestedFields(ctx, app, db, fields, t)
}

func (t *Role) restoreStruct(ctx context.Context, app *app.App, db pglxqb.BaseRunner) error {
	v := reflect.ValueOf(t)
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

func (t *Role) GetByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID) (*Role, error) {
	rows, err := pglxqb.SelectAll().From("roles").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error get person from DB")
		return nil, gqlerror.Errorf("Error get person from DB")
	}
	var tag Role
	for rows.Next() {
		if err := rows.StructScan(&tag); err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct Person")
			return nil, gqlerror.Errorf("Error scan response to struct Person")
		}
	}
	return &tag, nil
}

func (t *Role) GetParsedObjectByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID, column graphql.CollectedField) (*Role, error) {
	rows, err := pglxqb.SelectAll().From("roles").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return t.ParseRow(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}

func (t *Role) GetParsedObjectsByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid []*uuid.UUID, column graphql.CollectedField) ([]*Role, error) {
	rows, err := pglxqb.SelectAll().From("roles").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return t.ParseRows(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}
