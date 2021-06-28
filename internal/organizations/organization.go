package organizations

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/app"
	"github.com/sphera-erp/sphera/internal/middleware"
	"github.com/sphera-erp/sphera/internal/models"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type Resolver struct {
	env *app.App
	Resolvers
}

type Resolvers interface {
	OrganizationPositionMutation(ctx context.Context, organizationPosition *models.OrganizationPosition) (*models.OrganizationPosition, error)
	OrganizationContactMutation(ctx context.Context, organizationContact *models.OrganizationContact) (*models.OrganizationContact, error)
	OrganizationMutation(ctx context.Context, organization *models.Organization) (*models.Organization, error)
	OrganizationPosition(ctx context.Context, organizationPosition *models.OrganizationPosition) (*models.OrganizationPosition, error)
	OrganizationPositions(ctx context.Context, organizationPosition *models.OrganizationPosition) ([]*models.OrganizationPosition, error)
	OrganizationContact(ctx context.Context, organizationContact *models.OrganizationContact) (*models.OrganizationContact, error)
	OrganizationContacts(ctx context.Context, organizationContact *models.OrganizationContact) ([]*models.OrganizationContact, error)
	Organization(ctx context.Context, organization *models.Organization) (*models.Organization, error)
	Organizations(ctx context.Context, organization *models.Organization) ([]*models.Organization, error)
	ExcludePerson(ctx context.Context, organization uuid.UUID, person uuid.UUID) (*bool, error)
	OrganizationSub(ctx context.Context) (<-chan *models.Organization, error)
	GetOrganizationRating(ctx context.Context, organization *models.Organization) (*float64, error)
	RemoveParent(ctx context.Context, organization *models.Organization) (bool, error)
	ExcludePersonInObject(ctx context.Context, organization uuid.UUID, person uuid.UUID) (bool, error)
}

func NewOrganizationsResolvers(app *app.App) (*Resolver, error) {
	return &Resolver{
		env: app,
	}, nil
}

func (r *Resolver) ExcludePerson(ctx context.Context, organization uuid.UUID, person uuid.UUID) (bool, error) {
	tx, err := r.env.Cockroach.BeginX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "organizations").Str("func", "OrganizationMutation").Err(err).Msg("Error run transaction")
		return false, gqlerror.Errorf("Error run transaction")
	}
	defer tx.Rollback(ctx)
	sql := pglxqb.Select("uuid_persons").From("organizations")
	var persons []uuid.UUID
	err = sql.Where(pglxqb.Eq{"uuid": organization}).RunWith(tx).QueryRow(ctx).Scan(&persons)
	if err != nil {
		r.env.Logger.Error().Str("module", "organizations").Str("func", "OrganizationMutation").Err(err).Msg("Error get organization")
		return false, gqlerror.Errorf("Error run transaction")
	}
	var newPerson []uuid.UUID
	for _, vPerson := range persons {
		if person != vPerson {
			newPerson = append(newPerson, vPerson)
		}
	}
	_, err = pglxqb.Update("organizations").
		Set("uuid_persons", newPerson).
		Where("uuid = ?", organization).
		RunWith(tx).Exec(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "organizations").Str("func", "OrganizationMutation").Err(err).Msg("Error update organization")
		return false, gqlerror.Errorf("Error update organization")
	}
	err = tx.Commit(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "organizations").Str("func", "OrganizationMutation").Err(err).Msg("Error run transaction")
		return false, gqlerror.Errorf("Error run transaction")
	}
	return true, nil
}

func (r *Resolver) OrganizationPositionMutation(ctx context.Context, organizationPosition *models.OrganizationPosition) (*models.OrganizationPosition, error) {
	tx, err := r.env.Cockroach.BeginX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "organizations").Str("func", "OrganizationMutation").Err(err).Msg("Error run transaction")
		return nil, gqlerror.Errorf("Error run transaction")
	}
	defer tx.Rollback(ctx)
	columns := make(map[string]interface{})
	rows, _, err := organizationPosition.Mutation(ctx, tx, r.env, nil, columns)
	if err != nil {
		r.env.Logger.Error().Str("module", "organizations").Str("func", "OrganizationMutation").Err(err).Msg("Error mutation Organization")
		return nil, err
	}
	organizationPosition, err = organizationPosition.ParseRow(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, tx)
	if err != nil {
		r.env.Logger.Error().Str("module", "organizations").Str("func", "OrganizationMutation").Err(err).Msg("Error commit transaction")
		return nil, gqlerror.Errorf("Error commit transaction")
	}
	err = tx.Commit(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "organizations").Str("func", "OrganizationMutation").Err(err).Msg("Error commit transaction")
		return nil, gqlerror.Errorf("Error commit transaction")
	}
	return organizationPosition, err
}

