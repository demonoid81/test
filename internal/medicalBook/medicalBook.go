package medicalBook

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
	MedicalBookMutation(ctx context.Context, medicalBook *models.MedicalBook) (*models.MedicalBook, error)
	MedicalBook(ctx context.Context, medicalBook *models.MedicalBook) (*models.MedicalBook, error)
	MedicalBooks(ctx context.Context, medicalBook *models.MedicalBook) ([]*models.MedicalBook, error)
}

func NewMedicalBooksResolvers(app *app.App) (*Resolver, error) {
	return &Resolver{
		env: app,
	}, nil
}

func (r *Resolver) MedicalBookMutation(ctx context.Context, medicalBook *models.MedicalBook) (*models.MedicalBook, error) {
	tx, err := r.env.Cockroach.BeginX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBookMutation").Err(err).Msg("Error run transaction")
		return nil, gqlerror.Errorf("Error run transaction")
	}
	defer tx.Rollback(ctx)

	columns := make(map[string]interface{})
	rows, _, err := medicalBook.Mutation(ctx, tx, r.env, nil, columns)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBookMutation").Err(err).Msg("Error mutation medicalBook")
		return nil, err
	}
	medicalBook, err = medicalBook.ParseRow(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, tx)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBookMutation").Err(err).Msg("Error parse row in medicalBook")
		return nil, gqlerror.Errorf("Error parse row in medicalBook")
	}
	err = tx.Commit(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBookMutation").Err(err).Msg("Error commit transaction")
		return nil, gqlerror.Errorf("Error commit transaction")
	}
	return medicalBook, err
}

func (r *Resolver) MedicalBook(ctx context.Context, medicalBook models.MedicalBook) (*models.MedicalBook, error) {
	var err error
	sql := pglxqb.Select("medical_books.*").From("medical_books")
	result, sql, err := models.SqlGenSelectKeys(medicalBook, sql, "medical_books", 1)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBook").Err(err).Msg("Error generate select relations")
		return nil, gqlerror.Errorf("Error generate select relations")
	}
	rows, err := sql.Where(pglxqb.Eq(result)).RunWith(r.env.Cockroach).QueryX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBook").Err(err).Msg("Error select medicalBook")
		return nil, gqlerror.Errorf("Error select medicalBooks")
	}
	return medicalBook.ParseRow(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, r.env.Cockroach)
}

func (r *Resolver) MedicalBooks(ctx context.Context, medicalBook *models.MedicalBook) ([]*models.MedicalBook, error) {
	var err error
	sql := pglxqb.Select("medical_books.*").From("medical_books")
	result, sql, err := models.SqlGenSelectKeys(medicalBook, sql, "medical_books", 1)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBooks").Err(err).Msg("Error generate select relations")
		return nil, gqlerror.Errorf("Error generate select relations")
	}
	rows, err := sql.Where(pglxqb.Eq(result)).RunWith(r.env.Cockroach).QueryX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBooks").Err(err).Msg("Error select medicalBooks")
		return nil, gqlerror.Errorf("Error select medicalBooks")
	}
	return medicalBook.ParseRows(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, r.env.Cockroach)
}
