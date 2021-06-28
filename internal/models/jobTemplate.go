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

type JobTemplate struct {
	UUID             *uuid.UUID    `json:"uuid" db:"uuid"`
	Created          *time.Time    `json:"created" db:"created"`
	Updated          *time.Time    `json:"updated" db:"updated"`
	Name             *string       `json:"name" db:"name"`
	UUIDOrganization *uuid.UUID    `db:"uuid_organization"`
	Organization     *Organization `json:"organization" relay:"uuid_organization" link:"UUIDOrganization"`
	UUIDObject       *uuid.UUID    `db:"uuid_object"`
	Object           *Organization `json:"object" relay:"uuid_object" link:"UUIDObject"`
	UUIDRegion       *uuid.UUID    `db:"uuid_region"`
	Region           *Region       `json:"region" relay:"uuid_region" link:"UUIDRegion"`
	UUIDArea         *uuid.UUID    `db:"uuid_area"`
	Area             *Area         `json:"area" relay:"uuid_area" link:"UUIDArea"`
	UUIDCity         *uuid.UUID    `db:"uuid_city"`
	City             *City         `json:"city" relay:"uuid_city" link:"UUIDCity"`
	UUIDJobType      *uuid.UUID    `db:"uuid_job_type"`
	JobType          *JobType      `json:"jobType" relay:"uuid_job_type" link:"UUIDJobType"`
	Cost             *float64      `json:"cost" db:"cost"`
	Date             *time.Time    `json:"date" db:"date"`
	StartTime        *time.Time    `json:"startTime" db:"start_time"`
	EndTime          *time.Time    `json:"endTime" db:"end_time"`
	Duration         *int64        `json:"duration" db:"duration"`
	Description      *string       `json:"description" db:"description"`
	Published        *time.Time    `json:"published" db:"published"`
	IsDeleted        *bool         `json:"isDeleted" db:"is_deleted"`
}

type JobTemplateFilter struct {
	UUID         *UUIDFilter         `json:"uuid" db:"uuid"`
	Name         *StringFilter       `json:"name" db:"name"`
	Created      *DateTimeFilter     `json:"created" db:"created"`
	Updated      *DateTimeFilter     `json:"updated" db:"updated"`
	Organization *OrganizationFilter `json:"organization" table:"organization" link:"uuid_organization"`
	Object       *OrganizationFilter `json:"object" table:"organization" link:"uuid_object"`
	Region       *RegionFilter       `json:"region" table:"region" link:"uuid_region"`
	Area         *AreaFilter         `json:"area" table:"area" link:"uuid_area"`
	City         *CityFilter         `json:"city" table:"cities" link:"uuid_city"`
	JobType      *JobTypeFilter      `json:"jobType" table:"job_types" link:"uuid_job_type"`
	Cost         *FloatFilter        `json:"cost" db:"cost"`
	Date         *DateFilter         `json:"date" db:"date"`
	StartTime    *TimeFilter         `json:"startTime" db:"start_time"`
	EndTime      *TimeFilter         `json:"endTime" db:"end_time"`
	Duration     *IntFilter          `json:"duration" db:"duration"`
	Description  *StringFilter       `json:"description" db:"description"`
	Published    *DateFilter         `json:"published" db:"published"`
	IsDeleted    *bool               `json:"isDeleted" db:"is_deleted"`
}