func (r *Resolver) OrganizationPosition(ctx context.Context, organizationPosition *models.OrganizationPosition) (*models.OrganizationPosition, error) {
	var err error
	sql := pglxqb.Select("organization_positions.*").From("organization_positions")
	result, sql, err := models.SqlGenSelectKeys(organizationPosition, sql, "organization_positions", 1)
	if err != nil {
		r.env.Logger.Error().Str("module", "organizations").Str("func", "Organization").Err(err).Msg("Error generate select relations")
		return nil, gqlerror.Errorf("Error generate select relations")
	}
	rows, err := sql.Where(pglxqb.Eq(result)).RunWith(r.env.Cockroach).QueryX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "organizations").Str("func", "Organization").Err(err).Msg("Error select organization")
		return nil, gqlerror.Errorf("Error select person")
	}
	return organizationPosition.ParseRow(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, r.env.Cockroach)
}

func (r *Resolver) OrganizationPositions(ctx context.Context, organizationPosition *models.OrganizationPosition) ([]*models.OrganizationPosition, error) {
	var err error
	sql := pglxqb.Select("organization_positions.*").From("organization_positions")
	result, sql, err := models.SqlGenSelectKeys(organizationPosition, sql, "organization_positions", 1)
	if err != nil {
		r.env.Logger.Error().Str("module", "organizations").Str("func", "Organizations").Err(err).Msg("Error generate select relations")
		return nil, gqlerror.Errorf("Error generate select relations")
	}
	rows, err := sql.Where(pglxqb.Eq(result)).RunWith(r.env.Cockroach).QueryX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "organizations").Str("func", "Organizations").Err(err).Msg("Error select organizations")
		return nil, gqlerror.Errorf("Error select users")
	}
	return organizationPosition.ParseRows(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, r.env.Cockroach)
}

func (r *Resolver) OrganizationContactMutation(ctx context.Context, organizationContact *models.OrganizationContact) (*models.OrganizationContact, error) {
	tx, err := r.env.Cockroach.BeginX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "organizations").Str("func", "OrganizationMutation").Err(err).Msg("Error run transaction")
		return nil, gqlerror.Errorf("Error run transaction")
	}
	defer tx.Rollback(ctx)
	columns := make(map[string]interface{})
	rows, _, err := organizationContact.Mutation(ctx, tx, r.env, nil, columns)
	if err != nil {
		r.env.Logger.Error().Str("module", "organizations").Str("func", "OrganizationMutation").Err(err).Msg("Error mutation Organization")
		return nil, err
	}
	organizationContact, err = organizationContact.ParseRow(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, tx)
	if err != nil {
		r.env.Logger.Error().Str("module", "organizations").Str("func", "OrganizationMutation").Err(err).Msg("Error commit transaction")
		return nil, gqlerror.Errorf("Error commit transaction")
	}
	err = tx.Commit(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "organizations").Str("func", "OrganizationMutation").Err(err).Msg("Error commit transaction")
		return nil, gqlerror.Errorf("Error commit transaction")
	}
	return organizationContact, err
}

func (r *Resolver) OrganizationContact(ctx context.Context, organizationContact *models.OrganizationContact) (*models.OrganizationContact, error) {
	var err error
	sql := pglxqb.Select("organization_contacts.*").From("organization_contacts")
	result, sql, err := models.SqlGenSelectKeys(organizationContact, sql, "organization_contact", 1)
	if err != nil {
		r.env.Logger.Error().Str("module", "organizations").Str("func", "Organization").Err(err).Msg("Error generate select relations")
		return nil, gqlerror.Errorf("Error generate select relations")
	}
	rows, err := sql.Where(pglxqb.Eq(result)).RunWith(r.env.Cockroach).QueryX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "organizations").Str("func", "Organization").Err(err).Msg("Error select organization")
		return nil, gqlerror.Errorf("Error select person")
	}
	return organizationContact.ParseRow(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, r.env.Cockroach)
}

func (r *Resolver) OrganizationContacts(ctx context.Context, organizationContact *models.OrganizationContact) ([]*models.OrganizationContact, error) {
	var err error
	sql := pglxqb.Select("organization_contacts.*").From("organization_contacts")
	result, sql, err := models.SqlGenSelectKeys(organizationContact, sql, "organization_contact", 1)
	if err != nil {
		r.env.Logger.Error().Str("module", "organizations").Str("func", "Organizations").Err(err).Msg("Error generate select relations")
		return nil, gqlerror.Errorf("Error generate select relations")
	}
	rows, err := sql.Where(pglxqb.Eq(result)).RunWith(r.env.Cockroach).QueryX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "organizations").Str("func", "Organizations").Err(err).Msg("Error select organizations")
		return nil, gqlerror.Errorf("Error select users")
	}
	return organizationContact.ParseRows(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, r.env.Cockroach)
}

