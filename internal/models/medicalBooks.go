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
	"reflect"
	"time"
)

type MedicalBook struct {
	UUID                          *uuid.UUID   `json:"uuid" db:"uuid"`
	Number                        *string      `json:"number" db:"number"`
	MedicalExaminationDate        *time.Time   `json:"medicalExaminationDate" db:"medical_examination_date"`
	UUIDContents                  []*uuid.UUID `db:"uuid_contents"`
	Contents                      []*Content   `json:"contents" relay:"uuid_contents" link:"UUIDContents"`
	Person                        *Person      `json:"person" relay:"uuid_person" link:"UUIDPerson"`
	Created                       *time.Time   `json:"created" db:"created"`
	Updated                       *time.Time   `json:"updated" db:"updated"`
	IsDeleted                     *bool        `json:"isDeleted" db:"is_deleted"`
	HaveHealthRestrictions        *bool        `json:"haveHealthRestrictions" db:"have_health_restrictions"`
	HaveMedicalBook               *bool        `json:"haveMedicalBook" db:"have_medical_book"`
	DescriptionHealthRestrictions *string      `json:"descriptionHealthRestrictions" db:"description_health_restrictions"`
	Checked                       *bool        `json:"checked" db:"checked"`
	CheckedDate                   *time.Time   `json:"checkedDate" db:"checked_date"`
	UUIDCheckedPerson             *uuid.UUID   `db:"uuid_checked_person"`
	CheckedPerson                 *Person      `json:"checkedPerson" relay:"uuid_checked_person" link:"UUIDCheckedPerson"`
}

func (mb *MedicalBook) Mutation(ctx context.Context, db pglxqb.BaseRunner, app *app.App, rColumns interface{}, columns map[string]interface{}) (*pglx.Rows, *uuid.UUID, error) {
	// если есть uuid значит манипулируем обектом
	if mb.UUID != nil {
		// Дополним поля связями
		medicalBook, err := mb.GetByUUID(ctx, app, db, mb.UUID)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "Mutation").Err(err).Msg("Error get medicalBook")
			return nil, nil, gqlerror.Errorf("Error get medicalBook")
		}
		utils.RestoreUUID(mb, medicalBook)
		if err = medicalBook.restoreStruct(ctx, app, db); err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "Mutation").Err(err).Msg("Error restore struct MedicalBook")
			return nil, nil, gqlerror.Errorf("Error restore struct MedicalBook")
		}
	} else {
		// иначе создадим с нуля Обьект
		newUUID := uuid.New()
		mb.UUID = &newUUID
		if _, err := pglxqb.Insert("medical_books").
			Columns("uuid").
			Values(newUUID).
			RunWith(db).Exec(ctx); err != nil {
		}
	}
	// дополним пропущеные поля, если они есть
	parent := make(map[string]interface{})
	setColumns, err := SqlGenKeys(ctx, app, db, mb, columns, parent)
	if err != nil {
		app.Logger.Error().Str("module", "persons").Str("func", "Mutation").Err(err).Msg("Error generate map of keys")
		return nil, nil, err
	}
	// только одна колонка, и это uuid то удаляем обьект
	setColumns = utils.ClearSQLFields(mb, setColumns)
	if len(setColumns) > 0 {
		// todo Логика Обновления
		// Обновляем иначе
		rows, err := pglxqb.Update("medical_books").
			SetMap(setColumns).
			Where("uuid = ?", mb.UUID).
			Suffix(utils.PrepareSuffix(rColumns)).
			RunWith(db).QueryX(ctx)
		if err != nil {
			app.Logger.Error().Str("module", "medicalBooks").Str("func", "Mutation").Err(err).Msg("Error update medicalBook")
			return nil, nil, gqlerror.Errorf("Error update medicalBook")
		}
		return rows, mb.UUID, nil
	}
	return nil, mb.UUID, nil
}

func (mb *MedicalBook) ParseRows(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) ([]*MedicalBook, error) {
	var medicalBooks []*MedicalBook
	defer rows.Close()
	for rows.Next() {
		var medicalBook MedicalBook
		err := rows.StructScan(&medicalBook)
		if err != nil {
			app.Logger.Error().Str("module", "medicalBooks").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct medicalBook")
			return nil, gqlerror.Errorf("Error scan response to struct medicalBook")
		}
		err = medicalBook.parseRequestedFields(ctx, fields, app, db)
		if err != nil {
			app.Logger.Error().Str("module", "medicalBooks").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct medicalBook")
			return nil, gqlerror.Errorf("Error scan response to struct medicalBook")
		}
		medicalBooks = append(medicalBooks, &medicalBook)
	}
	return medicalBooks, nil
}

func (mb *MedicalBook) ParseRow(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) (*MedicalBook, error) {
	var err error
	var medicalBook MedicalBook
	defer rows.Close()
	for rows.Next() {
		err = rows.StructScan(&medicalBook)
		if err != nil {
			app.Logger.Error().Str("module", "medicalBooks").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct medicalBook")
			return nil, gqlerror.Errorf("Error scan response to struct person")
		}
	}
	err = medicalBook.parseRequestedFields(ctx, fields, app, db)
	if err != nil {
		app.Logger.Error().Str("module", "medicalBooks").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct medicalBook")
		return nil, gqlerror.Errorf("Error scan response to struct medicalBook")
	}
	return &medicalBook, nil
}

func (mb *MedicalBook) parseRequestedFields(ctx context.Context, fields []graphql.CollectedField, app *app.App, db pglxqb.BaseRunner) error {
	return parseRequestedFields(ctx, app, db, fields, mb)
}

func (mb *MedicalBook) restoreStruct(ctx context.Context, app *app.App, db pglxqb.BaseRunner) error {
	v := reflect.ValueOf(mb)
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

func (mb *MedicalBook) GetByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID) (*MedicalBook, error) {
	rows, err := pglxqb.SelectAll().From("medical_books").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error get person from DB")
		return nil, gqlerror.Errorf("Error get person from DB")
	}
	var medicalBook MedicalBook
	defer rows.Close()
	for rows.Next() {
		if err := rows.StructScan(&medicalBook); err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct Person")
			return nil, gqlerror.Errorf("Error scan response to struct Person")
		}
	}
	return &medicalBook, nil
}

func (mb *MedicalBook) GetParsedObjectByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID, column graphql.CollectedField) (*MedicalBook, error) {
	rows, err := pglxqb.SelectAll().From("medical_books").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return mb.ParseRow(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}

func (mb *MedicalBook) GetParsedObjectsByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid []*uuid.UUID, column graphql.CollectedField) ([]*MedicalBook, error) {
	rows, err := pglxqb.SelectAll().From("medical_books").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return mb.ParseRows(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}
