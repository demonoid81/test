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
	"io"
	"reflect"
	"strconv"
	"time"
)

type User struct {
	UUID              *uuid.UUID      `json:"uuid" db:"uuid"`
	Created           *time.Time      `json:"created" db:"created"`
	Updated           *time.Time      `json:"updated" db:"updated"`
	IsDeleted         *bool           `json:"isDeleted" db:"is_deleted"`
	IsBlocked         *bool           `json:"isBlocked" db:"is_blocked"`
	IsDisabled        *bool           `json:"isDisabled" db:"is_disabled"`
	UUIDContact       *uuid.UUID      `db:"uuid_contact"`
	Contact           *Contact        `json:"contact" relay:"uuid_contact" link:"UUIDContact"`
	UUIDPerson        *uuid.UUID      `db:"uuid_person"`
	Person            *Person         `json:"person" relay:"uuid_person" link:"UUIDPerson"`
	UserType          *UserType       `json:"type" db:"type"`
	UUIDOrganization  *uuid.UUID      `db:"uuid_organization"`
	Organization      *Organization   `json:"organization" relay:"uuid_organization" link:"UUIDOrganization"`
	UUIDGroups        []*uuid.UUID    `db:"uuid_groups"`
	Groups            []*Organization `json:"groups" relay:"uuid_groups" link:"UUIDGroups"`
	UUIDObjects       []*uuid.UUID    `db:"uuid_objects"`
	UUIDObject        *uuid.UUID      `db:"uuid_object"`
	Objects           []*Organization `json:"objects" relay:"uuid_objects" link:"UUIDObjects"`
	NotificationToken *string         `json:"notification_token" db:"notification_token"`
	UUIDRole          *uuid.UUID      `db:"uuid_role"`
	Role              *Role           `json:"role" relay:"uuid_role" link:"UUIDRole"`
}

type UserSort struct {
	Field *UserSortableField `json:"field"`
	Order *SortOrder         `json:"order"`
}

type UserFilter struct {
	//UUID	   *UUIDFilter
	Created    *DateTimeFilter `json:"created" db:"created"`
	Updated    *DateTimeFilter `json:"updated" db:"updated"`
	IsDeleted  *bool           `json:"isDeleted" db:"is_deleted"`
	IsBlocked  *bool           `json:"isBlocked" db:"is_blocked"`
	IsDisabled *bool           `json:"isDisabled" db:"is_disabled"`
	Contact    *ContactFilter  `json:"contact" table:"contacts" link:"uuid_contact"`
	Person     *PersonFilter   `json:"person" table:"persons" link:"uuid_person"` // Для расширения используем ссылки на таблицы

	And []UserFilter `json:"and"`
	Or  []UserFilter `json:"or"`
	Not *UserFilter  `json:"not"`
}

type UserType string

const (
	SelfEmployed UserType = "SelfEmployed"
	SystemUser   UserType = "SystemUser"
)

func (e UserType) IsValid() bool {
	switch e {
	case SelfEmployed, SystemUser:
		return true
	}
	return false
}

func (e UserType) String() string {
	return string(e)
}

func (e *UserType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}
	*e = UserType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid gender", str)
	}
	return nil
}

func (e UserType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type UserSortableField string

const (
	Created UserSortableField = "created"
	Updated UserSortableField = "updated"
)

func (e UserSortableField) IsValid() bool {
	switch e {
	case Created, Updated:
		return true
	}
	return false
}

func (e UserSortableField) String() string {
	return string(e)
}

func (e *UserSortableField) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}
	*e = UserSortableField(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid gender", str)
	}
	return nil
}

func (e UserSortableField) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

