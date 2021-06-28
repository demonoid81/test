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
	"reflect"
	"time"
)

type JobType struct {
	UUID                *uuid.UUID         `json:"uuid" db:"uuid"`
	Created             *time.Time         `json:"created" db:"created"`
	Updated             *time.Time         `json:"updated" db:"updated"`
	UUIDOrganization    *uuid.UUID         `db:"uuid_organization"`
	Organization        *Organization      `json:"organization" relay:"uuid_organization" link:"UUIDOrganization"`
	Name                *string            `json:"name" db:"name"`
	UUIDLocalityJobCost []*uuid.UUID       `db:"uuid_locality_job_cost"`
	LocalityJobCost     []*LocalityJobCost `json:"localityJobCost" relay:"uuid_locality_job_cost" link:"UUIDLocalityJobCost"`
	UUIDCourses         []*uuid.UUID       `db:"uuid_courses"`
	Courses             []*Course          `json:"courses" relay:"uuid_courses" link:"UUIDCourses"`
	IsDeleted           *bool              `json:"isDeleted" db:"is_deleted"`
	NeedMedicalBook     *bool              `json:"needMedicalBook" db:"need_medical_book"`
	Icon                *JobTypeIcon       `json:"jobTypeIcon" db:"job_type_icon"`
}

type JobTypeFilter struct {
	UUID            *UUIDFilter         `json:"uuid" db:"uuid"`
	Created         *DateTimeFilter     `json:"created" db:"created"`
	Updated         *DateTimeFilter     `json:"updated" db:"updated"`
	Icon            *JobTypeIcon        `json:"icon" db:"icon"`
	Organization    *OrganizationFilter `json:"organization" table:"organizations" link:"uuid_organization"`
	Name            *StringFilter       `json:"name" db:"name"`
	IsDeleted       *bool               `json:"isDeleted" db:"is_deleted"`
	NeedMedicalBook *bool               `json:"needMedicalBook" db:"need_medical_book"`
}

func (jt *JobType) Mutation(ctx context.Context, db pglxqb.BaseRunner, app *app.App, rColumns interface{}, columns map[string]interface{}) (*pglx.Rows, *uuid.UUID, error) {
	fmt.Println(jt)
	updateOrDelete := false
	// если есть uuid значит манипулируем обектом
	//if len(columns) == 0 && jt.UUID != nil {
	//	return nil, jt.UUID, nil
	//}
	if jt.UUID != nil {
		jobType, err := jt.GetByUUID(ctx, app, db, jt.UUID)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "Mutation").Err(err).Msg("Error get person")
			return nil, nil, gqlerror.Errorf("Error get person")
		}
		// востановим все ссылки
		utils.RestoreUUID(jt, jobType)
		// востановим подчиненные структуры
		if err = jobType.restoreStruct(ctx, app, db); err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "Mutation").Err(err).Msg("Error restore struct person")
			return nil, nil, gqlerror.Errorf("Error restore struct person")
		}
		updateOrDelete = true
		fmt.Println("restore uuid")
	} else {
		// иначе создадим с нуля Обьект
		newUUID := uuid.New()
		jt.UUID = &newUUID
		columns["uuid"] = newUUID
	}
	// дополним пропущеные поля, если они есть
	parent := make(map[string]interface{})
	setColumns, err := SqlGenKeys(ctx, app, db, jt, columns, parent)
	if err != nil {
		app.Logger.Error().Str("module", "persons").Str("func", "Mutation").Err(err).Msg("Error generate map of keys")
		return nil, nil, err
	}
	// только одна колонка, и это uuid то удаляем обьект
	setColumns = utils.ClearSQLFields(jt, setColumns)
	if len(setColumns) > 0 {
		if updateOrDelete {
			// Обновляем иначе
			rows, err := pglxqb.
				Update("job_types").
				SetMap(setColumns).
				Where("uuid = ?", jt.UUID).
				Suffix(utils.PrepareSuffix(rColumns)).
				RunWith(db).
				QueryX(ctx)
			if err != nil {
				app.Logger.Error().Str("module", "contact").Str("func", "Mutation").Err(err).Msg("Error update contact")
				return nil, nil, gqlerror.Errorf("Error update contact")
			}
			return rows, jt.UUID, nil
		} else {
			rows, err := pglxqb.
				Insert("job_types").
				SetMap(setColumns).
				Suffix(utils.PrepareSuffix(rColumns)).
				RunWith(db).
				QueryX(ctx)
			if err != nil {
				app.Logger.Error().Str("module", "contact").Str("func", "Mutation").Err(err).Msg("Error insert contact")
				return nil, nil, gqlerror.Errorf("Error insert contact")
			}
			return rows, jt.UUID, nil
		}
	}
	return nil, jt.UUID, nil
}

func (jt *JobType) ParseRows(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) ([]*JobType, error) {
	var jobTypes []*JobType
	defer rows.Close()
	for rows.Next() {
		var jobType JobType
		err := rows.StructScan(&jobType)
		if err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
		err = jobType.parseRequestedFields(ctx, fields, app, db)
		if err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
		jobTypes = append(jobTypes, &jobType)
	}
	return jobTypes, nil
}

func (jt *JobType) ParseRow(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) (*JobType, error) {
	var err error
	var jobType JobType
	for rows.Next() {
		err = rows.StructScan(&jobType)
		if err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
	}
	err = jobType.parseRequestedFields(ctx, fields, app, db)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return &jobType, nil
}

func (jt *JobType) parseRequestedFields(ctx context.Context, fields []graphql.CollectedField, app *app.App, db pglxqb.BaseRunner) error {
	return parseRequestedFields(ctx, app, db, fields, jt)
}

func (jt *JobType) restoreStruct(ctx context.Context, app *app.App, db pglxqb.BaseRunner) error {
	v := reflect.ValueOf(jt)
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

func (jt *JobType) GetByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID) (*JobType, error) {
	rows, err := pglxqb.SelectAll().From("job_types").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error get person from DB")
		return nil, gqlerror.Errorf("Error get person from DB")
	}
	var jobType JobType
	defer rows.Close()
	for rows.Next() {
		if err := rows.StructScan(&jobType); err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct Person")
			return nil, gqlerror.Errorf("Error scan response to struct Person")
		}
	}
	return &jobType, nil
}

func (jt *JobType) GetParsedObjectByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID, column graphql.CollectedField) (*JobType, error) {
	rows, err := pglxqb.Select("job_types.*").From("job_types").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return jt.ParseRow(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}

func (jt *JobType) GetParsedObjectsByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid []*uuid.UUID, column graphql.CollectedField) ([]*JobType, error) {
	rows, err := pglxqb.Select("job_types.*").From("job_types").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return jt.ParseRows(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}
