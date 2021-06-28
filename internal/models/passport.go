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

type Passport struct {
	UUID                    *uuid.UUID             `json:"uuid" db:"uuid"`
	Serial                  *string                `json:"serial" db:"serial"`
	Number                  *string                `json:"number" db:"number"`
	Department              *string                `json:"department" db:"department"`
	DateIssue               *time.Time             `json:"dateIssue" db:"date_issue"`
	DepartmentCode          *string                `json:"departmentCode" db:"department_code"`
	UUIDPerson              *uuid.UUID             `db:"uuid_person"`
	Person                  *Person                `json:"person" relay:"uuid_person" link:"UUIDPerson"`
	Created                 *time.Time             `json:"created" db:"created"`
	Updated                 *time.Time             `json:"updated" db:"updated"`
	UUIDScan                *uuid.UUID             `db:"uuid_scan"`
	Scan                    *Content               `json:"scan" relay:"uuid_scan" link:"UUIDScan"`
	UUIDAddressRegistration *uuid.UUID             `db:"uuid_address_registration"`
	AddressRegistration     *Address               `json:"addressRegistration" relay:"uuid_address_registration" link:"UUIDAddressRegistration"`
	UUIDPhotoRegistration   *uuid.UUID             `db:"uuid_photo_registration"`
	PhotoRegistration       *Content               `json:"photoRegistration" relay:"uuid_photo_registration" link:"UUIDPhotoRegistration"`
	IsDeleted        *bool         `json:"isDeleted" db:"is_deleted"`
}

type PassportFilter struct {
	Created        *DateTimeFilter `json:"created" db:"created"`
	Updated        *DateTimeFilter `json:"updated" db:"updated"`
	Serial         *StringFilter   `json:"serial" db:"serial"`
	Number         *StringFilter   `json:"number" db:"number"`
	DepartmentCode *StringFilter   `json:"departmentCode" db:"department_code"`
	Department     *StringFilter   `json:"department" db:"department"`
	DateIssue      *DateFilter     `json:"dateIssue" db:"date_issue"`
	And            []*UserFilter   `json:"and"`
	Or             []*UserFilter   `json:"or"`
	Not            *UserFilter     `json:"not"`
}

type PassportSort struct {
	Field *UserSortableField `json:"field"`
	Order *SortOrder         `json:"order"`
}

type PassportSortableField string

const (
	PassportSortableFieldUUID           PassportSortableField = "uuid"
	PassportSortableFieldCreated        PassportSortableField = "created"
	PassportSortableFieldUpdated        PassportSortableField = "updated"
	PassportSortableFieldSerial         PassportSortableField = "serial"
	PassportSortableFieldNumber         PassportSortableField = "number"
	PassportSortableFieldDateIssue      PassportSortableField = "dateIssue"
	PassportSortableFieldDepartmentCode PassportSortableField = "departmentCode"
	PassportSortableFieldDepartment     PassportSortableField = "department"
)

func (e PassportSortableField) IsValid() bool {
	switch e {
	case PassportSortableFieldUUID, PassportSortableFieldCreated, PassportSortableFieldUpdated, PassportSortableFieldSerial, PassportSortableFieldNumber, PassportSortableFieldDateIssue, PassportSortableFieldDepartmentCode, PassportSortableFieldDepartment:
		return true
	}
	return false
}

func (e PassportSortableField) String() string {
	return string(e)
}

func (e *PassportSortableField) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = PassportSortableField(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid PassportSortableField", str)
	}
	return nil
}

func (e PassportSortableField) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

