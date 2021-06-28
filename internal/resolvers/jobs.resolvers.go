package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/internal/models"
)

func (r *mutationResolver) CourseMutate(ctx context.Context, course *models.Course) (*models.Course, error) {
	return r.jobs.CourseMutate(ctx, course)
}

func (r *mutationResolver) LocalityJobCostMutate(ctx context.Context, localityJobCost *models.LocalityJobCost) (*models.LocalityJobCost, error) {
	return r.jobs.LocalityJobCostMutate(ctx, localityJobCost)
}

func (r *mutationResolver) JobTypeMutate(ctx context.Context, jobType *models.JobType) (*models.JobType, error) {
	return r.jobs.JobTypeMutate(ctx, jobType)
}

func (r *mutationResolver) JobTemplateMutate(ctx context.Context, jobTemplate *models.JobTemplate) (*models.JobTemplate, error) {
	return r.jobs.JobTemplateMutate(ctx, jobTemplate)
}

func (r *mutationResolver) JobMutate(ctx context.Context, job *models.Job) (*models.Job, error) {
	return r.jobs.JobMutate(ctx, job)
}

func (r *mutationResolver) CandidateMutate(ctx context.Context, candidate *models.Candidate) (*models.Candidate, error) {
	return r.jobs.CandidateMutate(ctx, candidate)
}

func (r *mutationResolver) StatusMutate(ctx context.Context, status *models.Status) (*models.Status, error) {
	return r.jobs.StatusMutate(ctx, status)
}

func (r *mutationResolver) TagMutate(ctx context.Context, tag *models.Tag) (*models.Tag, error) {
	return r.jobs.TagMutate(ctx, tag)
}

func (r *mutationResolver) MassCreationJobs(ctx context.Context, jobTemplate models.JobTemplate, objects []models.Organization, dates []*time.Time) (bool, error) {
	return r.jobs.MassCreationJobs(ctx, jobTemplate, objects, dates)
}

func (r *mutationResolver) SetJobRating(ctx context.Context, job uuid.UUID, rating float64, description *string) (bool, error) {
	return r.jobs.SetJobRating(ctx, job, rating, description)
}

func (r *queryResolver) GetTypeJobIcons(ctx context.Context) ([]string, error) {
	return []string{"cleaner", "cashier", "loader", "merchandiser"}, nil
}

func (r *queryResolver) Course(ctx context.Context, course *models.Course) (*models.Course, error) {
	return r.jobs.Course(ctx, course)
}

func (r *queryResolver) Courses(ctx context.Context, course *models.Course, offset *int, limit *int) ([]*models.Course, error) {
	return r.jobs.Courses(ctx, course, offset, limit)
}

func (r *queryResolver) LocalityJobCost(ctx context.Context, localityJobCost *models.LocalityJobCost) (*models.LocalityJobCost, error) {
	return r.jobs.LocalityJobCost(ctx, localityJobCost)
}

func (r *queryResolver) LocalityJobCosts(ctx context.Context, localityJobCost *models.LocalityJobCost, offset *int, limit *int) ([]*models.LocalityJobCost, error) {
	return r.jobs.LocalityJobCosts(ctx, localityJobCost, offset, limit)
}

func (r *queryResolver) JobType(ctx context.Context, jobType *models.JobType) (*models.JobType, error) {
	return r.jobs.JobType(ctx, jobType)
}

func (r *queryResolver) JobTypes(ctx context.Context, jobType *models.JobType, offset *int, limit *int) ([]*models.JobType, error) {
	return r.jobs.JobTypes(ctx, jobType, offset, limit)
}

func (r *queryResolver) JobTemplate(ctx context.Context, jobTemplate *models.JobTemplate) (*models.JobTemplate, error) {
	return r.jobs.JobTemplate(ctx, jobTemplate)
}

func (r *queryResolver) JobTemplates(ctx context.Context, jobTemplate *models.JobTemplate, offset *int, limit *int) ([]*models.JobTemplate, error) {
	return r.jobs.JobTemplates(ctx, jobTemplate, offset, limit)
}

func (r *queryResolver) Job(ctx context.Context, job *models.Job) (*models.Job, error) {
	return r.jobs.Job(ctx, job)
}

func (r *queryResolver) Jobs(ctx context.Context, job *models.Job, filter *models.JobFilter, sort []models.JobSort, offset *int, limit *int) ([]*models.Job, error) {
	return r.jobs.Jobs(ctx, job, filter, sort, offset, limit)
}

func (r *queryResolver) ActiveJobs(ctx context.Context, job *models.Job, filter *models.JobFilter, sort []models.JobSort, offset *int, limit *int) ([]*models.Job, error) {
	return r.jobs.ActiveJobs(ctx, job, filter, sort, offset, limit)
}

func (r *queryResolver) Candidate(ctx context.Context, candidate *models.Candidate) (*models.Candidate, error) {
	return r.jobs.Candidate(ctx, candidate)
}

func (r *queryResolver) Candidates(ctx context.Context, candidate *models.Candidate, offset *int, limit *int) ([]*models.Candidate, error) {
	return r.jobs.Candidates(ctx, candidate, offset, limit)
}

func (r *queryResolver) Status(ctx context.Context, status *models.Status) (*models.Status, error) {
	return r.jobs.Status(ctx, status)
}

func (r *queryResolver) Statuses(ctx context.Context, status *models.Status, offset *int, limit *int) ([]*models.Status, error) {
	return r.jobs.Statuses(ctx, status, offset, limit)
}

func (r *queryResolver) Tag(ctx context.Context, tag *models.Tag) (*models.Tag, error) {
	return r.jobs.Tag(ctx, tag)
}

func (r *queryResolver) Tags(ctx context.Context, tag *models.Tag, offset *int, limit *int) ([]*models.Tag, error) {
	return r.jobs.Tags(ctx, tag, offset, limit)
}

func (r *subscriptionResolver) JobSub(ctx context.Context) (<-chan *models.Job, error) {
	return r.jobs.JobSub(ctx)
}

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *subscriptionResolver) JobsSub(ctx context.Context, job *models.Job, filter *models.JobFilter, sort []models.JobSort, offset *int, limit *int) (<-chan *models.Job, error) {
	panic(fmt.Errorf("not implemented"))
}
