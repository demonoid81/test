package bindPartnerWithInn

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

func BindPartnerWithInn(app *app.App, inn string) (string, error) {

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
			return "", err
		}
	}

	params := smz_models.PostBindPartnerWithInnRequest{
		Inn:         inn,
		Permissions: []string{"INCOME_REGISTRATION", "PAYMENT_INFORMATION", "TAX_PAYMENT", "INCOME_LIST", "INCOME_SUMMARY"},
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
		fmt.Println("Error", err)
		return "", err
	}

	MessageID := reply.MessageId

	for {
		select {
		case <-time.After(30 * time.Second):
			return "", errors.New("Time out FNS Request")
		case <-time.After(5 * time.Second):
			fmt.Println("GetMessage")
			reply, err := service.GetMessage(&smz.GetMessageRequest{
				MessageId: MessageID,
			})
			if err != nil {
				fmt.Println(err)
				return "", err
			}

			MessageID := reply

			fmt.Println(*MessageID.ProcessingStatus)

			if *MessageID.ProcessingStatus != "PROCESSING" {
				if MessageID.Message.SmzPlatformError.Code == "TAXPAYER_ALREADY_BOUND" {
					return "", errors.New(MessageID.Message.SmzPlatformError.Message)
				}
				if MessageID.Message.SmzPlatformError.Code == "TAXPAYER_UNREGISTERED" {
					return "", errors.New(MessageID.Message.SmzPlatformError.Message)
				}
				if MessageID.Message.SmzPlatformError.Code == "PARTNER_DENY" {
					return "", errors.New(MessageID.Message.SmzPlatformError.Message)
				}
				fmt.Println(MessageID.Message.PostBindPartnerWithInnResponse.Id)
				return MessageID.Message.PostBindPartnerWithInnResponse.Id, nil
			}
		}
	}
}
