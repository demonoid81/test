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
	"github.com/sphera-erp/sphera/internal/middleware"
	"github.com/sphera-erp/sphera/internal/utils"
	"github.com/sphera-erp/sphera/pkg/pglx"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type Job struct {
	UUID              *uuid.UUID    `json:"uuid" db:"uuid"`
	Created           *time.Time    `json:"created" db:"created"`
	Updated           *time.Time    `json:"updated" db:"updated"`
	Name              *string       `json:"name" db:"name"`
	Date              *time.Time    `json:"date" db:"date"`
	StartTime         *time.Time    `json:"startTime" db:"start_time"`
	EndTime           *time.Time    `json:"endTime" db:"end_time"`
	Duration          *int64        `json:"duration" db:"duration"`
	Cost              *float64      `json:"cost" db:"cost"`
	UUIDObject        *uuid.UUID    `db:"uuid_object"`
	Object            *Organization `json:"object" relay:"uuid_object" link:"UUIDObject"`
	UUIDJobTemplate   *uuid.UUID    `db:"uuid_job_template"`
	JobTemplate       *JobTemplate  `json:"jobTemplate" relay:"uuid_job_template" link:"UUIDJobTemplate"`
	UUIDJobType       *uuid.UUID    `db:"uuid_job_type"`
	JobType           *JobType      `json:"jobType" relay:"uuid_job_type" link:"UUIDJobType"`
	Description       *string       `json:"description" db:"description"`
	IsHot             *bool         `json:"isHot" db:"is_hot"`
	UUIDCandidates    []*uuid.UUID  `db:"uuid_candidates"`
	Candidates        []*Candidate  `json:"candidates" relay:"uuid_candidates" link:"UUIDCandidates"`
	UUIDExecutor      *uuid.UUID    `db:"uuid_executor"`
	Executor          *Person       `json:"executor" relay:"uuid_executor" link:"UUIDExecutor"`
	UUIDStatuses      []*uuid.UUID  `db:"uuid_statuses"`
	Statuses          []*Status     `json:"statuses" relay:"uuid_statuses" link:"UUIDStatuses"`
	IsDeleted         *bool         `json:"isDeleted" db:"is_deleted"`
	Status            *JobStatus    `json:"status" db:"status"`
	Published         *time.Time    `json:"published" db:"published"`
	Rating            *float64      `json:"rating" db:"rating"`
	RatingDescription *string       `json:"ratingDescription" db:"rating_description"`
}

type JobFilter struct {
	UUID              *UUIDFilter         `json:"uuid" db:"uuid"`
	Created           *DateTimeFilter     `json:"created" db:"created"`
	Updated           *DateTimeFilter     `json:"updated" db:"updated"`
	Name              *StringFilter       `json:"name" db:"name"`
	Date              *DateFilter         `json:"date" db:"date"`
	StartTime         *TimeFilter         `json:"startTime" db:"start_time"`
	EndTime           *TimeFilter         `json:"endTime" db:"end_time"`
	Duration          *IntFilter          `json:"duration" db:"duration"`
	Cost              *FloatFilter        `json:"cost" db:"cost"`
	Object            *OrganizationFilter `json:"object" table:"organizations" link:"uuid_object"`
	JobTemplate       *JobTemplateFilter  `json:"jobTemplate" table:"job_templates" link:"uuid_job_template"`
	JobType           *JobTypeFilter      `json:"jobType" table:"job_types" link:"uuid_job_type"`
	Description       *StringFilter       `json:"description" db:"description"`
	IsHot             *bool               `json:"isHot" db:"is_hot"`
	Executor          *PersonFilter       `json:"executor" table:"persons" link:"uuid_executor"`
	IsDeleted         *bool               `json:"isDeleted" db:"is_deleted"`
	Status            *StringFilter       `json:"status" db:"status"`
	Published         *DateTimeFilter     `json:"published" db:"published"`
	Rating            *FloatFilter        `json:"rating" db:"rating"`
	RatingDescription *StringFilter       `json:"ratingDescription" db:"rating_description"`
	And               []JobFilter         `json:"and"`
	Or                []JobFilter         `json:"or"`
	Not               *JobFilter          `json:"not"`
}

type JobSort struct {
	Field *JobSortableField `json:"field"`
	Order *SortOrder        `json:"order"`
}

type JobSortableField string

const (
	JobSortableFieldUUID        JobSortableField = "uuid"
	JobSortableFieldCreated     JobSortableField = "created"
	JobSortableFieldUpdated     JobSortableField = "updated"
	JobSortableFieldDate        JobSortableField = "date"
	JobSortableFieldStartTime   JobSortableField = "startTime"
	JobSortableFieldEndTime     JobSortableField = "endTime"
	JobSortableFieldDuration    JobSortableField = "duration"
	JobSortableFieldDescription JobSortableField = "description"
)

