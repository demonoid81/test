package jobs

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/internal/middleware"
	"github.com/sphera-erp/sphera/internal/models"
	"github.com/sphera-erp/sphera/internal/utils"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func (r *Resolver) ActiveJobs(ctx context.Context, job *models.Job, filter *models.JobFilter, sort []models.JobSort, offset *int, limit *int) ([]*models.Job, error) {
	var err error

	userType, err := middleware.ExtractUserTypeInTokenMetadata(ctx, r.env)
	if err != nil {
		r.env.Logger.Error().Str("module", "jobs").Str("func", "ActiveJobs").Err(err).Msg("Error get user type")
		return nil, gqlerror.Errorf("Error get user type")
	}

	if userType == models.SystemUser.String() {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBooks").Err(err).Msg("Error method allowed only SelfEmployers")
		return nil, gqlerror.Errorf("Error method allowed only SelfEmployers")
	}

	table := "jobs"
	sql := pglxqb.Select(fmt.Sprintf("%s.*", table)).From(table)
	if filter != nil {
		sql = utils.ReflectFilter(table, sql, filter)
	} else if job != nil {
		var result map[string]interface{}
		result, sql, err = models.SqlGenSelectKeys(job, sql, table, 1)
		if err != nil {
			r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBooks").Err(err).Msg("Error generate select relations")
			return nil, gqlerror.Errorf("Error generate select relations")
		}
		if len(result) > 0 {
			sql = sql.Where(pglxqb.Eq(result))
		}
	}

	for _, sortItem := range sort {
		sql = sql.OrderBy(fmt.Sprintf("%s.%s %s", table, sortItem.Field, sortItem.Order))
	}

	if limit != nil {
		sql = sql.Limit(uint64(*limit))
	}
	if offset != nil {
		sql = sql.Offset(uint64(*offset))
	}

	userUUID, err := middleware.ExtractUserInTokenMetadata(ctx, r.env)
	if err != nil {
		r.env.Logger.Error().Str("module", "jobs").Str("func", "jobs").Err(err).Msg("Error get user uuid from context")
		return nil, gqlerror.Errorf("Error get user uuid from context")
	}
	var personUUID *uuid.UUID
	var reward *float64
	if err = pglxqb.Select("uuid", "reward").
		From("persons").
		Where(pglxqb.Eq{"uuid_user": userUUID}).
		RunWith(r.env.Cockroach).QueryRow(ctx).
		Scan(&personUUID, &reward); err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error Select person from user ")
		return nil, gqlerror.Errorf("Error Select person from user")
	}

	sql = sql.Where(pglxqb.Eq{"jobs.status": models.JobStatusPublish})

	var excludeUUID []uuid.UUID
	jRows, err := pglxqb.Select("uuid_job").
		From("candidates").
		Where(pglxqb.Eq{"uuid_person": personUUID}).
		RunWith(r.env.Cockroach).Query(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error Select person from user ")
		return nil, gqlerror.Errorf("Error Select person from user")
	}
	defer jRows.Close()

	for jRows.Next() {
		var jUUID uuid.UUID
		if err := jRows.Scan(&jUUID); err != nil {
			r.env.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct PersonCourse")
			return nil, gqlerror.Errorf("Error scan response to struct PersonCourse")
		}
		excludeUUID = append(excludeUUID, jUUID)
	}

	if len(excludeUUID) > 0 {
		sql = sql.Where(pglxqb.NotEq{"jobs.uuid": excludeUUID})
	}
	if personUUID != nil {
		sql = sql.Where(pglxqb.Or{pglxqb.Expr("jobs.uuid_executor not in(?)", personUUID), pglxqb.Expr("jobs.uuid_executor is null")})
	}

	if reward != nil {
		// уберем все работы где есть главный кадидат
		var excludeUUID []uuid.UUID
		jRows, err := pglxqb.Select("uuid_job").
			From("candidates").
			Where(pglxqb.Eq{"candidate_tag": models.Primary.String()}).
			RunWith(r.env.Cockroach).Query(ctx)
		if err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error Select person from user ")
			return nil, gqlerror.Errorf("Error Select person from user")
		}
		defer jRows.Close()

		for jRows.Next() {
			var jUUID uuid.UUID
			if err := jRows.Scan(&jUUID); err != nil {
				r.env.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct PersonCourse")
				return nil, gqlerror.Errorf("Error scan response to struct PersonCourse")
			}
			excludeUUID = append(excludeUUID, jUUID)
		}

		if len(excludeUUID) > 0 {
			sql = sql.Where(pglxqb.NotEq{"jobs.uuid": excludeUUID})
		}
	}

	sql = sql.Where("(jobs.date + jobs.start_time) > current_timestamp")

	rows, err := sql.RunWith(r.env.Cockroach).QueryX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBooks").Err(err).Msg("Error select medicalBooks")
		return nil, gqlerror.Errorf("Error select medicalBooks")
	}
	return job.ParseRows(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, r.env.Cockroach)
}
