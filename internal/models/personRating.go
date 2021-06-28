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

type PersonRating struct {
	UUID        *uuid.UUID `json:"uuid" db:"uuid"`
	Created     *time.Time `json:"created" db:"created"`
	Updated     *time.Time `json:"updated" db:"updated"`
	IsDeleted   *bool      `json:"isDeleted" db:"is_deleted"`
	UUIDPerson  *uuid.UUID `db:"uuid_person"`
	Person      *Person    `json:"person" relay:"uuid_person" link:"UUIDPerson"`
	UUIDJob     *uuid.UUID `db:"uuid_job"`
	Job         *Job       `json:"job" relay:"uuid_job" link:"UUIDJob"`
	Rating      *float64   `json:"rating" db:"rating"`
	Description *string    `json:"description" db:"description"`
}

func (p *PersonRating) Mutation(ctx context.Context, db pglxqb.BaseRunner, app *app.App, rColumns interface{}, columns map[string]interface{}) (*pglx.Rows, *uuid.UUID, error) {
	logger := app.Logger.Error().Str("package", "models").Str("model", "PersonRating").Str("func", "Mutation")
	// Уберем лишние колонки
	if len(columns) > 0 {
		columns = utils.ClearSQLFields(p, columns)
	}
	update := false
	// если есть uuid значит манипулируем объектом
	if p.UUID != nil {
		if utils.CountFillFields(p) == 1 && len(columns) == 0 {
			return nil, p.UUID, nil
		}
		// получим Обьект
		user, err := p.GetByUUID(ctx, app, db, p.UUID)
		if err != nil {
			logger.Err(err).Msg("Error get PersonRating")
			return nil, nil, gqlerror.Errorf("Error get PersonRating")
		}
		// Если не меняются родители то вернем uuid
		if Compare(user, columns) && utils.CountFillFields(p) == 1 {
			return nil, p.UUID, nil
		}
		// сравним два Объекта
		utils.RestoreUUID(p, user)
		if err = p.restoreStruct(ctx, app, db); err != nil {
			logger.Err(err).Msg("Error restore struct PersonRating")
			return nil, nil, gqlerror.Errorf("Error restore struct PersonRating")
		}
		update = true
	} else {
		newUUID := uuid.New()
		p.UUID = &newUUID
		_, err := pglxqb.Insert("person_ratings").
			Columns("uuid").
			Values(newUUID).
			RunWith(db).Exec(ctx)
		if err != nil {
			logger.Err(err).Msg("Error create person rating")
			return nil, nil, gqlerror.Errorf("Error create person rating")
		}
		update = true
	}
	// Дополним поля связями c пользователями
	parentColumns := map[string]interface{}{}
	// сгенерим структуру для вставки
	setColumns, err := SqlGenKeys(ctx, app, db, p, columns, parentColumns)
	if err != nil {
		logger.Err(err).Msg("Error generate map of keys")
		return nil, nil, err
	}
	// Уберем лишние колонки
	setColumns = utils.ClearSQLFields(p, setColumns)
	// выполним мутацию
	if len(setColumns) > 0 {
		if update {
			// todo Логика Обновления
			// Обновляем иначе
			rows, err := pglxqb.Update("person_ratings").
				SetMap(setColumns).
				Where("uuid = ?", p.UUID).
				Suffix(utils.PrepareSuffix(rColumns)).
				RunWith(db).QueryX(ctx)
			if err != nil {
				logger.Err(err).Msg("Error update person PersonCourse")
				return nil, nil, gqlerror.Errorf("Error update person PersonCourse")
			}
			return rows, p.UUID, nil
		} else {
			// todo Логика вставки
			rows, err := pglxqb.Insert("person_ratings").
				SetMap(setColumns).
				Suffix(utils.PrepareSuffix(rColumns)).
				RunWith(db).QueryX(ctx)
			if err != nil {
				logger.Err(err).Msg("Error insert PersonCourse")
				return nil, nil, gqlerror.Errorf("Error insert PersonCourse")
			}
			return rows, p.UUID, nil
		}
	}
	return nil, p.UUID, nil
}

func (u *PersonRating) ParseRow(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) (*PersonRating, error) {
	var personRating PersonRating
	logger := app.Logger.Error().Str("package", "models").Str("model", "personCourse").Str("func", "ParseRow")
	// разберем полученые поля, но вернем только последнее поле, так как дублей не может просто быть
	for rows.Next() {
		if err := rows.StructScan(&personRating); err != nil {
			logger.Err(err).Msg("Error scan response to struct PersonCourse")
			return nil, gqlerror.Errorf("Error scan response to struct PersonCourse")
		}
	}
	// уберем лишние поля из запроса
	if err := personRating.parseRequestedFields(ctx, fields, app, db); err != nil {
		logger.Err(err).Msg("Error scan response to struct PersonCourse")
		return nil, gqlerror.Errorf("Error scan response to struct PersonCourse")
	}
	return &personRating, nil
}

func (p *PersonRating) ParseRows(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) ([]*PersonRating, error) {
	var personRatings []*PersonRating
	defer rows.Close()
	for rows.Next() {
		var personRating PersonRating
		if err := rows.StructScan(&personRating); err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct PersonCourse")
			return nil, gqlerror.Errorf("Error scan response to struct PersonCourse")
		}

		personRatings = append(personRatings, &personRating)
	}
	for _, personRating := range personRatings {
		if err := personRating.parseRequestedFields(ctx, fields, app, db); err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct PersonCourse")
			return nil, gqlerror.Errorf("Error scan response to struct PersonCourse")
		}
	}
	return personRatings, nil
}

func (p *PersonRating) parseRequestedFields(ctx context.Context, fields []graphql.CollectedField, app *app.App, db pglxqb.BaseRunner) error {
	return parseRequestedFields(ctx, app, db, fields, p)
}

func (p *PersonRating) restoreStruct(ctx context.Context, app *app.App, db pglxqb.BaseRunner) error {
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

func (p *PersonRating) GetByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID) (*PersonRating, error) {
	rows, err := pglxqb.SelectAll().From("person_ratings").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error get person course from DB")
		return nil, gqlerror.Errorf("Error get person course from DB")
	}
	count := 0
	var personRating PersonRating
	for rows.Next() {
		count++
		if err := rows.StructScan(&personRating); err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct PersonCourse")
			return nil, gqlerror.Errorf("Error scan response to struct PersonCourse")
		}
	}
	if count == 0 {
		return nil, gqlerror.Errorf("Error no person rating found by UUID")
	}
	return &personRating, nil
}

func (p *PersonRating) GetParsedObjectByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID, column graphql.CollectedField) (*PersonRating, error) {
	rows, err := pglxqb.SelectAll().From("person_ratings").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error get person rating by uuid")
		return nil, gqlerror.Errorf("Error get person rating by uuid")
	}
	return p.ParseRow(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}

func (p *PersonRating) GetParsedObjectsByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid []*uuid.UUID, column graphql.CollectedField) ([]*PersonRating, error) {
	rows, err := pglxqb.SelectAll().From("person_ratings").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error get person rating by array of uuids")
		return nil, gqlerror.Errorf("Error get person rating by array of uuids")
	}
	return p.ParseRows(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}
