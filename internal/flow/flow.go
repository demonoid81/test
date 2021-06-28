package flow

import (
	"context"
	"github.com/sphera-erp/sphera/app"
	"github.com/sphera-erp/sphera/internal/models"
)

type Resolver struct {
	env *app.App
	Resolvers
}

type Resolvers interface {
	AgreeToJob(ctx context.Context, job *models.Job, user *models.User) (*models.InfoAboutJob, error)
	PublishJob(ctx context.Context, job *models.Job) (bool, error)
	RunJobRequest(ctx context.Context, job *models.Job, lat *float64, lon *float64) (*string, error)
	RunJob(ctx context.Context, code *string) (bool, error)
	EndJobRequest(ctx context.Context, job *models.Job, lat *float64, lon *float64) (*string, error)
	EndJob(ctx context.Context, code *string, rating *float64, ratingDescription *string) (*models.PersonRating, error)
	RejectPerson(ctx context.Context, job *models.Job, person *models.Person, reason string) (bool, error)
	RefuseJob(ctx context.Context, job *models.Job, reason string) (bool, error)
	OnPlace(ctx context.Context, job *models.Job, lat *float64, lon *float64) (bool, error)
	Check(ctx context.Context, job *models.Job, lat *float64, lon *float64, user *models.User) (bool, error)
	ConflictOnJob(ctx context.Context, job *models.Job, reason string) (bool, error)
	CancelJob(ctx context.Context, job *models.Job, reason string) (bool, error)
	AddMsg(ctx context.Context, job *models.Job, description string, content []*models.Content) (bool, error)
	GetMsgStats(ctx context.Context) ([]*models.MsgStat, error)
	MsgStatSub(ctx context.Context) (<-chan *models.MsgStat, error)
	ReadMsg(ctx context.Context, job models.Job) (bool, error)
	UserMsg(ctx context.Context, status *models.Status, offset *int, limit *int) ([]*models.Status, error)
}

func NewJobFlowResolvers(app *app.App) (*Resolver, error) {
	return &Resolver{
		env: app,
	}, nil
}
