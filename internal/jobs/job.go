package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/app"
	"github.com/sphera-erp/sphera/internal/middleware"
	"github.com/sphera-erp/sphera/internal/models"
	"github.com/sphera-erp/sphera/internal/utils"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type Resolver struct {
	env *app.App
	Resolvers
}

type Resolvers interface {
	CourseMutate(ctx context.Context, course *models.Course) (*models.Course, error)
	LocalityJobCostMutate(ctx context.Context, localityJobCost *models.LocalityJobCost) (*models.LocalityJobCost, error)
	JobTypeMutate(ctx context.Context, jobType *models.JobType) (*models.JobType, error)
	JobTemplateMutate(ctx context.Context, jobTemplate *models.JobTemplate) (*models.JobTemplate, error)
	JobMutate(ctx context.Context, job *models.Job) (*models.Job, error)
	CandidateMutate(ctx context.Context, candidate *models.Candidate) (*models.Candidate, error)
	StatusMutate(ctx context.Context, status *models.Status) (*models.Status, error)
	TagMutate(ctx context.Context, tag *models.Tag) (*models.Tag, error)
	MassCreationJobs(ctx context.Context, jobTemplate models.JobTemplate, objects []models.Organization, dates []*time.Time) (bool, error)
	//
	Course(ctx context.Context, course *models.Course) (*models.Course, error)
	Courses(ctx context.Context, course *models.Course, offset *int, limit *int) ([]*models.Course, error)
	LocalityJobCost(ctx context.Context, localityJobCost *models.LocalityJobCost) (*models.LocalityJobCost, error)
	LocalityJobCosts(ctx context.Context, localityJobCost *models.LocalityJobCost, offset *int, limit *int) ([]*models.LocalityJobCost, error)
	JobType(ctx context.Context, jobType *models.JobType) (*models.JobType, error)
	JobTypes(ctx context.Context, jobType *models.JobType, offset *int, limit *int) ([]*models.JobType, error)
	JobTemplate(ctx context.Context, jobTemplate *models.JobTemplate) (*models.JobTemplate, error)
	JobTemplates(ctx context.Context, jobTemplate *models.JobTemplate, offset *int, limit *int) ([]*models.JobTemplate, error)
	Job(ctx context.Context, job *models.Job) (*models.Job, error)
	Jobs(ctx context.Context, job *models.Job, filter *models.JobFilter, sort []models.JobSort, offset *int, limit *int) ([]*models.Job, error)
	ActiveJobs(ctx context.Context, job *models.Job, filter *models.JobFilter, sort []models.JobSort, offset *int, limit *int) ([]*models.Job, error)
	Candidate(ctx context.Context, candidate *models.Candidate) (*models.Candidate, error)
	Candidates(ctx context.Context, candidates *models.Candidate, offset *int, limit *int) ([]*models.Candidate, error)
	Status(ctx context.Context, status *models.Status) (*models.Status, error)
	Statuses(ctx context.Context, status *models.Status, offset *int, limit *int) ([]*models.Status, error)
	Tag(ctx context.Context, tag *models.Tag) (*models.Tag, error)
	Tags(ctx context.Context, tag *models.Tag, offset *int, limit *int) ([]*models.Tag, error)
	SetJobRating(ctx context.Context, job uuid.UUID, rating float64, description *string) (bool, error)

	JobSub(ctx context.Context) (<-chan *models.Job, error)
}

func NewJobsResolvers(app *app.App) (*Resolver, error) {
	return &Resolver{
		env: app,
	}, nil
}

func (r *Resolver) JobMutate(ctx context.Context, job *models.Job) (*models.Job, error) {
	tx, err := r.env.Cockroach.BeginX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBookMutation").Err(err).Msg("Error run transaction")
		return nil, gqlerror.Errorf("Error run transaction")
	}
	defer tx.Rollback(ctx)
	columns := make(map[string]interface{})
	rows, _, err := job.Mutation(ctx, tx, r.env, nil, columns)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBookMutation").Err(err).Msg("Error mutation medicalBook")
		return nil, err
	}
	job, err = job.ParseRow(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, tx)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBookMutation").Err(err).Msg("Error parse row in medicalBook")
		return nil, gqlerror.Errorf("Error parse row in medicalBook")
	}
	err = tx.Commit(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBookMutation").Err(err).Msg("Error commit transaction")
		return nil, gqlerror.Errorf("Error commit transaction")
	}
	for _, c := range SubscriptionsMutateJobResults.MutateJobResults[uuid.Nil] {
		jobSub := job
		if err := jobSub.ParseRequestedFields(ctx, graphql.CollectFieldsCtx(c.SubContext, nil), r.env, r.env.Cockroach); err != nil {
			r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBookMutation").Err(err).Msg("Error parse row in medicalBook")
			// return nil, gqlerror.Errorf("Error parse row in medicalBook")
		}
		c.Chanel <- jobSub
	}
	return job, err
}