func (r *Resolver) OrganizationMutation(ctx context.Context, organization *models.Organization) (*models.Organization, error) {
	tx, err := r.env.Cockroach.BeginX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "organizations").Str("func", "OrganizationMutation").Err(err).Msg("Error run transaction")
		return nil, gqlerror.Errorf("Error run transaction")
	}
	defer tx.Rollback(ctx)
	columns := make(map[string]interface{})
	rows, _, err := organization.Mutation(ctx, tx, r.env, nil, columns)
	if err != nil {
		r.env.Logger.Error().Str("module", "organizations").Str("func", "OrganizationMutation").Err(err).Msg("Error mutation Organization")
		return nil, err
	}
	organization, err = organization.ParseRow(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, tx)
	if err != nil {
		r.env.Logger.Error().Str("module", "organizations").Str("func", "OrganizationMutation").Err(err).Msg("Error commit transaction")
		return nil, gqlerror.Errorf("Error commit transaction")
	}
	err = tx.Commit(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "organizations").Str("func", "OrganizationMutation").Err(err).Msg("Error commit transaction")
		return nil, gqlerror.Errorf("Error commit transaction")
	}

	for _, c := range SubscriptionsMutateOrganizationResults.MutateOrganizationResults[uuid.Nil] {
		if err := organization.ParseRequestedFields(ctx, graphql.CollectFieldsCtx(c.SubContext, nil), r.env, r.env.Cockroach); err != nil {
			r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBookMutation").Err(err).Msg("Error parse row in medicalBook")
		}
		c.Chanel <- organization
	}
	return organization, err
}

func (r *Resolver) Organization(ctx context.Context, organization *models.Organization) (*models.Organization, error) {
	var err error
	sql := pglxqb.Select("organizations.*").From("organizations")
	result, sql, err := models.SqlGenSelectKeys(organization, sql, "organizations", 1)
	if err != nil {
		r.env.Logger.Error().Str("module", "organizations").Str("func", "Organization").Err(err).Msg("Error generate select relations")
		return nil, gqlerror.Errorf("Error generate select relations")
	}
	rows, err := sql.Where(pglxqb.Eq(result)).RunWith(r.env.Cockroach).QueryX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "organizations").Str("func", "Organization").Err(err).Msg("Error select organization")
		return nil, gqlerror.Errorf("Error select person")
	}
	return organization.ParseRow(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, r.env.Cockroach)
}

func (r *Resolver) Organizations(ctx context.Context, organization *models.Organization) ([]*models.Organization, error) {
	var err error
	sql := pglxqb.Select("organizations.*").From("organizations")
	result, sql, err := models.SqlGenSelectKeys(organization, sql, "organizations", 1)
	if err != nil {
		r.env.Logger.Error().Str("module", "organizations").Str("func", "Organizations").Err(err).Msg("Error generate select relations")
		return nil, gqlerror.Errorf("Error generate select relations")
	}
	sql = sql.Where(pglxqb.Eq(result))

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
			case "system":
				if organization == nil {
					sql = sql.Where(pglxqb.Eq{"organizations.uuid_parent_organization": nil})
				}
			case "organizationManager":
				if uuidOrganization != nil {
					sql = sql.Where(pglxqb.Eq{"organizations.uuid_parent_organization": uuidOrganization})
				}
			case "branchManager":
				if groups != nil {
					sql = sql.Where(pglxqb.Eq{"organizations.uuid_parent": groups})
				}
			case "objectManager":
				if objects != nil {
					sql = sql.Where(pglxqb.Eq{"organizations.uuid": objects})
				} else {
					sql = sql.Where(pglxqb.Eq{"organizations.uuid": nil})
				}
			}
		} else {
			if organization == nil {
				sql = sql.Where(pglxqb.Eq{"organizations.uuid_parent_organization": nil})
			}
		}
	} else {
		if organization == nil {
			sql = sql.Where(pglxqb.Eq{"organizations.uuid_parent_organization": nil})
		}
	}

	rows, err := sql.OrderBy("organizations.is_group DESC").OrderBy("organizations.name ASC").RunWith(r.env.Cockroach).QueryX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "organizations").Str("func", "Organizations").Err(err).Msg("Error select organizations")
		return nil, gqlerror.Errorf("Error select users")
	}

	organizations, err := organization.ParseRows(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, r.env.Cockroach)

	if role != nil && *role == "objectManager" {
		for _, o := range organizations {
			o.Parent = nil
		}
	}

	return organizations, err

}
