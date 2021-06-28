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

type CourseType string

const (
	CourseTypeCourse   CourseType = "course"
	CourseTypeBriefing CourseType = "briefing"
)

func (e CourseType) IsValid() bool {
	switch e {
	case CourseTypeCourse, CourseTypeBriefing:
		return true
	}
	return false
}

func (e CourseType) String() string {
	return string(e)
}

func (e *CourseType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = CourseType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid CourseType", str)
	}
	return nil
}

func (e CourseType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type Course struct {
	UUID             *uuid.UUID    `json:"uuid" db:"uuid"`
	Created          *time.Time    `json:"created" db:"created"`
	Updated          *time.Time    `json:"updated" db:"updated"`
	IsDeleted        *bool         `json:"isDeleted" db:"is_deleted"`
	Name             *string       `json:"name" db:"name"`
	CourseType       CourseType    `json:"courseType" db:"course_type"`
	Content          *string       `json:"content" db:"content"`
	UUIDOrganization *uuid.UUID    `db:"uuid_organization"`
	Organization     *Organization `json:"organization" relay:"uuid_organization" link:"UUIDOrganization"`
	PassingScore     *int          `json:"passingScore" db:"passing_score"`
}

type CourseFilter struct {
	UUID         *UUIDFilter         `json:"uuid" db:"uuid"`
	Created      *DateTimeFilter     `json:"created" db:"created"`
	Updated      *DateTimeFilter     `json:"updated" db:"updated"`
	IsDeleted    *bool               `json:"isDeleted" db:"is_deleted"`
	Name         *StringFilter       `json:"name" db:"name"`
	CourseType   *CourseType         `json:"courseType" db:"course_type"`
	Content      *StringFilter       `json:"content" db:"content"`
	Organization *OrganizationFilter `json:"organization" table:"organizations" link:"uuid_organization"`
	PassingScore *IntFilter          `json:"passingScore" db:"passing_score"`
}

func (c *Course) Mutation(ctx context.Context, db pglxqb.BaseRunner, app *app.App, rColumns interface{}, columns map[string]interface{}) (*pglx.Rows, *uuid.UUID, error) {
	// Уберем лишние колонки
	if len(columns) > 0 {
		columns = utils.ClearSQLFields(c, columns)
	}
	update := false
	// если есть uuid значит манипулируем обектом
	if c.UUID != nil {
		fmt.Println(c)
		if utils.CountFillFields(c) == 1 && len(columns) == 0 {
			return nil, c.UUID, nil
		}
		course, err := c.GetByUUID(ctx, app, db, c.UUID)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "Mutation").Err(err).Msg("Error get person")
			return nil, nil, gqlerror.Errorf("Error get person")
		}
		// Если не меняются родители то вернем uuid
		if Compare(course, columns) && utils.CountFillFields(c) == 1 {
			return nil, c.UUID, nil
		}
		// востановим все ссылки
		utils.RestoreUUID(c, course)
		// востановим подчиненные структуры
		if err = c.restoreStruct(ctx, app, db); err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "Mutation").Err(err).Msg("Error restore struct person")
			return nil, nil, gqlerror.Errorf("Error restore struct person")
		}
		update = true
	} else {
		// иначе создадим с нуля Обьект
		newUUID := uuid.New()
		c.UUID = &newUUID
		columns["uuid"] = newUUID
	}
	// дополним пропущеные поля, если они есть
	parent := make(map[string]interface{})
	setColumns, err := SqlGenKeys(ctx, app, db, c, columns, parent)
	if err != nil {
		app.Logger.Error().Str("module", "persons").Str("func", "Mutation").Err(err).Msg("Error generate map of keys")
		return nil, nil, err
	}
	// только одна колонка, и это uuid то удаляем обьект
	setColumns = utils.ClearSQLFields(c, setColumns)
	if len(setColumns) > 0 {
		if update {
			// Обновляем иначе
			rows, err := pglxqb.Update("courses").
				SetMap(setColumns).
				Where("uuid = ?", c.UUID).
				Suffix(utils.PrepareSuffix(rColumns)).
				RunWith(db).QueryX(ctx)
			if err != nil {
				app.Logger.Error().Str("module", "contact").Str("func", "Mutation").Err(err).Msg("Error update contact")
				return nil, nil, gqlerror.Errorf("Error update contact")
			}
			return rows, c.UUID, nil
		} else {
			rows, err := pglxqb.Insert("courses").
				SetMap(setColumns).
				Suffix(utils.PrepareSuffix(rColumns)).
				RunWith(db).QueryX(ctx)
			if err != nil {
				app.Logger.Error().Str("module", "contact").Str("func", "Mutation").Err(err).Msg("Error insert contact")
				return nil, nil, gqlerror.Errorf("Error insert contact")
			}
			return rows, c.UUID, nil
		}
	}
	return nil, c.UUID, nil
}

func (c *Course) ParseRows(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) ([]*Course, error) {
	var Courses []*Course
	defer rows.Close()
	for rows.Next() {
		var course Course
		if err := rows.StructScan(&course); err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
		Courses = append(Courses, &course)
	}
	for _, course := range Courses {
		if err := course.parseRequestedFields(ctx, fields, app, db); err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
	}
	return Courses, nil
}

func (c *Course) ParseRow(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) (*Course, error) {
	var course Course
	for rows.Next() {
		if err := rows.StructScan(&course); err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
	}
	if err := course.parseRequestedFields(ctx, fields, app, db); err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return &course, nil
}

func (c *Course) parseRequestedFields(ctx context.Context, fields []graphql.CollectedField, app *app.App, db pglxqb.BaseRunner) error {
	return parseRequestedFields(ctx, app, db, fields, c)
}

func (c *Course) restoreStruct(ctx context.Context, app *app.App, db pglxqb.BaseRunner) error {
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

func (c *Course) GetByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID) (*Course, error) {
	rows, err := pglxqb.SelectAll().From("courses").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error get person from DB")
		return nil, gqlerror.Errorf("Error get person from DB")
	}
	var course Course
	defer rows.Close()
	for rows.Next() {
		if err := rows.StructScan(&course); err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct Person")
			return nil, gqlerror.Errorf("Error scan response to struct Person")
		}
	}
	return &course, nil
}

func (c *Course) GetParsedObjectByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID, column graphql.CollectedField) (*Course, error) {
	rows, err := pglxqb.SelectAll().From("courses").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return c.ParseRow(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}

func (c *Course) GetParsedObjectsByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid []*uuid.UUID, column graphql.CollectedField) ([]*Course, error) {
	rows, err := pglxqb.SelectAll().From("courses").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return c.ParseRows(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}
