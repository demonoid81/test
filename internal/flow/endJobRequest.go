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

func (r *Resolver) EndJobRequest(ctx context.Context, job *models.Job, lat *float64, lon *float64) (*string, error) {
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
	code := fmt.Sprintf("%06d", utils.NewCryptoRand())
	result := new(JobStartRequest)
	result.UUIDPerson = personUUID
	result.UUIDJob = *job.UUID
	//result.Lat = *lat
	//result.Lon = *lon
	result.Time = time.Now().UTC()
	result.Code = code
	stringResult, err := json.Marshal(result)
	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error create json req to end job")
		return nil, gqlerror.Errorf("Error create json req to end job")
	}
	JobStartRequestCodes[code] = string(stringResult)
	return &code, nil
}
