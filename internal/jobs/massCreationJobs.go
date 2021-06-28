package jobs

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/internal/middleware"
	"github.com/sphera-erp/sphera/internal/models"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func (r *Resolver) MassCreationJobs(ctx context.Context, jobTemplate models.JobTemplate, objects []models.Organization, dates []*time.Time) (bool, error) {
	logger := r.env.Logger.Error().Str("package", "models").Str("model", "job").Str("func", "Mutation")
	tx, err := r.env.Cockroach.BeginX(ctx)
	if err != nil {
		logger.Err(err).Msg("Error run transaction")
		return false, gqlerror.Errorf("Error run transaction")
	}
	defer tx.Rollback(ctx)
	userUUID, err := middleware.ExtractUserInTokenMetadata(ctx, r.env)
	if err != nil {
		return false, gqlerror.Errorf("Error get user uuid from context")
	}

	var personUUID uuid.UUID
	if err = pglxqb.Select("uuid").
		From("persons").
		Where(pglxqb.Eq{"uuid_user": userUUID}).
		RunWith(tx).QueryRow(ctx).Scan(&personUUID); err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "AddMsg").Err(err).Msg("Error Select person from user ")
		return false, gqlerror.Errorf("Error run transaction")
	}

	var template models.JobTemplate

	rows, err := pglxqb.SelectAll().
		From("job_templates").
		Where(pglxqb.Eq{"uuid": jobTemplate.UUID}).
		RunWith(tx).QueryX(ctx)
	if err != nil {
		logger.Err(err).Msg("Error select JobTemplate")
		return false, gqlerror.Errorf("Error select JobTemplate")
	}
	for rows.Next() {
		if err := rows.StructScan(&template); err != nil {
			logger.Err(err).Msg("Error scan response to struct JobTemplate")
			return false, gqlerror.Errorf("Error scan response to struct JobTemplate")
		}
	}
	setColumns := structToMap(template)

	j := models.Job{}
	setColumns["uuid_job_template"] = jobTemplate.UUID
	if len(dates) == 0 {
		if setColumns["date"] == nil {
			logger.Err(err).Msg("Error in job date win not nil")
			return false, gqlerror.Errorf("Error in job date win not nil")
		}
	}
	if setColumns["start_time"] == nil || setColumns["duration"] == nil {
		logger.Err(err).Msg("Error empty values in job template")
		return false, gqlerror.Errorf("Error empty values in job template")
	}
	delete(setColumns, "updated")
	delete(setColumns, "created")
	delete(setColumns, "uuid")
	delete(setColumns, "uuid_region")
	delete(setColumns, "uuid_area")
	delete(setColumns, "uuid_city")
	delete(setColumns, "uuid_organization")
	fmt.Println(setColumns)
	for _, object := range objects {
		if object.UUID != nil {
			setColumns["uuid_object"] = object.UUID
			setColumns["status"] = "publish"
			if len(dates) > 0 {
				for _, date := range dates {
					setColumns["date"] = date
					jobUUID := uuid.New()
					if _, err := pglxqb.Insert("jobs").
						Columns("uuid").Values(jobUUID).
						RunWith(tx).Exec(ctx); err != nil {
						logger.Err(err).Msg("Error insert job")
						return false, gqlerror.Errorf("Error insert job")
					}
					statusUUID := uuid.New()
					if _, err = pglxqb.Insert("statuses").
						Columns("uuid", "uuid_job", "status", "uuid_person").
						Values(statusUUID, j.UUID, models.JobStatusCreated, personUUID).
						RunWith(tx).Exec(ctx); err != nil {
						logger.Err(err).Msg("Error set job status")
						return false, gqlerror.Errorf("Error Set status")
					}
					if _, err := pglxqb.Update("jobs").
						SetMap(setColumns).
						Set("uuid_statuses", []uuid.UUID{statusUUID}).
						Where(pglxqb.Eq{"uuid": jobUUID}).
						RunWith(tx).Exec(ctx); err != nil {
						logger.Err(err).Msg("Error update job")
						return false, gqlerror.Errorf("Error update job")
					}
				}
			} else {
				jobUUID := uuid.New()
				if _, err := pglxqb.Insert("jobs").
					Columns("uuid").Values(jobUUID).
					RunWith(tx).Exec(ctx); err != nil {
					logger.Err(err).Msg("Error insert job")
					return false, gqlerror.Errorf("Error insert job")
				}
				statusUUID := uuid.New()
				if _, err = pglxqb.Insert("statuses").
					Columns("uuid", "uuid_job", "status", "uuid_person").
					Values(statusUUID, j.UUID, models.JobStatusCreated, personUUID).
					RunWith(tx).Exec(ctx); err != nil {
					logger.Err(err).Msg("Error set job status")
					return false, gqlerror.Errorf("Error Set status")
				}
				if _, err := pglxqb.Update("jobs").
					SetMap(setColumns).
					Set("uuid_statuses", []uuid.UUID{statusUUID}).
					Where(pglxqb.Eq{"uuid": jobUUID}).
					RunWith(tx).Exec(ctx); err != nil {
					logger.Err(err).Msg("Error update job")
					return false, gqlerror.Errorf("Error update job")
				}
			}
		}
	}
	err = tx.Commit(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error commit transaction")
		return false, gqlerror.Errorf("Error commit transaction")
	}
	return true, nil
}

func structToMap(src interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	vField := reflect.ValueOf(src)
	for i := 0; i < vField.NumField(); i++ {
		db := vField.Type().Field(i).Tag.Get("db")
		if db != "" {
			fmt.Println(db, vField.Field(i).Interface())
			if db == "uuid" {
				result["uuid_job_template"] = vField.Field(i).Interface()
			} else {
				result[db] = vField.Field(i).Interface()
			}
		}
	}
	return result
}
