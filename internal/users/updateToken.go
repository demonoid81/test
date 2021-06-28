package users

import (
	"context"
	"fmt"

	"github.com/sphera-erp/sphera/internal/middleware"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func (r *Resolver) UpdateToken(ctx context.Context, token string) (bool, error) {
	tx, err := r.env.Cockroach.BeginX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "users").Str("func", "UpdateToken").Err(err).Msg("Error run transaction")
		return false, gqlerror.Errorf("Error run transaction")
	}
	defer tx.Rollback(ctx)

	fmt.Println("********************************* Обновляю Токен ***************************************************")
	fmt.Println("*********************************", token, "***************************************************")

	UUIDUser, err := middleware.ExtractUserInTokenMetadata(ctx, r.env)
	if err != nil {
		r.env.Logger.Err(err).Msg("Error select users")
		return false, gqlerror.Errorf("Error select users")
	}

	fmt.Println(UUIDUser)
	// todo нужно найти сначало токен удалить его если он не на той машине а после поменять
	_, err = pglxqb.Update("users").
		Set("notification_token", token).
		Where(pglxqb.Eq{"uuid": UUIDUser}).
		RunWith(tx).Exec(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "user").Str("func", "UpdateToken").Err(err).Msg("Error update token in DB")
		return false, gqlerror.Errorf("Error Error update token in DB")
	}
	err = tx.Commit(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "users").Str("func", "UpdateToken").Err(err).Msg("Error commit transaction")
		return false, gqlerror.Errorf("Error commit transaction")
	}
	return true, nil
}