func (u *User) Mutation(ctx context.Context, db pglxqb.BaseRunner, app *app.App, rColumns interface{}, columns map[string]interface{}) (*pglx.Rows, *uuid.UUID, error) {
	logger := app.Logger.Error().Str("package", "models").Str("model", "user").Str("func", "Mutation")
	// Уберем лишние колонки
	if len(columns) > 0 {
		columns = utils.ClearSQLFields(u, columns)
	}
	update := false
	// если есть uuid значит манипулируем обектом
	if u.UUID != nil {
		if utils.CountFillFields(u) == 1 && len(columns) == 0 {
			return nil, u.UUID, nil
		}
		// получим Обьект
		user, err := u.GetByUUID(ctx, app, db, u.UUID)
		if err != nil {
			logger.Err(err).Msg("Error get user")
			return nil, nil, gqlerror.Errorf("Error get user")
		}
		// Если не меняются родители то вернем uuid
		if Compare(user, columns) && utils.CountFillFields(u) == 1 {
			return nil, u.UUID, nil
		}
		// сравним два Объекта
		utils.RestoreUUID(u, user)
		if err = u.restoreStruct(ctx, app, db); err != nil {
			logger.Err(err).Msg("Error restore struct user")
			return nil, nil, gqlerror.Errorf("Error restore struct user")
		}
		update = true
	} else {
		newUUID := uuid.New()
		u.UUID = &newUUID
		_, err := pglxqb.Insert("users").
			Columns("uuid").
			Values(newUUID).
			RunWith(db).Exec(ctx)
		if err != nil {
			logger.Err(err).Msg("Error insert user")
			return nil, nil, gqlerror.Errorf("Error insert user")
		}
		update = true
	}
	// Дополним поля связями c пользователями
	parentColumns := map[string]interface{}{"uuid_user": u.UUID}
	// сгенерим структуру для вставки
	setColumns, err := SqlGenKeys(ctx, app, db, u, columns, parentColumns)
	if err != nil {
		logger.Err(err).Msg("Error generate map of keys")
		return nil, nil, err
	}
	// Уберем лишние колонки
	setColumns = utils.ClearSQLFields(u, setColumns)
	// выполним мутацию
	if len(setColumns) > 0 {
		if update {
			// todo Логика Обновления
			// Обновляем иначе
			rows, err := pglxqb.Update("users").
				SetMap(setColumns).
				Where("uuid = ?", u.UUID).
				Suffix(utils.PrepareSuffix(rColumns)).
				RunWith(db).QueryX(ctx)
			if err != nil {
				logger.Err(err).Msg("Error update user")
				return nil, nil, gqlerror.Errorf("Error update user")
			}
			return rows, u.UUID, nil
		} else {
			// todo Логика вставки
			rows, err := pglxqb.Insert("users").
				SetMap(setColumns).
				Suffix(utils.PrepareSuffix(rColumns)).
				RunWith(db).QueryX(ctx)
			if err != nil {
				logger.Err(err).Msg("Error insert user")
				return nil, nil, gqlerror.Errorf("Error insert user")
			}
			return rows, u.UUID, nil
		}
	}
	return nil, u.UUID, nil
}

func (u *User) ParseRows(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) ([]*User, error) {
	var users []*User
	logger := app.Logger.Error().Str("package", "models").Str("model", "user").Str("func", "ParseRows")
	// отработаем полученый колоки
	for rows.Next() {
		var user User
		if err := rows.StructScan(&user); err != nil {
			logger.Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
		users = append(users, &user)
	}
	// уберем лишние поля из запроса
	for _, user := range users {
		if err := user.ParseRequestedFields(ctx, fields, app, db); err != nil {
			logger.Err(err).Msg("Error parse requested fields for user")
			return nil, gqlerror.Errorf("Error parse requested fields for user")
		}
	}
	return users, nil
}

func (u *User) ParseRow(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) (*User, error) {
	var user User
	logger := app.Logger.Error().Str("package", "models").Str("model", "user").Str("func", "ParseRow")
	// разберем полученые поля, но вернем только последнее поле, так как дублей не может просто быть
	for rows.Next() {
		if err := rows.StructScan(&user); err != nil {
			logger.Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
	}
	// уберем лишние поля из запроса
	if err := user.ParseRequestedFields(ctx, fields, app, db); err != nil {
		logger.Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return &user, nil
}

func (u *User) ParseRequestedFields(ctx context.Context, fields []graphql.CollectedField, app *app.App, db pglxqb.BaseRunner) error {
	return parseRequestedFields(ctx, app, db, fields, u)
}

func (u *User) restoreStruct(ctx context.Context, app *app.App, db pglxqb.BaseRunner) error {
	v := reflect.ValueOf(u)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if !v.IsValid() {
		return nil
	}
	for i := 0; i < v.NumField(); i++ {
		if err := restoreStructReflect(ctx, app, db, v, v.Field(i), v.Type().Field(i)); err != nil {
			app.Logger.Error().Str("package", "models").Str("model", "user").Str("func", "restoreStruct").Err(err).Msg("Error restore struct user")
			return err
		}
	}
	return nil
}

func (u *User) GetByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID) (*User, error) {
	logger := app.Logger.Error().Str("package", "models").Str("model", "user").Str("func", "GetByUUID")
	rows, err := pglxqb.SelectAll().From("users").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		logger.Err(err).Msg("Error get user from DB by uuid")
		return nil, gqlerror.Errorf("Error get user from DB by uuid")
	}
	var user User
	for rows.Next() {
		if err := rows.StructScan(&user); err != nil {
			logger.Err(err).Msg("Error scan response to struct User")
			return nil, gqlerror.Errorf("Error scan response to struct User")
		}
	}
	return &user, nil
}

func (u *User) GetParsedObjectByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID, column graphql.CollectedField) (*User, error) {
	rows, err := pglxqb.SelectAll().From("users").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("package", "models").Str("model", "user").
			Str("func", "GetParsedObjectByUUID").Err(err).Msg("Error get user from DB by uuid")
		return nil, gqlerror.Errorf("Error get user from DB by uuid")
	}
	return u.ParseRow(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}

func (u *User) GetParsedObjectsByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid []*uuid.UUID, column graphql.CollectedField) ([]*User, error) {
	rows, err := pglxqb.SelectAll().From("users").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("package", "models").Str("model", "user").
			Str("func", "GetParsedObjectsByUUID").Err(err).Msg("Error get users from DB by uuid`s")
		return nil, gqlerror.Errorf("Error get users from DB by uuid`s")
	}
	return u.ParseRows(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}
