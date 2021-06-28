package persons

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/app"
	"github.com/sphera-erp/sphera/internal/middleware"
	"github.com/sphera-erp/sphera/pkg/nalogSoap/bindPartnerStatus"
	"github.com/sphera-erp/sphera/pkg/nalogSoap/bindPartnerWithInn"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func (r *Resolver) ReqToPartner(ctx context.Context) (bool, error) {
	userUUID, err := middleware.ExtractUserInTokenMetadata(ctx, r.env)
	if err != nil {
		r.env.Logger.Error().Str("module", "jobs").Str("func", "jobs").Err(err).Msg("Error get user uuid from context")
		return false, gqlerror.Errorf("Error get user uuid from context")
	}
	// достанем персону из пользователя
	var personUUID uuid.UUID
	var inn *string
	err = pglxqb.Select("uuid", "inn").From("persons").Where(pglxqb.Eq{"uuid_user": userUUID}).RunWith(r.env.Cockroach).QueryRow(ctx).Scan(&personUUID, &inn)
	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error Select person from user")
		return false, gqlerror.Errorf("Error Select person from user")
	}

	if inn == nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error person not have inn")
		return false, gqlerror.Errorf("Error person not have inn")
	}

	reqId, err := bindPartnerWithInn.BindPartnerWithInn(r.env, *inn)
	if err != nil {
		fmt.Println(err)
		r.env.Logger.Error().Str("module", "flow").Str("func", "ReqToPartner").Err(err).Msg("Error in forming a request for a partnership")
		return false, gqlerror.Errorf(err.Error())
	}

	go getAnswerFromFNS(r.env, reqId, userUUID)

	return true, nil
}

func getAnswerFromFNS(app *app.App, reqId string, userUUID uuid.UUID) {
	var token *string
	err := pglxqb.Select("notification_token").
		From("users").
		Where(pglxqb.Eq{"uuid": userUUID}).RunWith(app.Cockroach).QueryRow(context.Background()).Scan(&token)
	if err != nil {
		app.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error select status from jobs")
		return
	}
	for {
		select {
		case <-time.After(10 * time.Minute):
			fmt.Println("Time out FNS Request")

			if token != nil {
				text := "В течении 30 мин вы не подтвердили партнерство с платформой"
				app.SendPush("192.168.10.244:9999", []string{*token}, text)
			}
			return
		case <-time.After(1 * time.Minute):
			fmt.Println("*************************************************** Validate **************************************************************")
			result, err := bindPartnerStatus.BindPartnerStatus(app, reqId)
			if err != nil {
				fmt.Println(err)
				if err.Error() != "Timeout" {
					if token != nil {
						text := "Произошла ошибка при запросе партнерства, " + err.Error()
						app.SendPush("192.168.10.244:9999", []string{*token}, text)
					}
					fmt.Println("Произошла ошибка при запросе партнерства, " + err.Error())
					return
				}
			}
			if result {
				if _, err = pglxqb.Update("persons").
					Set("validated", true).
					Where(pglxqb.Eq{"uuid_user": userUUID}).
					RunWith(app.Cockroach).Exec(context.Background()); err != nil {
					if token != nil {
						text := "Произошла ошибка при запросе партнерства, попробуйте поптыку позже"
						app.SendPush("192.168.10.244:9999", []string{*token}, text)
					}
					return
				}
				if token != nil {
					text := "Вы являетесь нашим партнером"
					app.SendPush("192.168.10.244:9999", []string{*token}, text)
				}
				return
			}
		}
	}
}
