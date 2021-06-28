package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/sphera-erp/sphera/internal/models"
)

func (r *mutationResolver) JobFlowMutation(ctx context.Context, jobFlow *models.JobFlow) (*models.JobFlow, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) PublishJob(ctx context.Context, job *models.Job) (bool, error) {
	return r.jobFlow.PublishJob(ctx, job)
}

func (r *mutationResolver) AgreeToJob(ctx context.Context, job *models.Job, user *models.User) (*models.InfoAboutJob, error) {
	return r.jobFlow.AgreeToJob(ctx, job, user)
}

func (r *mutationResolver) RefuseJob(ctx context.Context, job *models.Job, reason string) (bool, error) {
	return r.jobFlow.RefuseJob(ctx, job, reason)
}

func (r *mutationResolver) Check(ctx context.Context, job *models.Job, lat *float64, lon *float64, user *models.User) (bool, error) {
	return r.jobFlow.Check(ctx, job, lat, lon, user)
}

func (r *mutationResolver) OnPlace(ctx context.Context, job *models.Job, lat *float64, lon *float64) (bool, error) {
	return r.jobFlow.OnPlace(ctx, job, lat, lon)
}

func (r *mutationResolver) ConflictOnJob(ctx context.Context, job *models.Job, reason string) (bool, error) {
	return r.jobFlow.ConflictOnJob(ctx, job, reason)
}

func (r *mutationResolver) ChangeStatusJob(ctx context.Context, job *models.Job, status *models.Status) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CancelJob(ctx context.Context, job *models.Job, reason string) (bool, error) {
	return r.jobFlow.CancelJob(ctx, job, reason)
}

func (r *mutationResolver) RejectPerson(ctx context.Context, job *models.Job, person *models.Person, reason string) (bool, error) {
	return r.jobFlow.RejectPerson(ctx, job, person, reason)
}

func (r *mutationResolver) BrokenJob(ctx context.Context, job *models.Job, reason string) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CloseJob(ctx context.Context, job *models.Job, percentagePayment *int) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) RunJob(ctx context.Context, code *string) (bool, error) {
	return r.jobFlow.RunJob(ctx, code)
}

func (r *mutationResolver) EndJob(ctx context.Context, code *string, rating *float64, ratingDescription *string) (*models.PersonRating, error) {
	return r.jobFlow.EndJob(ctx, code)
}

func (r *mutationResolver) AddMsg(ctx context.Context, job *models.Job, description string, content []*models.Content) (bool, error) {
	return r.jobFlow.AddMsg(ctx, job, description, content)
}

func (r *mutationResolver) ReadMsg(ctx context.Context, job models.Job) (bool, error) {
	return r.jobFlow.ReadMsg(ctx, job)
}

func (r *queryResolver) JobFlow(ctx context.Context, jobFlow *models.JobFlow) (*models.JobFlow, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) RunJobRequest(ctx context.Context, job *models.Job, lat *float64, lon *float64) (*string, error) {
	return r.jobFlow.RunJobRequest(ctx, job, lat, lon)
}

func (r *queryResolver) EndJobRequest(ctx context.Context, job *models.Job, lat *float64, lon *float64) (*string, error) {
	return r.jobFlow.EndJobRequest(ctx, job, lat, lon)
}

func (r *queryResolver) SignToHotJob(ctx context.Context) (*bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) GetMsgStats(ctx context.Context) ([]*models.MsgStat, error) {
	return r.jobFlow.GetMsgStats(ctx)
}

func (r *queryResolver) UserMsg(ctx context.Context, status *models.Status, offset *int, limit *int) ([]*models.Status, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *subscriptionResolver) MsgStatSub(ctx context.Context) (<-chan *models.MsgStat, error) {
	return r.jobFlow.MsgStatSub(ctx)
}
