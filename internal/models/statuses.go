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

type Status struct {
	UUID         *uuid.UUID   `json:"uuid" db:"uuid"`
	Created      *time.Time   `json:"created" db:"created"`
	Updated      *time.Time   `json:"updated" db:"updated"`
	UUIDPerson   *uuid.UUID   `db:"uuid_person"`
	Person       *Person      `json:"person" relay:"uuid_person" link:"UUIDPerson"`
	UUIDJob      *uuid.UUID   `db:"uuid_job"`
	Job          *Job         `json:"job" relay:"uuid_job" link:"UUIDJob"`
	Description  *string      `json:"description" db:"description"`
	UUIDContent  []*uuid.UUID `db:"uuid_content"`
	Content      []*Content   `json:"content" relay:"uuid_content" link:"UUIDContent"`
	UUIDTags     []*uuid.UUID `db:"uuid_tags"`
	Tags         []*Tag       `json:"tags" relay:"uuid_tags" link:"UUIDTags"`
	IsDeleted    *bool        `json:"isDeleted" db:"is_deleted"`
	Status       *JobStatus   `json:"status" db:"status"`
	Lat          *float64     `json:"lat" db:"lat"`
	Lon          *float64     `json:"lon" db:"lon"`
	UUIDExecutor *uuid.UUID   `db:"uuid_executor"`
}

func (s *Status) Mutation(ctx context.Context, db pglxqb.BaseRunner, app *app.App, rColumns interface{}, columns map[string]interface{}) (*pglx.Rows, *uuid.UUID, error) {
	update := false
	// Уберем лишние колонки
	if len(columns) > 0 {
		columns = utils.ClearSQLFields(s, columns)
	}
	// если есть uuid значит манипулируем обектом
	if s.UUID != nil {
		if utils.CountFillFields(s) == 1 && len(columns) == 0 {
			return nil, s.UUID, nil
		}
		status, err := s.GetByUUID(ctx, app, db, s.UUID)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "Mutation").Err(err).Msg("Error get status")
			return nil, nil, gqlerror.Errorf("Error get status")
		}
		if Compare(status, columns) && utils.CountFillFields(s) == 1 {
			return nil, s.UUID, nil
		}
		// воcстановим все ссылки
		utils.RestoreUUID(s, status)
		// восстановим подчиненные структуры
		if err = status.restoreStruct(ctx, app, db); err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "Mutation").Err(err).Msg("Error restore struct person")
			return nil, nil, gqlerror.Errorf("Error restore struct person")
		}
		update = true
	} else {
		// иначе создадим с нуля объект
		newUUID := uuid.New()
		s.UUID = &newUUID
		_, err := pglxqb.Insert("users").
			Columns("uuid").
			Values(newUUID).
			RunWith(db).Exec(ctx)
		if err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "Mutation").Err(err).Msg("Error insert user")
			return nil, nil, gqlerror.Errorf("Error insert user")
		}
		update = true
	}
	// дополним пропущеные поля, если они есть
	parent := make(map[string]interface{})
	setColumns, err := SqlGenKeys(ctx, app, db, s, columns, parent)
	if err != nil {
		app.Logger.Error().Str("module", "persons").Str("func", "Mutation").Err(err).Msg("Error generate map of keys")
		return nil, nil, err
	}
	// только одна колонка, и это uuid то удаляем обьект
	setColumns = utils.ClearSQLFields(s, setColumns)
	if len(setColumns) > 0 {
		if update {
			// Обновляем иначе
			rows, err := pglxqb.Update("statuses").
				SetMap(setColumns).
				Where("uuid = ?", s.UUID).
				Suffix(utils.PrepareSuffix(rColumns)).
				RunWith(db).QueryX(ctx)
			if err != nil {
				app.Logger.Error().Str("module", "contact").Str("func", "Mutation").Err(err).Msg("Error update contact")
				return nil, nil, gqlerror.Errorf("Error update contact")
			}
			return rows, s.UUID, nil
		} else {
			rows, err := pglxqb.Insert("statuses").
				SetMap(setColumns).
				Suffix(utils.PrepareSuffix(rColumns)).
				RunWith(db).QueryX(ctx)
			if err != nil {
				app.Logger.Error().Str("module", "contact").Str("func", "Mutation").Err(err).Msg("Error insert contact")
				return nil, nil, gqlerror.Errorf("Error insert contact")
			}
			return rows, s.UUID, nil
		}
	}
	return nil, s.UUID, nil
}

func (s *Status) ParseRows(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) ([]*Status, error) {
	var statuses []*Status
	defer rows.Close()
	for rows.Next() {
		var status Status
		err := rows.StructScan(&status)
		if err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
		statuses = append(statuses, &status)
	}
	for _, status := range statuses {
		if err := status.parseRequestedFields(ctx, fields, app, db); err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
	}
	return statuses, nil
}

func (s *Status) ParseRow(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) (*Status, error) {
	var err error
	var status Status
	defer rows.Close()
	for rows.Next() {
		err = rows.StructScan(&status)
		if err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
	}
	err = status.parseRequestedFields(ctx, fields, app, db)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return &status, nil
}

func (s *Status) parseRequestedFields(ctx context.Context, fields []graphql.CollectedField, app *app.App, db pglxqb.BaseRunner) error {
	return parseRequestedFields(ctx, app, db, fields, s)
}

func (s *Status) restoreStruct(ctx context.Context, app *app.App, db pglxqb.BaseRunner) error {
	v := reflect.ValueOf(s)
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

func (s *Status) GetByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID) (*Status, error) {
	rows, err := pglxqb.SelectAll().From("statuses").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error get person from DB")
		return nil, gqlerror.Errorf("Error get person from DB")
	}
	var status Status
	defer rows.Close()
	for rows.Next() {
		if err := rows.StructScan(&status); err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct Person")
			return nil, gqlerror.Errorf("Error scan response to struct Person")
		}
	}
	return &status, nil
}

func (s *Status) GetParsedObjectByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID, column graphql.CollectedField) (*Status, error) {
	rows, err := pglxqb.SelectAll().From("statuses").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return s.ParseRow(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}

func (s *Status) GetParsedObjectsByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid []*uuid.UUID, column graphql.CollectedField) ([]*Status, error) {
	rows, err := pglxqb.SelectAll().From("statuses").Where(pglxqb.Eq{"uuid": uuid}).OrderBy("created DESC").RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return s.ParseRows(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}
