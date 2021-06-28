package contacts

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/sphera-erp/sphera/app"
	"github.com/sphera-erp/sphera/internal/models"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type Resolver struct {
	env *app.App
	Resolvers
}

type Resolvers interface {
	ContactMutation(ctx context.Context, contact *models.Contact) (*models.Contact, error)
	Contact(ctx context.Context, contact models.Contact) (*models.Contact, error)
	Contacts(ctx context.Context, contact *models.Contact) ([]*models.Contact, error)
	ContactTypeMutation(ctx context.Context, contactType *models.ContactType) (*models.ContactType, error)
	ContactType(ctx context.Context, contactType *models.ContactType) (*models.ContactType, error)
	ContactTypes(ctx context.Context, contactType *models.ContactType) ([]*models.ContactType, error)
}

func NewContactsResolvers(app *app.App) (*Resolver, error) {
	return &Resolver{
		env: app,
	}, nil
}

func (r *Resolver) ContactMutation(ctx context.Context, contact *models.Contact) (*models.Contact, error) {
	tx, err := r.env.Cockroach.BeginX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "contacts").Str("func", "ContactMutation").Err(err).Msg("Error run transaction")
		return nil, gqlerror.Errorf("Error run transaction")
	}
	defer tx.Rollback(ctx)
	columns := make(map[string]interface{})
	rows, _, err := contact.Mutation(ctx, tx, r.env, nil, columns)
	if err != nil {
		r.env.Logger.Error().Str("module", "contacts").Str("func", "ContactMutation").Err(err).Msg("Error mutation contact")
		return nil, err
	}
	contact, err = contact.ParseRow(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, tx)
	if err != nil {
		r.env.Logger.Error().Str("module", "contacts").Str("func", "ContactMutation").Err(err).Msg("Error commit transaction")
		return nil, gqlerror.Errorf("Error commit transaction")
	}
	err = tx.Commit(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "contacts").Str("func", "ContactMutation").Err(err).Msg("Error commit transaction")
		return nil, gqlerror.Errorf("Error commit transaction")
	}
	return contact, nil
}

func (r *Resolver) Contact(ctx context.Context, contact models.Contact) (*models.Contact, error) {
	var err error
	sql := pglxqb.Select("contacts.*").From("contacts")
	result, sql, err := models.SqlGenSelectKeys(contact, sql, "contact", 1)
	if err != nil {
		r.env.Logger.Error().Str("module", "contacts").Str("func", "Contact").Err(err).Msg("Error generate select relations")
		return nil, gqlerror.Errorf("Error generate select relations")
	}
	rows, err := sql.Where(pglxqb.Eq(result)).RunWith(r.env.Cockroach).QueryX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "contacts").Str("func", "Contact").Err(err).Msg("Error select contact")
		return nil, gqlerror.Errorf("Error select users")
	}
	return contact.ParseRow(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, r.env.Cockroach)
}

func (r *Resolver) Contacts(ctx context.Context, contact *models.Contact) ([]*models.Contact, error) {
	var err error
	sql := pglxqb.Select("contacts.*").From("contacts")
	result, sql, err := models.SqlGenSelectKeys(contact, sql, "contact", 1)
	if err != nil {
		r.env.Logger.Error().Str("module", "contacts").Str("func", "Contacts").Err(err).Msg("Error generate select relations")
		return nil, gqlerror.Errorf("Error generate select relations")
	}
	rows, err := sql.Where(pglxqb.Eq(result)).RunWith(r.env.Cockroach).QueryX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "contacts").Str("func", "Contacts").Err(err).Msg("Error select contacts")
		return nil, gqlerror.Errorf("Error select users")
	}
	return contact.ParseRows(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, r.env.Cockroach)
}

func (r *Resolver) ContactTypeMutation(ctx context.Context, contactType *models.ContactType) (*models.ContactType, error) {
	tx, err := r.env.Cockroach.BeginX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "contacts").Str("func", "ContactTypeMutation").Err(err).Msg("Error run transaction")
		return nil, gqlerror.Errorf("Error run transaction")
	}
	defer tx.Rollback(ctx)
	columns := make(map[string]interface{})
	rows, _, err := contactType.Mutation(ctx, tx, r.env, nil, columns)
	if err != nil {
		r.env.Logger.Error().Str("module", "contacts").Str("func", "ContactTypeMutation").Err(err).Msg("Error mutation contactType")
		return nil, err
	}
	contactType, err = contactType.ParseRow(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, tx)
	if err != nil {
		r.env.Logger.Error().Str("module", "contacts").Str("func", "ContactTypeMutation").Err(err).Msg("Error parse rows")
		return nil, gqlerror.Errorf("Error commit transaction")
	}
	err = tx.Commit(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "contacts").Str("func", "ContactTypeMutation").Err(err).Msg("Error commit transaction")
		return nil, gqlerror.Errorf("Error commit transaction")
	}
	return contactType, nil
}

func (r *Resolver) ContactType(ctx context.Context, contactType *models.ContactType) (*models.ContactType, error) {
	var err error
	sql := pglxqb.Select("contact_type.*").From("contact_type")
	result, sql, err := models.SqlGenSelectKeys(contactType, sql, "contact_type", 1)
	if err != nil {
		r.env.Logger.Error().Str("module", "contacts").Str("func", "Users").Err(err).Msg("Error generate select relations")
		return nil, gqlerror.Errorf("Error generate select relations")
	}
	rows, err := sql.Where(pglxqb.Eq(result)).RunWith(r.env.Cockroach).QueryX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "contacts").Str("func", "Users").Err(err).Msg("Error select contactType")
		return nil, gqlerror.Errorf("Error select users")
	}
	return contactType.ParseRow(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, r.env.Cockroach)
}

func (r *Resolver) ContactTypes(ctx context.Context, contactType *models.ContactType) ([]*models.ContactType, error) {
	var err error
	sql := pglxqb.Select("contact_type.*").From("contact_type")
	result, sql, err := models.SqlGenSelectKeys(contactType, sql, "contact_type", 1)
	if err != nil {
		r.env.Logger.Error().Str("module", "contacts").Str("func", "Users").Err(err).Msg("Error generate select relations")
		return nil, gqlerror.Errorf("Error generate select relations")
	}
	rows, err := sql.Where(pglxqb.Eq(result)).RunWith(r.env.Cockroach).QueryX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "contacts").Str("func", "Users").Err(err).Msg("Error select contactTypes")
		return nil, gqlerror.Errorf("Error select users")
	}
	return contactType.ParseRows(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, r.env.Cockroach)
}