func (r *Resolver) Job(ctx context.Context, job *models.Job) (*models.Job, error) {
	var err error
	sql := pglxqb.Select("jobs.*").From("jobs")
	result, sql, err := models.SqlGenSelectKeys(job, sql, "jobs", 1)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBook").Err(err).Msg("Error generate select relations")
		return nil, gqlerror.Errorf("Error generate select relations")
	}
	rows, err := sql.Where(pglxqb.Eq(result)).RunWith(r.env.Cockroach).QueryX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBook").Err(err).Msg("Error select medicalBook")
		return nil, gqlerror.Errorf("Error select medicalBooks")
	}
	return job.ParseRow(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, r.env.Cockroach)
}

func (r *Resolver) Jobs(ctx context.Context, job *models.Job, filter *models.JobFilter, sort []models.JobSort, offset *int, limit *int) ([]*models.Job, error) {
	var err error
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
	userType, err := middleware.ExtractUserTypeInTokenMetadata(ctx, r.env)
	if err != nil {
		r.env.Logger.Error().Str("module", "jobs").Str("func", "jobs").Err(err).Msg("Error get type current user")
		return nil, gqlerror.Errorf("Error get type current user")
	}
	if userType == models.SystemUser.String() {

		userUUID, err := middleware.ExtractUserInTokenMetadata(ctx, r.env)
		if err != nil {
			return nil, gqlerror.Errorf("Error get user uuid from context")
		}

		var uuidOrganization *uuid.UUID
		var objects []*uuid.UUID
		var groups []*uuid.UUID
		var uuidRole *uuid.UUID
		err = pglxqb.Select("uuid_objects, uuid_role, uuid_organization, uuid_groups").From("users").
			Where(pglxqb.Eq{"uuid": userUUID}).
			RunWith(r.env.Cockroach).QueryRow(ctx).Scan(&objects, &uuidRole, &uuidOrganization, &groups)
		if err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error Select person from user ")
			return nil, gqlerror.Errorf("Error Select person from user")
		}

		sql = sql.InnerJoin("organizations organizations_filter on organizations_filter.uuid = jobs.uuid_object")

		var role *string
		if uuidRole != nil {
			if uuidRole != nil {
				err = pglxqb.Select("role_type").From("roles").
					Where(pglxqb.Eq{"uuid": uuidRole}).
					RunWith(r.env.Cockroach).QueryRow(ctx).Scan(&role)
				if err != nil {
					r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error Select person from user ")
					return nil, gqlerror.Errorf("Error Select person from user")
				}
			}
			if role != nil {
				switch *role {
				case "organizationManager":
					if uuidOrganization != nil {
						sql = sql.Where(pglxqb.Eq{"organizations_filter.uuid_parent_organization": uuidOrganization})
					}
				case "branchManager":
					if groups != nil {
						sql = sql.Where(pglxqb.Eq{"organizations_filter.uuid_parent": groups})
					}
				case "objectManager":
					if objects != nil {
						sql = sql.Where(pglxqb.Eq{"organizations_filter.uuid": objects})
					} else {
						sql = sql.Where(pglxqb.Eq{"organizations_filter.uuid": nil})
					}
				}
			} else {
				if objects != nil {
					sql = sql.Where(pglxqb.Eq{"organizations_filter.uuid": objects})
				} else {
					sql = sql.Where(pglxqb.Eq{"organizations_filter.uuid": nil})
				}
			}
		} else {
			if objects != nil {
				sql = sql.Where(pglxqb.Eq{"organizations_filter.uuid": objects})
			} else {
				sql = sql.Where(pglxqb.Eq{"organizations_filter.uuid": nil})
			}
		}
	}

	if sort != nil {
		for _, sortItem := range sort {
			sql = sql.OrderBy(fmt.Sprintf("%s.%s %s", table, sortItem.Field, sortItem.Order))
		}
	}
	if limit != nil {
		sql = sql.Limit(uint64(*limit))
	}
	if offset != nil {
		sql = sql.Offset(uint64(*offset))
	}

	if userType == models.SelfEmployed.String() {
		sql = sql.Where(pglxqb.NotEq{"status": models.JobStatusDraft})
	}
	rows, err := sql.RunWith(r.env.Cockroach).QueryX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBooks").Err(err).Msg("Error select medicalBooks")
		return nil, gqlerror.Errorf("Error select medicalBooks")
	}
	return job.ParseRows(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, r.env.Cockroach)
}