func (jt *JobTemplate) Mutation(ctx context.Context, db pglxqb.BaseRunner, app *app.App, rColumns interface{}, columns map[string]interface{}) (*pglx.Rows, *uuid.UUID, error) {
	update := false
	// если есть uuid значит манипулируем обектом
	if utils.CountFillFields(jt) == 1 && len(columns) == 0 && jt.UUID != nil {
		return nil, jt.UUID, nil
	}
	if jt.UUID != nil {
		jobTemplate, err := jt.GetByUUID(ctx, app, db, jt.UUID)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "Mutation").Err(err).Msg("Error get person")
			return nil, nil, gqlerror.Errorf("Error get person")
		}
		// востановим все ссылки
		utils.RestoreUUID(jt, jobTemplate)
		// востановим подчиненные структуры
		if err = jobTemplate.restoreStruct(ctx, app, db); err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "Mutation").Err(err).Msg("Error restore struct person")
			return nil, nil, gqlerror.Errorf("Error restore struct person")
		}
		update = true
	} else {
		// иначе создадим с нуля Обьект
		newUUID := uuid.New()
		jt.UUID = &newUUID
		columns["uuid"] = newUUID
	}
	parent := make(map[string]interface{})
	// дополним пропущеные поля, если они есть
	setColumns, err := SqlGenKeys(ctx, app, db, jt, columns, parent)
	if err != nil {
		app.Logger.Error().Str("module", "persons").Str("func", "Mutation").Err(err).Msg("Error generate map of keys")
		return nil, nil, err
	}
	// только одна колонка, и это uuid то удаляем обьект
	setColumns = utils.ClearSQLFields(jt, setColumns)
	if len(setColumns) > 0 {
		if update {
			// Обновляем иначе
			rows, err := pglxqb.Update("job_templates").
				SetMap(setColumns).
				Where("uuid = ?", jt.UUID).
				Suffix(utils.PrepareSuffix(rColumns)).
				RunWith(db).QueryX(ctx)
			if err != nil {
				app.Logger.Error().Str("module", "contact").Str("func", "Mutation").Err(err).Msg("Error update contact")
				return nil, nil, gqlerror.Errorf("Error update contact")
			}
			return rows, jt.UUID, nil
		} else {
			rows, err := pglxqb.Insert("job_templates").
				SetMap(setColumns).
				Suffix(utils.PrepareSuffix(rColumns)).
				RunWith(db).QueryX(ctx)
			if err != nil {
				app.Logger.Error().Str("module", "contact").Str("func", "Mutation").Err(err).Msg("Error insert contact")
				return nil, nil, gqlerror.Errorf("Error insert contact")
			}
			return rows, jt.UUID, nil
		}
	}
	return nil, jt.UUID, nil
}

func (jt *JobTemplate) ParseRows(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) ([]*JobTemplate, error) {
	var jobTemplates []*JobTemplate
	defer rows.Close()
	for rows.Next() {
		var jobTemplate JobTemplate
		err := rows.StructScan(&jobTemplate)
		if err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
		err = jobTemplate.parseRequestedFields(ctx, fields, app, db)
		if err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
		jobTemplates = append(jobTemplates, &jobTemplate)
	}
	return jobTemplates, nil
}

func (jt *JobTemplate) ParseRow(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) (*JobTemplate, error) {
	var err error
	var jobTemplate JobTemplate
	defer rows.Close()
	for rows.Next() {
		err = rows.StructScan(&jobTemplate)
		if err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
	}
	err = jobTemplate.parseRequestedFields(ctx, fields, app, db)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return &jobTemplate, nil
}

func (jt *JobTemplate) parseRequestedFields(ctx context.Context, fields []graphql.CollectedField, app *app.App, db pglxqb.BaseRunner) error {
	return parseRequestedFields(ctx, app, db, fields, jt)
}

func (jt *JobTemplate) restoreStruct(ctx context.Context, app *app.App, db pglxqb.BaseRunner) error {
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

func (jt *JobTemplate) GetByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID) (*JobTemplate, error) {
	rows, err := pglxqb.SelectAll().From("job_templates").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error get person from DB")
		return nil, gqlerror.Errorf("Error get person from DB")
	}
	var jobTemplate JobTemplate
	for rows.Next() {
		if err := rows.StructScan(&jobTemplate); err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct Person")
			return nil, gqlerror.Errorf("Error scan response to struct Person")
		}
	}
	return &jobTemplate, nil
}

func (jt *JobTemplate) GetParsedObjectByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID, column graphql.CollectedField) (*JobTemplate, error) {
	rows, err := pglxqb.SelectAll().From("job_templates").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return jt.ParseRow(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}

func (jt *JobTemplate) GetParsedObjectsByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid []*uuid.UUID, column graphql.CollectedField) ([]*JobTemplate, error) {
	rows, err := pglxqb.SelectAll().From("job_templates").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return jt.ParseRows(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}
