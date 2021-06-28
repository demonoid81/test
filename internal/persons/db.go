package persons

import (
	"context"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func GetPersonaPhotoByUserUUID(ctx context.Context, UserUUID uuid.UUID, db pglxqb.BaseRunner, logger *zerolog.Logger) (bucket *string, objectUUID *uuid.UUID, err error) {
	// SELECT c.uuid, c.bucket from persons Left Join content c on persons.uuid_photo = c.uuid where persons.uuid = $
	err = pglxqb.
		Select("c.uuid", "c.bucket").
		From("persons").
		LeftJoin("content c on persons.uuid_photo = c.uuid").
		Where(pglxqb.Eq{"uuid_user": UserUUID}).
		RunWith(db).
		QueryRow(ctx).Scan(&objectUUID, &bucket)
	if err != nil {
		logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return
}

func GetPassportPhotoByUserUUID(ctx context.Context, UserUUID uuid.UUID, db pglxqb.BaseRunner, logger *zerolog.Logger) (bucket *string, objectUUID *uuid.UUID, err error) {
	// SELECT c.uuid, c.bucket from persons Left Join content c on persons.uuid_photo = c.uuid where persons.uuid = $
	err = pglxqb.
		Select("c.uuid", "c.bucket").
		From("persons").
		LeftJoin("passports p on persons.uuid_passport = p.uuid").
		LeftJoin("content c on p.uuid_scan = c.uuid").
		Where(pglxqb.Eq{"uuid_user": UserUUID}).
		RunWith(db).
		QueryRow(ctx).Scan(&objectUUID, &bucket)
	if err != nil {
		logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return
}
