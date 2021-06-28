package bindPartnerStatus

import (
	"errors"
	"fmt"
	"time"

	"github.com/sphera-erp/sphera/app"
	"github.com/sphera-erp/sphera/pkg/nalogSoap/authRequest"
	"github.com/sphera-erp/sphera/pkg/nalogSoap/models/smz"
	"github.com/sphera-erp/sphera/pkg/nalogSoap/pkg/soap"
	"github.com/sphera-erp/sphera/pkg/nalogSoap/smz"
)

func BindPartnerStatus(app *app.App, id string) (bool, error) {
	var err error
	app.Cfg.Api.FnsTempToken.Lock()
	expire := app.Cfg.Api.FnsTempToken.Expire
	token := app.Cfg.Api.FnsTempToken.Token
	app.Cfg.Api.FnsTempToken.Unlock()
	fmt.Println("************************************************************:    ", expire.Before(time.Now()))
	if expire.Before(time.Now()) {
		token, err = authRequest.AuthRequest(app)
		if err != nil {
			fmt.Println("Error update token")
			return false, err
		}
	}

	params := smz_models.GetBindPartnerStatusRequest{
		Id: id,
	}

	headers := map[string]string{
		"FNS-OpenApi-Token":     token,
		"FNS-OpenApi-UserToken": app.Cfg.Api.MasterToken,
	}

	client := soap.NewClient("https://himself-ktr-api.nalog.ru:8090/ais3/smz/SmzIntegrationService?wsdl", soap.WithHTTPHeaders(headers))
	service := smz.NewOpenApiAsyncSMZ(client)
	reply, err := service.SendMessage(&smz.SendMessageRequest{
		Message: smz.SendMessage{
			Message: params,
		},
	})
	if err != nil {
		fmt.Println(err)
		return false, err
	}

	MessageID := reply.MessageId

	for {
		select {
		case <-time.After(30 * time.Second):
			return false, errors.New("Timeout")
		case <-time.After(5 * time.Second):
			fmt.Println("GetMessage")
			reply, err := service.GetMessage(&smz.GetMessageRequest{
				MessageId: MessageID,
			})
			if err != nil {
				fmt.Println(err)
				return false, err
			}

			MessageID := reply

			fmt.Println(*MessageID.ProcessingStatus)

			if *MessageID.ProcessingStatus != "PROCESSING" {
				if MessageID.Message.GetBindPartnerStatusResponse.Result == "COMPLETED" {
					return true, nil
				}
				return false, errors.New("Timeout")
			}
		}
	}
}