func (e JobSortableField) IsValid() bool {
	switch e {
	case JobSortableFieldUUID,
		JobSortableFieldCreated,
		JobSortableFieldUpdated,
		JobSortableFieldDate,
		JobSortableFieldStartTime,
		JobSortableFieldEndTime,
		JobSortableFieldDescription,
		JobSortableFieldDuration:
		return true
	}
	return false
}

func (e JobSortableField) String() string {
	return string(e)
}

func (e *JobSortableField) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = JobSortableField(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid JobSortableField", str)
	}
	return nil
}

func (e JobSortableField) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type JobStatus string

const (
	JobStatusCreated  JobStatus = "created"
	JobStatusDraft    JobStatus = "draft"
	JobStatusPublish  JobStatus = "publish"
	JobStatusFull     JobStatus = "full"
	JobStatusReady    JobStatus = "ready"
	JobStatusOnObject JobStatus = "onObject"
	JobStatusStart    JobStatus = "start"
	JobStatusEnd      JobStatus = "end"
	JobStatusCancel   JobStatus = "cancel"
	JobStatusReject   JobStatus = "reject"
	JobStatusDispute  JobStatus = "dispute"
	JobStatusRefuse   JobStatus = "refuse"
)

func (e JobStatus) IsValid() bool {
	switch e {
	case JobStatusCreated, JobStatusDraft, JobStatusPublish, JobStatusFull,
		JobStatusReady, JobStatusOnObject, JobStatusStart, JobStatusEnd,
		JobStatusCancel, JobStatusReject, JobStatusDispute, JobStatusRefuse:
		return true
	}
	return false
}

func (e JobStatus) String() string {
	return string(e)
}

func (e *JobStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = JobStatus(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid JobStatus", str)
	}
	return nil
}

func (e JobStatus) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