func (p *Passport) Mutation(ctx context.Context, db pglxqb.BaseRunner, app *app.App, rColumns interface{}, columns map[string]interface{}) (*pglx.Rows, *uuid.UUID, error) {
	updateOrDelete := false
	// если есть uuid значит манипулируем обектом
	if p.UUID != nil {
		// Дополним поля связями
		passport, err := p.GetByUUID(ctx, app, db, p.UUID)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "Mutation").Err(err).Msg("Error get passport")
			return nil, nil, gqlerror.Errorf("Error get passport")
		}
		utils.RestoreUUID(p, passport)
		if err = passport.restoreStruct(ctx, app, db); err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "Mutation").Err(err).Msg("Error restore struct Passport")
			return nil, nil, gqlerror.Errorf("Error restore struct Passport")
		}
		updateOrDelete = true
	} else {
		// иначе создадим с нуля Обьект
		newUUID := uuid.New()
		p.UUID = &newUUID
		columns["uuid"] = newUUID
	}
	parentColumns := map[string]interface{}{"uuid_passport": p.UUID}
	// дополним пропущеные поля, если они есть
	setColumns, err := SqlGenKeys(ctx, app, db, p, columns, parentColumns)
	if err != nil {
		app.Logger.Error().Str("module", "persons").Str("func", "Mutation").Err(err).Msg("Error generate map of keys")
		return nil, nil, err
	}
	// только одна колонка, и это uuid то удаляем обьект
	setColumns = utils.ClearSQLFields(p, setColumns)
	if len(setColumns) > 0 {
		if updateOrDelete {
			// todo Логика Обновления
			// Обновляем иначе
			rows, err := pglxqb.Update("passports").
				SetMap(setColumns).
				Where("uuid = ?", p.UUID).
				Suffix(utils.PrepareSuffix(rColumns)).
				RunWith(db).QueryX(ctx)
			if err != nil {
				app.Logger.Error().Str("module", "passports").Str("func", "manipulate").Err(err).Msg("Error update passport")
				return nil, nil, gqlerror.Errorf("Error update passport")
			}
			return rows, p.UUID, nil
		} else {
			// todo Логика вставки
			rows, err := pglxqb.Insert("passports").
				SetMap(setColumns).
				Suffix(utils.PrepareSuffix(rColumns)).
				RunWith(db).QueryX(ctx)
			if err != nil {
				app.Logger.Error().Str("module", "passports").Str("func", "manipulate").Err(err).Msg("Error insert passport")
				return nil, nil, gqlerror.Errorf("Error insert passport")
			}
			return rows, p.UUID, nil
		}
	}
	return nil, p.UUID, nil
}

func (p *Passport) ParseRows(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) ([]*Passport, error) {
	var passports []*Passport
	defer rows.Close()
	for rows.Next() {
		var passport Passport
		err := rows.StructScan(&passport)
		if err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct person")
			return nil, gqlerror.Errorf("Error scan response to struct person")
		}
		err = passport.parseRequestedFields(ctx, fields, app, db)
		if err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct person")
			return nil, gqlerror.Errorf("Error scan response to struct person")
		}
		passports = append(passports, &passport)
	}
	return passports, nil
}

func (p *Passport) ParseRow(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) (*Passport, error) {
	var err error
	var passport Passport
	defer rows.Close()
	for rows.Next() {
		err = rows.StructScan(&passport)
		if err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct person")
			return nil, gqlerror.Errorf("Error scan response to struct person")
		}
	}
	err = passport.parseRequestedFields(ctx, fields, app, db)
	if err != nil {
		app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct person")
		return nil, gqlerror.Errorf("Error scan response to struct person")
	}
	return &passport, nil
}

func (p *Passport) parseRequestedFields(ctx context.Context, fields []graphql.CollectedField, app *app.App, db pglxqb.BaseRunner) error {
	return parseRequestedFields(ctx, app, db, fields, p)
}

func (p *Passport) restoreStruct (ctx context.Context, app *app.App, db pglxqb.BaseRunner) error {
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

func (p *Passport) GetByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID) (*Passport, error) {
	rows, err := pglxqb.SelectAll().From("passports").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error get person from DB")
		return nil, gqlerror.Errorf("Error get person from DB")
	}
	var passport Passport
	defer rows.Close()
	for rows.Next() {
		if err := rows.StructScan(&passport); err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct Person")
			return nil, gqlerror.Errorf("Error scan response to struct Person")
		}
	}
	return &passport, nil
}

func (p *Passport) GetParsedObjectByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID, column graphql.CollectedField) (*Passport, error) {
	rows, err := pglxqb.SelectAll().From("passports").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return p.ParseRow(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}

func (p *Passport) GetParsedObjectsByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid []*uuid.UUID, column graphql.CollectedField) ([]*Passport, error) {
	rows, err := pglxqb.SelectAll().From("passports").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return p.ParseRows(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}