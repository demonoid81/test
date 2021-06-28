package persons

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/internal/models"
	"github.com/sphera-erp/sphera/internal/organizations"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func (r *Resolver) RemoveContact(ctx context.Context, person *models.Person, contact *models.Contact) (bool, error) {
	if person == nil || person.UUID == nil {
		return false, gqlerror.Errorf("Error person  is nil")
	}

	if contact == nil || contact.UUID == nil {
		return false, gqlerror.Errorf("Error contact is nil")
	}

	tx, err := r.env.Cockroach.BeginX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "RemoveContact").Err(err).Msg("Error run transaction")
		return false, gqlerror.Errorf("Error run transaction")
	}
	defer tx.Rollback(ctx)

	var contacts []uuid.UUID
	if err := pglxqb.Select("uuid_contacts").
		From("persons").
		Where(pglxqb.Eq{"uuid": person.UUID}).
		RunWith(tx).QueryRow(ctx).Scan(&contacts); err != nil {
		r.env.Logger.Error().Str("module", "persons").Str("func", "RemoveContact").Err(err).Msg("Error get person contacts")
		return false, gqlerror.Errorf("Error get person contacts")
	}
	var newContacts []uuid.UUID
	for _, vContact := range contacts {
		if contact.UUID.String() != vContact.String() {
			newContacts = append(newContacts, vContact)
		}
	}

	_, err = pglxqb.Update("persons").
		Set("uuid_contacts", newContacts).
		Where("uuid = ?", person.UUID).
		RunWith(tx).Exec(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "persons").Str("func", "RemoveContact").Err(err).Msg("Error update persons")
		return false, gqlerror.Errorf("Error update organization")
	}
	err = tx.Commit(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "persons").Str("func", "RemoveContact").Err(err).Msg("Error run transaction")
		return false, gqlerror.Errorf("Error run transaction")
	}

	// получим еще раз пользователя
	var organization models.Organization
	oRows, err := pglxqb.Select("organizations.*").
		From("organizations").
		LeftJoin("users u on organizations.uuid = u.uuid_organization").
		LeftJoin("persons p on p.uuid_user = u.uuid").
		Where(pglxqb.Eq{"p.uuid": person.UUID}).
		RunWith(r.env.Cockroach).QueryX(ctx)

	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "UserMutation").Err(err).Msg("Error Select Org from user ")
		return false, gqlerror.Errorf("Error Select Org from user")
	}

	for oRows.Next() {
		if err := oRows.StructScan(&organization); err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "UserMutation").Err(err).Msg("Error parse org ")
			return false, gqlerror.Errorf("Error parse org")
		}
	}

	for _, c := range organizations.SubscriptionsMutateOrganizationResults.MutateOrganizationResults[uuid.Nil] {
		if err := organization.ParseRequestedFields(ctx, graphql.CollectFieldsCtx(c.SubContext, nil), r.env, r.env.Cockroach); err != nil {
			r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBookMutation").Err(err).Msg("Error parse row in medicalBook")
		}
		c.Chanel <- &organization
	}

	return true, nil
}