func (j *Job) Mutation(ctx context.Context, db pglxqb.BaseRunner, app *app.App, rColumns interface{}, columns map[string]interface{}) (*pglx.Rows, *uuid.UUID, error) {
	logger := app.Logger.Error().Str("package", "models").Str("model", "job").Str("func", "Mutation")
	// update := false
	// если есть uuid значит манипулируем объектом
	if j.UUID != nil {
		if utils.CountFillFields(j) == 1 && len(columns) == 0 {
			return nil, j.UUID, nil
		}
		job, err := j.GetByUUID(ctx, app, db, j.UUID)
		if err != nil {
			logger.Err(err).Msg("Error get job")
			return nil, nil, gqlerror.Errorf("Error get job")
		}
		switch *job.Status {
		case JobStatusFull, JobStatusReady, JobStatusOnObject, JobStatusStart, JobStatusEnd, JobStatusCancel, JobStatusReject, JobStatusDispute:
			logger.Msg("Error job not editable")
			return nil, nil, gqlerror.Errorf("Error job not editable")
		}
		// восстановим все ссылки
		utils.RestoreUUID(j, job)
		// восстановим подчиненные структуры
		if err = job.restoreStruct(ctx, app, db); err != nil {
			logger.Err(err).Msg("Error restore struct job")
			return nil, nil, gqlerror.Errorf("Error restore struct job")
		}
		// update = true
	} else {
		// иначе создадим с нуля Объект
		newUUID := uuid.New()
		j.UUID = &newUUID
		_, err := pglxqb.Insert("jobs").
			Columns("uuid").
			Values(newUUID).
			RunWith(db).Exec(ctx)
		if err != nil {
			app.Logger.Error().Str("module", "job").Str("func", "Mutation").Err(err).Msg("Error insert job")
			return nil, nil, gqlerror.Errorf("Error insert job")
		}
		// update = false
	}
	// дополним пропущенные поля, если они есть
	parent := make(map[string]interface{})
	setColumns, err := SqlGenKeys(ctx, app, db, j, columns, parent)
	if err != nil {
		logger.Err(err).Msg("Error generate map of keys")
		return nil, nil, err
	}
	// только одна колонка, и это uuid то удаляем объект
	setColumns = utils.ClearSQLFields(j, setColumns)

	userUUID, err := middleware.ExtractUserInTokenMetadata(ctx, app)
	if err != nil {
		logger.Err(err).Msg("get user uuid from context")
		return nil, nil, gqlerror.Errorf("Error get user uuid from context")
	}

	// достанем персону из пользователя
	var personUUID uuid.UUID
	err = pglxqb.Select("uuid").From("persons").Where(pglxqb.Eq{"uuid_user": userUUID}).RunWith(db).QueryRow(ctx).Scan(&personUUID)
	if err != nil {
		logger.Err(err).Msg("Error Select person from user ")
		return nil, nil, gqlerror.Errorf("Error Select person from user")
	}

	if len(setColumns) > 0 {
		// if update {
		// delete(setColumns, "status")
		var statusesUUID []uuid.UUID
		if err = pglxqb.Select("uuid_statuses").
			From("jobs").Where(pglxqb.Eq{"uuid": j.UUID}).
			RunWith(db).QueryRow(ctx).Scan(&statusesUUID); err != nil {
			logger.Err(err).Msg("Error select status from jobs")
			return nil, nil, gqlerror.Errorf("Error select status from jobs")
		}

		statusUUID := uuid.New()
		if _, err := pglxqb.Insert("statuses").
			Columns("uuid", "uuid_job", "status").
			Values(statusUUID, j.UUID, setColumns["status"]).
			RunWith(db).Exec(ctx); err != nil {
			logger.Err(err).Msg("Error set job status")
			return nil, nil, gqlerror.Errorf("Error Set status")
		}

		// Обновляем иначе
		rows, err := pglxqb.Update("jobs").
			SetMap(setColumns).
			Set("uuid_statuses", append(statusesUUID, statusUUID)).
			Where("uuid = ?", j.UUID).
			Suffix(utils.PrepareSuffix(rColumns)).
			RunWith(db).
			QueryX(ctx)
		if err != nil {
			logger.Err(err).Msg("Error update contact")
			return nil, nil, gqlerror.Errorf("Error update contact")
		}
		return rows, j.UUID, nil
		// } else {
		// 	statusUUID := uuid.New()
		// 	if _, err = pglxqb.Insert("statuses").
		// 		Columns("uuid", "uuid_job", "status").
		// 		Values(statusUUID, j.UUID, JobStatusCreated).
		// 		RunWith(db).Exec(ctx); err != nil {
		// 		logger.Err(err).Msg("Error set job status")
		// 		return nil, nil, gqlerror.Errorf("Error Set status")
		// 	}
		// 	rows, err := pglxqb.Insert("jobs").
		// 		SetMap(setColumns).
		// 		Columns("uuid_statuses").Values([]uuid.UUID{statusUUID}).
		// 		Suffix(utils.PrepareSuffix(rColumns)).
		// 		RunWith(db).
		// 		QueryX(ctx)
		// 	if err != nil {
		// 		logger.Err(err).Msg("Error insert contact")
		// 		return nil, nil, gqlerror.Errorf("Error insert contact")
		// 	}
		// 	return rows, j.UUID, nil
		// }

	}
	return nil, j.UUID, nil
}

func (j *Job) ParseRows(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) ([]*Job, error) {
	var jobs []*Job
	defer rows.Close()
	for rows.Next() {
		var job Job
		err := rows.StructScan(&job)
		if err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}

		jobs = append(jobs, &job)
	}
	for _, job := range jobs {
		if err := job.ParseRequestedFields(ctx, fields, app, db); err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
	}
	return jobs, nil
}

func (j *Job) ParseRow(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) (*Job, error) {
	var err error
	var job Job
	for rows.Next() {
		err = rows.StructScan(&job)
		if err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
	}
	err = job.ParseRequestedFields(ctx, fields, app, db)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return &job, nil
}

func (j *Job) ParseRequestedFields(ctx context.Context, fields []graphql.CollectedField, app *app.App, db pglxqb.BaseRunner) error {
	return parseRequestedFields(ctx, app, db, fields, j)
}

func (j *Job) restoreStruct(ctx context.Context, app *app.App, db pglxqb.BaseRunner) error {
	v := reflect.ValueOf(j)
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

func (j *Job) GetByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID) (*Job, error) {
	rows, err := pglxqb.SelectAll().From("jobs").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error get person from DB")
		return nil, gqlerror.Errorf("Error get person from DB")
	}
	var job Job
	defer rows.Close()
	for rows.Next() {
		if err := rows.StructScan(&job); err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct Person")
			return nil, gqlerror.Errorf("Error scan response to struct Person")
		}
	}
	return &job, nil
}

func (j *Job) GetParsedObjectByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID, column graphql.CollectedField) (*Job, error) {
	rows, err := pglxqb.SelectAll().From("jobs").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return j.ParseRow(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}

func (j *Job) GetParsedObjectsByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid []*uuid.UUID, column graphql.CollectedField) ([]*Job, error) {
	rows, err := pglxqb.SelectAll().From("jobs").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return j.ParseRows(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}
