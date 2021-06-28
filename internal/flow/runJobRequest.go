package flow

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/internal/middleware"
	"github.com/sphera-erp/sphera/internal/models"
	"github.com/sphera-erp/sphera/internal/utils"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"time"
)

type JobStartRequest struct {
	UUIDPerson uuid.UUID
	UUIDJob    uuid.UUID
	Time       time.Time
	Lat        float64
	Lon        float64
	Code       string
}

var JobStartRequestCodes map[string]string

func init() {
	JobStartRequestCodes = make(map[string]string)
}

func (r *Resolver) RunJobRequest(ctx context.Context, job *models.Job, lat *float64, lon *float64) (*string, error) {
	if job == nil || job.UUID == nil {
		return nil, gqlerror.Errorf("Error not same request field")
	}
	userUUID, err := middleware.ExtractUserInTokenMetadata(ctx, r.env)
	if err != nil {
		return nil, gqlerror.Errorf("Error get user uuid from context")
	}
	var personUUID uuid.UUID
	err = pglxqb.Select("uuid").From("persons").Where(pglxqb.Eq{"uuid_user": userUUID}).RunWith(r.env.Cockroach).QueryRow(ctx).Scan(&personUUID)
	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error Select person from user ")
		return nil, gqlerror.Errorf("Error run transaction")
	}

	rows, err := pglxqb.SelectAll().
		From("jobs").
		Where(pglxqb.Eq{"uuid": job.UUID}).
		RunWith(r.env.Cockroach).QueryX(ctx)

	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "RunJob").Err(err).Msg("Error Select person from user ")
		return nil, gqlerror.Errorf("Error run transaction")
	}

	for rows.Next() {
		if err = rows.StructScan(&job); err != nil {
			r.env.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
	}

	var uuidOrganization uuid.UUID
	if err := pglxqb.Select("uuid_parent_organization").
		From("organizations").
		Where(pglxqb.Eq{"uuid": job.UUIDObject}).
		RunWith(r.env.Cockroach).QueryRow(ctx).Scan(&job.UUIDObject); err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "RunJob").Err(err).Msg("Error Select Org from user ")
		return nil, gqlerror.Errorf("Error run transaction")
	}

	var organizationStDistance *float64
	var organizationStTime *time.Duration
	if err := pglxqb.Select("st_distance, st_time").
		From("organizations").
		Where(pglxqb.Eq{"uuid": uuidOrganization}).
		RunWith(r.env.Cockroach).QueryRow(ctx).Scan(&organizationStDistance, &organizationStTime); err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "RunJob").Err(err).Msg("Error Select Org from user ")
		return nil, gqlerror.Errorf("Error run transaction")
	}

	if organizationStDistance != nil && *organizationStDistance != 0 {
		if lat != nil && lon != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "RunJobRequest").Err(err).Msg("Error lat and lon selfemployer is empty")
			return nil, gqlerror.Errorf("Error lat and lon selfemployer is empty")
		}
		query := `WITH o AS (Select a.lat lat, a.lon lon from organizations
					Left Join addresses a on a.uuid = organizations.uuid_address_fact
					where organizations.uuid = '02bbe651-7453-4bff-a847-469e9a1c11c4')
			Select st_distance(st_makepoint(o.lat, o.lon), st_makepoint(55.7713762,37.586412)) as distance from o ;`

		var distance *float64
		if err := r.env.Cockroach.QueryRow(ctx, query, job.UUIDObject).Scan(&distance); err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error get distance beetwen selfemployer and object")
			return nil, gqlerror.Errorf("Error get distance beetwen selfemployer and object")
		}
		if *distance > *organizationStDistance {
			r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error selfemployed being far from the object")
			return nil, gqlerror.Errorf("Error Self-employed being far from the object")
		}
	}

	if organizationStTime != nil {
		if job.Date.Add(time.Duration(job.StartTime.Hour())).Add(time.Duration(job.StartTime.Minute())).Add(time.Duration(job.StartTime.Second())).After(time.Now().Add(*organizationStTime)) {
			r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error selfemployed being far from the object")
			return nil, gqlerror.Errorf("Error Self-employed being far from the object")
		}
	}

	code := fmt.Sprintf("%06d", utils.NewCryptoRand())
	result := new(JobStartRequest)
	result.UUIDPerson = personUUID
	result.UUIDJob = *job.UUID
	result.Lat = *lat
	result.Lon = *lon
	result.Time = time.Now().UTC()
	result.Code = code
	stringResult, err := json.Marshal(result)
	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error create json req to start job")
		return nil, gqlerror.Errorf("Error create json req to start job")
	}
	JobStartRequestCodes[code] = string(stringResult)
	return &code, nil
}
