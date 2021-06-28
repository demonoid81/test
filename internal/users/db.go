package users

import (
	"context"
	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/app"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func createEmptyUser(ctx context.Context, app *app.App, phone string) (uuid.UUID, error) {
	if phone == "" {
		return uuid.Nil, gqlerror.Errorf("Error empty phone")
	}

	tx, err := app.Cockroach.BeginX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "user").Str("func", "createEmptyUser").Err(err).Msg("Error run transaction")
		return uuid.Nil, gqlerror.Errorf("Error run transaction")
	}
	defer tx.Rollback(ctx)

	uuidUser := uuid.New()
	if _, err = pglxqb.Insert("users").
		Columns("uuid").
		Values(uuidUser).
		RunWith(tx).Exec(ctx); err != nil {
		app.Logger.Error().Str("module", "user").Str("func", "createEmptyUser").Err(err).Msg("Error insert empty user in DB")
		return uuid.Nil, gqlerror.Errorf("Error insert phone number in DB")
	}
	var uuidPhone uuid.UUID
	err = pglxqb.Insert("contacts").
		Columns("presentation").
		Values(phone).
		Suffix("RETURNING uuid").
		RunWith(tx).
		Scan(ctx, &uuidPhone)
	if err != nil {
		app.Logger.Error().Str("module", "user").Str("func", "createEmptyUser").Err(err).Msg("Error insert phone number in DB")
		return uuid.Nil, gqlerror.Errorf("Error insert phone number in DB")
	}
	var UUIDPerson uuid.UUID
	err = pglxqb.Insert("persons").
		Columns("uuid_actual_contact", "uuid_user").
		Values(uuidPhone, uuidUser).
		Suffix("RETURNING uuid").
		RunWith(tx).
		Scan(ctx, &UUIDPerson)
	if err != nil {
		app.Logger.Error().Str("module", "user").Str("func", "createEmptyUser").Err(err).Msg("Error insert phone number in DB")
		return uuid.Nil, gqlerror.Errorf("Error insert phone number in DB")
	}
	_, err = pglxqb.Update("users").
		Set("uuid_contact", uuidPhone).
		Set("uuid_person", UUIDPerson).
		Where(pglxqb.Eq{"uuid": uuidUser}).
		RunWith(tx).Exec(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "user").Str("func", "createEmptyUser").Err(err).Msg("Error create null user in DB")
		return uuid.Nil, gqlerror.Errorf("Error create null user in DB")
	}
	err = tx.Commit(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "user").Str("func", "createEmptyUser").Err(err).Msg("Error commit transaction")
		return uuid.Nil, gqlerror.Errorf("Error commit transaction")
	}
	return uuidUser, nil
}

func countIdenticalPhones(ctx context.Context, app *app.App, phone string) (int, error) {
	var count int
	err := pglxqb.Select("COUNT(*)").
		From("contacts").
		Where("presentation = ?", phone).
		RunWith(app.Cockroach).
		Scan(ctx, &count)
	if err != nil {
		app.Logger.Error().Str("module", "user").Str("func", "countIdenticalPhones").Err(err).Msg("Error count Identical phones in db")
		return 0, gqlerror.Errorf("Error count Identical phones in db")
	}
	return count, nil
}

func getUserByContact(ctx context.Context, app *app.App, contact string) (uuidUser uuid.UUID, userType string, err error) {
	err = pglxqb.Select("users.uuid, users.type").
		From("users").LeftJoin("contacts as c on c.uuid=users.uuid_contact").
		Where("c.presentation = ?", contact).
		RunWith(app.Cockroach).
		Scan(ctx, &uuidUser, &userType)
	if err != nil {
		app.Logger.Error().Str("module", "user").Str("func", "countIdenticalPhones").Err(err).Msg("Error get user by contact in db")
		return uuid.Nil, "", gqlerror.Errorf("Error get user by contact in db")
	}
	return
}
